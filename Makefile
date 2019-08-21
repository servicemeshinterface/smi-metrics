ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

HAS_TILT := $(shell command -v tilt;)
HAS_GOX := $(shell command -v gox;)
HAS_GHR := $(shell command -v ghr;)
HAS_HELM := $(shell command -v helm;)


IMAGE_NAME      ?= deislabs/smi-metrics

GIT_COMMIT      ?= $(shell git rev-parse --short HEAD)
ifdef CIRCLE_TAG
IMAGE           := ${IMAGE_NAME}:${CIRCLE_TAG}
VERSION         := $(shell echo ${CIRCLE_TAG} | cut -c2- )
else
IMAGE           := ${IMAGE_NAME}:git-${GIT_COMMIT}
endif

.PHONY: release-bootstrap
release-bootstrap:
	@#Check for gox
ifndef HAS_GOX
	@echo "Installing gox"
	go get -u github.com/mitchellh/gox
endif
	@#Check for ghr
ifndef HAS_GHR
	@echo "Installing ghr"
	go get -u github.com/tcnksm/ghr
endif
ifndef HAS_HELM
	set -x; curl -L https://git.io/get_helm.sh | bash
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

tmp:
	mkdir tmp

.PHONY: build-chart
build-chart: tmp release-bootstrap
ifndef CIRCLE_TAG
	@echo "Missing CIRCLE_TAG, is this being run from circleci?"
	@exit 1
endif
	cp -R chart tmp/smi-metrics
	sed -i.bak 's/CHART_VERSION/${VERSION}/g' tmp/smi-metrics/Chart.yaml
	helm package tmp/smi-metrics -d tmp --save=false
	rm -rf tmp/smi-metrics/

.PHONY: dev
dev: bootstrap
	tilt up

.PHONY: release
release: release-bootstrap build-chart
ifndef CIRCLE_TAG
	@echo "Missing CIRCLE_TAG, is this being run from circleci?"
	@exit 1
endif
ifndef GITHUB_TOKEN
	@echo "Requires a GITHUB_TOKEN with edit permissions for releases."
	@exit 1
endif
	ghr -u deislabs \
		${CIRCLE_TAG} \
		tmp/smi-metrics-*

.PHONY: push
push: build
	docker push $(IMAGE)

.PHONY: build
build:
	docker build -t $(IMAGE) .
