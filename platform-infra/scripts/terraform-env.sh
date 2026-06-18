#!/usr/bin/env bash
set -euo pipefail

environment="${1:-dev}"
command="${2:-plan}"
env_dir="envs/${environment}"

if [[ ! -d "${env_dir}" ]]; then
  echo "Unknown environment: ${environment}" >&2
  exit 1
fi

case "${command}" in
  fmt)
    terraform fmt -recursive
    ;;
  init)
    terraform -chdir="${env_dir}" init -backend-config=backend.hcl
    ;;
  validate)
    terraform -chdir="${env_dir}" validate
    ;;
  plan)
    terraform -chdir="${env_dir}" plan -out=tfplan
    ;;
  apply)
    terraform -chdir="${env_dir}" apply tfplan
    ;;
  *)
    echo "Usage: $0 <dev|prod> <fmt|init|validate|plan|apply>" >&2
    exit 1
    ;;
esac

