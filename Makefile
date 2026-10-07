BINARY=terraform-provider-superset
VERSION=1.0.0

.PHONY: all build test clean format install

all: build test

build:
	go build -o $(BINARY) .

test:
	go test -v ./...

format:
	go fmt ./...

clean:
	rm -f $(BINARY)
	rm -rf examples/.terraform examples/terraform.tfstate* examples/.terraform.lock.hcl

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/xtruong/superset/$(VERSION)/linux_amd64/
	cp $(BINARY) ~/.terraform.d/plugins/registry.terraform.io/xtruong/superset/$(VERSION)/linux_amd64/
