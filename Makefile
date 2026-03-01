BINARY=terraform-provider-ironwifi
VERSION?=0.1.0
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
OS_ARCH=$(GOOS)_$(GOARCH)
HOSTNAME=registry.terraform.io
NAMESPACE=ironwifi
NAME=ironwifi
INSTALL_DIR=~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)

default: build

build:
	go build -o $(BINARY)

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v -timeout 120m

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

clean:
	rm -f $(BINARY)

.PHONY: default build install test testacc fmt vet lint clean
