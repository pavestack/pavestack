# 4. Karpenter over Cluster Autoscaler, with a static system node group

Date: 2026-07-01

## Status

Accepted

## Context

The EKS cluster (`platform-infra/modules/eks`) needs a mechanism to scale
worker capacity to match pod demand. The two mainstream options are the
Kubernetes Cluster Autoscaler, which scales existing `aws_eks_node_group`
/ Auto Scaling Groups up and down, and Karpenter, which provisions
right-sized EC2 instances directly (bypassing ASGs) based on unschedulable
pod requirements.

Cluster Autoscaler scales pre-defined node groups of fixed instance shapes;
matching capacity to workload shape means maintaining several node groups
per instance family/size, and scale-up latency is bounded by ASG launch
semantics. Karpenter observes unschedulable pods directly and launches the
instance type that best fits them, consolidating underutilized nodes
afterward — better bin-packing and faster scale-up for the golden-path
services this platform exists to run, at the cost of a controller that
itself needs meaningful IAM (`modules/karpenter/main.tf` reproduces the
official Karpenter v1 controller policy for EC2 fleet, launch template, and
interruption-queue actions) and an SQS interruption queue.

Whichever autoscaler is chosen, *something* still has to run the
cluster-critical add-ons (CoreDNS, the EBS CSI driver, and Karpenter's own
controller pod) before Karpenter itself is up and able to schedule
anything — a chicken-and-egg problem if the entire node fleet is
Karpinter-managed from boot.

## Decision

Use Karpenter as the workload autoscaler, and keep a small static
`aws_eks_node_group` ("default", `platform-infra/modules/eks/main.tf`) sized
by `node_desired_size`/`node_min_size`/`node_max_size` per environment for
system capacity.

The static group runs the add-ons that must exist before Karpenter can
schedule anything: `aws_eks_addon` resources (`vpc-cni`, `kube-proxy`,
`coredns`, `eks-pod-identity-agent`, `aws-ebs-csi-driver`) all
`depends_on` the node group in `platform-infra/modules/eks/main.tf`, and its
nodes carry the label `pavestack.io/node-pool = default` so system pods
(and Karpenter's own controller, which must be running before it can
provision anything) have a stable place to land regardless of what
Karpenter itself is doing. `kube-system` and other controller namespaces are
also carved out of Kyverno's baseline policies for the same reason
(ADR 6) — system components on the static group aren't held to the same
admission rules as tenant workloads.

Karpenter's `NodePool` and `EC2NodeClass` CRDs are not created by Terraform
— per ADR 3, they ship as an Argo CD-reconciled kustomize base
(`platform-config/templates/karpenter/nodepool.yaml`), since they're CRD
instances that need the Karpenter controller's CRDs already installed and
change independently of the controller itself (adjusting instance type
constraints or consolidation policy doesn't need a Terraform apply).

## Consequences

- Tenant workloads get faster scale-up and tighter bin-packing than
  ASG-based scaling would give, without per-instance-family node groups to
  maintain.
- The cluster carries two scaling mechanisms rather than one: the static
  node group (fixed min/max, resized per environment in
  `envs/{dev,prod}/main.tf`) plus Karpenter's dynamic fleet. Anyone
  reasoning about total node capacity has to add both.
- Karpenter's controller needs its own IAM role (`karpenter-controller`),
  SQS interruption queue, and node IAM role — more moving parts than
  Cluster Autoscaler's simpler ASG-permission footprint.
- Changing Karpenter's provisioning constraints (instance types, capacity
  types, consolidation) is an Argo CD-reconciled `NodePool` edit, not a
  Terraform apply — consistent with ADR 3's ownership split, but means
  Karpenter's Terraform module and its `platform-config` counterpart must
  both be checked when reasoning about node behavior.

See also: ADR 3 (Terraform/Argo CD ownership boundary), ADR 6 (Kyverno
baseline policy exclusions for `kube-system`/controller namespaces).
