locals {
  bucket_name = "${var.name_prefix}-${var.environment}-tfstate-${data.aws_caller_identity.current.account_id}"

  tags = {
    Project     = "pavestack"
    Repository  = "platform-infra"
    Environment = var.environment
    ManagedBy   = "terraform"
    CostCenter  = var.cost_center
    Team        = var.team
  }
}

data "aws_caller_identity" "current" {}

resource "aws_kms_key" "state" {
  description             = "KMS key for Pavestack Terraform state"
  deletion_window_in_days = 30
  enable_key_rotation     = true
  tags                    = local.tags
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
}

resource "aws_kms_alias" "state" {
  name          = "alias/${var.name_prefix}-${var.environment}-tfstate"
  target_key_id = aws_kms_key.state.key_id
}

resource "aws_s3_bucket" "state" {
  # checkov:skip=CKV2_AWS_61:Lifecycle configuration is not required for remote state bucket
  # checkov:skip=CKV2_AWS_62:Event notifications are not required for remote state bucket
  bucket        = local.bucket_name
  force_destroy = var.force_destroy
  tags          = local.tags
}

resource "aws_s3_bucket_versioning" "state" {
  bucket = aws_s3_bucket.state.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "state" {
  bucket = aws_s3_bucket.state.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.state.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "state" {
  bucket                  = aws_s3_bucket.state.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

