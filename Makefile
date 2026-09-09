VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/sameeralam3127/kuberescue/internal/cli.version=$(VERSION)

.PHONY: build test lint fmt vet tidy vulncheck vulncheck-test docker clean

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/kuberescue ./cmd/kuberescue

test:
	go test -race -cover ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

lint: vet
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed on:" && gofmt -l . && exit 1)

tidy:
	go mod tidy

vulncheck:
	./scripts/vulncheck.sh --mode full

vulncheck-test:
	./scripts/vulncheck_test.sh

docker:
	docker build --build-arg VERSION=$(VERSION) -t kuberescue:local .

clean:
	rm -rf bin/
