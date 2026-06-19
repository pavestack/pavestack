.PHONY: help all test lint fmt pave install-pave build-portal generate-catalog infra-fmt infra-validate clean

help:
	@echo "Pavestack monorepo targets"
	@echo ""
	@echo "  Build"
	@echo "    make all              Run tests, build CLI, build portal"
	@echo "    make pave             Build pave CLI to ./bin/pave"
	@echo "    make install-pave     Build and install pave to $$GOPATH/bin"
	@echo "    make build-portal     Build static developer portal"
	@echo "    make generate-catalog Regenerate catalog.json from service metadata"
	@echo ""
	@echo "  Quality"
	@echo "    make test             Run all unit tests (Go + TS)"
	@echo "    make lint             Run all linters"
	@echo "    make fmt              Check formatting"
	@echo "    make infra-fmt        Terraform fmt check (platform-infra)"
	@echo "    make infra-validate   Terraform validate (platform-infra, no backend)"
	@echo ""
	@echo "  Maintenance"
	@echo "    make clean            Remove build artifacts"

all: test pave build-portal

test:
	cd service-template-api && go test ./...
	cd pave && go test ./...
	cd tests && go test ./...
	cd pavestack-portal && npm run test

lint:
	cd service-template-api && go vet ./...
	cd pave && go vet ./...
	cd tests && go vet ./...
	cd pavestack-portal && npm ci --silent && npx tsc --noEmit

fmt: infra-fmt
	cd service-template-api && test -z "$$(gofmt -l .)" || (gofmt -d . && exit 1)
	cd pave && test -z "$$(gofmt -l .)" || (gofmt -d . && exit 1)
	cd tests && test -z "$$(gofmt -l .)" || (gofmt -d . && exit 1)

pave:
	cd pave && go build -ldflags="-X github.com/pavestack/pave/internal/cli.Version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o ../bin/pave ./cmd/pave

install-pave:
	cd pave && go install -ldflags="-X github.com/pavestack/pave/internal/cli.Version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev)" ./cmd/pave

generate-catalog:
	cd pavestack-portal && node scripts/generate-catalog.mjs

build-portal: generate-catalog
	cd pavestack-portal && npm ci --silent && npm run build

infra-fmt:
	cd platform-infra && terraform fmt -recursive -check

infra-validate:
	cd platform-infra/envs/dev && terraform init -backend=false && terraform validate
	cd platform-infra/envs/prod && terraform init -backend=false && terraform validate

clean:
	rm -rf bin/
	rm -rf pavestack-portal/out/
	rm -f pavestack-portal/public/catalog.json
