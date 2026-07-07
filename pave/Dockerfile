# syntax=docker/dockerfile:1
#
# Builds pave-api only (not the `pave` CLI, which is meant to run on a
# developer's own machine, not in a container). Mirrors
# service-template-api/Dockerfile's multi-stage/distroless/nonroot pattern
# for supply-chain parity - see AGENTS.md's action-pinning conventions for
# why base images are pinned by digest.
#
# This image intentionally ships without `git`/`gh` on PATH (distroless
# has no shell at all). pave-api degrades gracefully without them -
# GitOpsToolsAvailable() logs a warning at startup and every job's
# open_pr step reports "skipped" rather than crashing - so the image is
# safe to run as-is for PAVE_API_DRY_RUN=true (the default). A deployment
# that needs pave-api to actually open real pull requests needs git/gh
# available some other way (a sidecar, or a different base image) - that's
# a deployment-specific decision, not something to bake into this image.

FROM golang:1.23-alpine@sha256:a7ecaac5efda22510d8c903bdc6b19026543f1eac3317d47363680df22161bd8 AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/pave-api ./cmd/pave-api

FROM gcr.io/distroless/static-debian12:nonroot@sha256:d093aa3e30dbadd3efe1310db061a14da60299baff8450a17fe0ccc514a16639
WORKDIR /app

COPY --from=builder /out/pave-api /app/pave-api

USER nonroot:nonroot
EXPOSE 8787

HEALTHCHECK NONE
ENTRYPOINT ["/app/pave-api"]
