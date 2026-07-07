data "aws_caller_identity" "current" {}

locals {
  oidc_issuer_host = replace(var.oidc_issuer_url, "https://", "")

  bucket_name      = "${var.cluster_name}-velero-backups-${data.aws_caller_identity.current.account_id}"
  velero_name      = "${var.cluster_name}-velero"
  velero_namespace = "velero"
  velero_sa        = "velero"
}

# ---------------------------------------------------------------------------
# Velero backup storage: S3 bucket + dedicated KMS key
# ---------------------------------------------------------------------------

resource "aws_kms_key" "velero" {
  description             = "KMS key for ${var.cluster_name} Velero backup bucket encryption"
  deletion_window_in_days = 30
  enable_key_rotation     = true
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      }
    ]
  })

  tags = merge(var.tags, {
    Name = "${var.cluster_name}-velero-backups"
  })
}

resource "aws_kms_alias" "velero" {
  name          = "alias/${var.cluster_name}-velero-backups"
  target_key_id = aws_kms_key.velero.key_id
}

resource "aws_s3_bucket" "velero" {
  bucket = local.bucket_name

  tags = merge(var.tags, {
    Name = local.bucket_name
  })
}

resource "aws_s3_bucket_versioning" "velero" {
  bucket = aws_s3_bucket.velero.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "velero" {
  bucket = aws_s3_bucket.velero.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.velero.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "velero" {
  bucket                  = aws_s3_bucket.velero.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "velero" {
  bucket = aws_s3_bucket.velero.id

  rule {
    id     = "expire-backups"
    status = "Enabled"

    filter {}

    expiration {
      days = var.backup_retention_days
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

data "aws_iam_policy_document" "velero_bucket" {
  statement {
    sid    = "DenyInsecureTransport"
    effect = "Deny"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions   = ["s3:*"]
    resources = [aws_s3_bucket.velero.arn, "${aws_s3_bucket.velero.arn}/*"]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}

resource "aws_s3_bucket_policy" "velero" {
  bucket = aws_s3_bucket.velero.id
  policy = data.aws_iam_policy_document.velero_bucket.json
}

# ---------------------------------------------------------------------------
# IRSA role for the Velero service account
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "velero_assume_role" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [var.oidc_provider_arn]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer_host}:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer_host}:sub"
      values   = ["system:serviceaccount:${local.velero_namespace}:${local.velero_sa}"]
    }
  }
}

resource "aws_iam_role" "velero" {
  name               = local.velero_name
  assume_role_policy = data.aws_iam_policy_document.velero_assume_role.json

  tags = merge(var.tags, {
    Name = local.velero_name
  })
}

data "aws_iam_policy_document" "velero" {
  statement {
    sid = "VeleroEBSSnapshots"
    actions = [
      "ec2:DescribeVolumes",
      "ec2:DescribeSnapshots",
      "ec2:CreateTags",
      "ec2:CreateVolume",
      "ec2:CreateSnapshot",
      "ec2:DeleteSnapshot",
    ]
    resources = ["*"]
  }

  statement {
    sid = "VeleroS3ObjectAccess"
    actions = [
      "s3:GetObject",
      "s3:DeleteObject",
      "s3:PutObject",
      "s3:AbortMultipartUpload",
      "s3:ListMultipartUploadParts",
    ]
    resources = ["${aws_s3_bucket.velero.arn}/*"]
  }

  statement {
    sid       = "VeleroS3BucketList"
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.velero.arn]
  }

  statement {
    sid = "VeleroKMSBucketKeyUsage"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:GenerateDataKey",
      "kms:GenerateDataKeyWithoutPlaintext",
      "kms:DescribeKey",
    ]
    resources = [aws_kms_key.velero.arn]
  }
}

resource "aws_iam_policy" "velero" {
  name        = local.velero_name
  description = "Least-privilege EBS snapshot and S3/KMS backup-bucket permissions for Velero on ${var.cluster_name}."
  policy      = data.aws_iam_policy_document.velero.json

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "velero" {
  role       = aws_iam_role.velero.name
  policy_arn = aws_iam_policy.velero.arn
}

# ---------------------------------------------------------------------------
# Velero Helm release
# ---------------------------------------------------------------------------

resource "helm_release" "velero" {
  name             = "velero"
  repository       = "https://vmware-tanzu.github.io/helm-charts"
  chart            = "velero"
  version          = var.chart_version
  namespace        = local.velero_namespace
  create_namespace = true
  timeout          = 900
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      initContainers = [
        {
          name            = "velero-plugin-for-aws"
          image           = "velero/velero-plugin-for-aws:v1.11.1"
          imagePullPolicy = "IfNotPresent"
          volumeMounts = [
            {
              mountPath = "/target"
              name      = "plugins"
            }
          ]
        }
      ]
      serviceAccount = {
        server = {
          annotations = {
            "eks.amazonaws.com/role-arn" = aws_iam_role.velero.arn
          }
        }
      }
      credentials = {
        useSecret = false
      }
      configuration = {
        backupStorageLocation = [
          {
            provider = "aws"
            bucket   = aws_s3_bucket.velero.bucket
            prefix   = var.cluster_name
            config = {
              region   = var.region
              kmsKeyId = aws_kms_key.velero.key_id
            }
          }
        ]
        volumeSnapshotLocation = [
          {
            name     = "default"
            provider = "aws"
            config = {
              region = var.region
            }
          }
        ]
      }
      schedules = {
        daily-cluster = {
          schedule = var.backup_schedule
          template = {
            ttl                = "720h0m0s"
            snapshotVolumes    = true
            includedNamespaces = ["*"]
          }
        }
      }
      deployNodeAgent = var.enable_node_agent
    })
  ], var.values)

  depends_on = [
    aws_iam_role_policy_attachment.velero,
    aws_s3_bucket_policy.velero,
  ]
}
