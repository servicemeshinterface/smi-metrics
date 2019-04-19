ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

HAS_TILT := $(shell command -v tilt;)

.PHONY: bootstrap
bootstrap:
	@# Bootstrap the required binaries
ifndef HAS_TILT
	echo "Install tilt from https://docs.tilt.dev/install.html"
endif

.PHONY: dev
dev: bootstrap
	tilt up
