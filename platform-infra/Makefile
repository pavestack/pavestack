SHELL := /usr/bin/env bash

ENV ?= dev
ENV_DIR := envs/$(ENV)
PLAN_FILE := $(ENV_DIR)/tfplan

.PHONY: fmt init validate plan apply destroy clean

fmt:
	terraform fmt -recursive

init:
	terraform -chdir=$(ENV_DIR) init -backend-config=backend.hcl

validate:
	terraform -chdir=$(ENV_DIR) validate

plan:
	terraform -chdir=$(ENV_DIR) plan -out=tfplan

apply:
	terraform -chdir=$(ENV_DIR) apply tfplan

destroy:
	terraform -chdir=$(ENV_DIR) destroy

clean:
	find . -type d -name .terraform -prune -exec rm -rf {} +
	find . -type f \( -name '*.tfplan' -o -name 'tfplan' \) -delete

