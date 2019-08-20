ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

HAS_TILT := $(shell command -v tilt;)

ifeq ($(TAG),)
TAG := $(shell git describe --exact-match --tags $(git log -n1 --pretty='%h'))
endif

ifeq ($(IMAGE),)
IMAGE := thomasr/smi-metrics
endif

.PHONY: bootstrap
bootstrap:
	@# Bootstrap the required binaries
ifndef HAS_TILT
	echo "Install tilt from https://docs.tilt.dev/install.html"
endif

.PHONY: lint
lint:
	./bin/lint --verbose

.PHONY: test
test:
	go test -cover -v -race ./...

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: dep
dep:
	go mod download

.PHONY: binaries
binaries:
	gox -os="linux darwin windows" -arch="amd64" -output="dist/smi_metrics_{{.OS}}_{{.Arch}}" ./cmd/smi-metrics

.PHONY: dev
dev: bootstrap
	tilt up

.PHONY: push
push: build
	docker push $(IMAGE):$(TAG)

.PHONY: build
build:
	docker build -t $(IMAGE):$(TAG) .
