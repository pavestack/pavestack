# syntax=docker/dockerfile:1

FROM golang:1.23-alpine@sha256:a7ecaac5efda22510d8c903bdc6b19026543f1eac3317d47363680df22161bd8 AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot@sha256:d093aa3e30dbadd3efe1310db061a14da60299baff8450a17fe0ccc514a16639
WORKDIR /app

COPY --from=builder /out/server /app/server

USER nonroot:nonroot
EXPOSE 8080

HEALTHCHECK NONE
ENTRYPOINT ["/app/server"]
