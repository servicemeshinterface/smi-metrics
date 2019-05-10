ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

HAS_TILT := $(shell command -v tilt;)

TAG := $(shell git describe --exact-match --tags $(git log -n1 --pretty='%h'))

IMAGE := thomasr/smi-metrics

.PHONY: bootstrap
bootstrap:
	@# Bootstrap the required binaries
ifndef HAS_TILT
	echo "Install tilt from https://docs.tilt.dev/install.html"
endif

.PHONY: dev
dev: bootstrap
	tilt up

.PHONY: push
push: build
	docker push $(IMAGE):$(TAG)

.PHONY: build
build:
	NETRC=$$(cat ~/.netrc) && \
		docker build --build-arg NETRC -t $(IMAGE):$(TAG) .
