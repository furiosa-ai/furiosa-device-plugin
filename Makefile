SHELL := /bin/bash

# make assumption that hwloc is installed with brew command "brew install hwloc"
ifeq ($(shell uname -s),Darwin)
    CGO_CFLAGS := -I/opt/homebrew/opt/hwloc/include
    CGO_LDFLAGS := -L/opt/homebrew/opt/hwloc/lib
endif

# make assumption that golang is installed on the underlying machine.
define install_deps_function
    @UNAME_S=$$(uname -s); \
    if [ "$$UNAME_S" = "Linux" ]; then \
        echo "Installing for Ubuntu/Debian familly"; \
        sudo apt-get install hwloc libhwloc-dev ginkgo; \
    elif [ "$$UNAME_S" = "Darwin" ]; then \
        echo "macOS detected. Installing using Homebrew..."; \
        brew install hwloc; \
        go env -w GO111MODULE=on; \
        go install github.com/onsi/ginkgo/v2/ginkgo@latest; \
        export PATH=$$PATH:$$(go env GOPATH)/bin; \
        echo $$PATH; \
    else \
        echo "Unsupported Operating System"; \
        exit 1; \
    fi
endef

# regexp to filter some directories from testing
EXCLUDE_DIR_REGEXP := E2E

.PHONY: all
all: build fmt lint vet test tidy vendor

.PHONY: build
build:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go build cmd/main.go

.PHONY: fmt
fmt:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go fmt ./...

.PHONY: lint
lint:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) golangci-lint run --timeout=30m --no-config --verbose

.PHONY: vet
vet:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go vet -v ./...

.PHONY: test
test:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go test -skip $(EXCLUDE_DIR_REGEXP) ./...

.PHONY: cover
cover:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go test -skip $(EXCLUDE_DIR_REGEXP) -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out

.PHONY: tidy
tidy:
	go mod tidy

.PHONY:vendor
vendor:
	go mod vendor

.PHONY: install-deps
install-deps:
	$(call install_deps_function)

.PHONY: image
image:
	docker build . -t ghcr.io/furiosa-ai/furiosa-device-plugin:devel --progress=plain --platform=linux/amd64

.PHONY: image-no-cache
image-no-cache:
	docker build . --no-cache -t ghcr.io/furiosa-ai/furiosa-device-plugin:devel --progress=plain --platform=linux/amd64

.PHONY: helm-lint
helm-lint:
	helm lint ./deployments/helm

.PHONY: e2e-inference-pod-image
e2e-inference-pod-image:
	docker build --build-arg FURIOSA_IAM_KEY=$(FURIOSA_IAM_KEY) --build-arg FURIOSA_IAM_PWD=$(FURIOSA_IAM_PWD) . -t ghcr.io/furiosa-ai/furiosa-device-plugin/e2e/inference:latest --no-cache --progress=plain --platform=linux/amd64 -f ./e2e/inference_pod/Dockerfile

.PHONY: e2e-verification
e2e-verification:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go build e2e/verification_pod/verification.go

.PHONY: e2e-verification-image
e2e-verification-image:
	docker build . -t ghcr.io/furiosa-ai/furiosa-device-plugin/e2e/verification:latest --progress=plain --platform=linux/amd64 -f ./e2e/verification_pod/Dockerfile

.PHONY:e2e
e2e:
	# build container image
	# run e2e test framework
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) ginkgo ./e2e
