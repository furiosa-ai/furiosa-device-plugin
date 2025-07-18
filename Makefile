SHELL := /bin/bash

CGO_CFLAGS := -I/usr/local/include
CGO_LDFLAGS := -L/usr/local/lib

# make assumption that golang is installed on the underlying machine.
define install_deps_function
    @UNAME_S=$$(uname -s); \
    if [ "$$UNAME_S" = "Linux" ]; then \
        echo "Installing for Ubuntu/Debian familly"; \
        sudo apt-get install ginkgo; \
    elif [ "$$UNAME_S" = "Darwin" ]; then \
        echo "macOS detected."; \
        go env -w GO111MODULE=on; \
        go install github.com/onsi/ginkgo/v2/ginkgo@latest; \
        export PATH=$$PATH:$$(go env GOPATH)/bin; \
        echo $$PATH; \
    else \
        echo "Unsupported Operating System"; \
        exit 1; \
    fi
endef

# Detect the OS and set the appropriate library path variable
ifeq ($(shell uname), Linux)
    LIBRARY_PATH_VAR := LD_LIBRARY_PATH
else ifeq ($(shell uname), Darwin)
    LIBRARY_PATH_VAR := DYLD_LIBRARY_PATH
else
    $(error Unsupported OS)
endif

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
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) golangci-lint run --timeout=30m

.PHONY: vet
vet:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go vet -v ./...

.PHONY: test
test:
	$(LIBRARY_PATH_VAR)=/usr/local/lib CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go test ./...

.PHONY: cover
cover:
	$(LIBRARY_PATH_VAR)=/usr/local/lib CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go test -coverprofile=coverage.out ./...
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
	docker build . -t registry.corp.furiosa.ai/furiosa/furiosa-device-plugin:devel --progress=plain --platform=linux/amd64

.PHONY: image-no-cache
image-no-cache:
	docker build . --no-cache -t registry.corp.furiosa.ai/furiosa/furiosa-device-plugin:devel --progress=plain --platform=linux/amd64

.PHONY: image-rel
image-rel:
	docker build . -t registry.corp.furiosa.ai/furiosa/furiosa-device-plugin:latest --progress=plain --platform=linux/amd64

.PHONY: image-no-cache-rel
image-no-cache-rel:
	docker build . --no-cache -t registry.corp.furiosa.ai/furiosa/furiosa-device-plugin:latest --progress=plain --platform=linux/amd64
