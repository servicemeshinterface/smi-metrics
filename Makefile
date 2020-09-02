ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

HAS_TILT := $(shell command -v tilt;)
HAS_GHR := $(shell command -v ghr;)


IMAGE_NAME      ?= servicemeshinterface/smi-metrics

GIT_COMMIT      ?= $(shell git rev-parse --short HEAD)
ifdef RELEASE_VERSION
IMAGE           := ${IMAGE_NAME}:${RELEASE_VERSION}
VERSION         := $(shell echo ${RELEASE_VERSION} | cut -c2- )
else
IMAGE           := ${IMAGE_NAME}:git-${GIT_COMMIT}
endif

.PHONY: release-bootstrap
release-bootstrap:
	@echo "Installing Helm v2"
	set -x; curl -L https://git.io/get_helm.sh | bash
	@#Check for ghr
ifndef HAS_GHR
	@echo "Installing ghr"
	go get -u github.com/tcnksm/ghr
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
ifndef RELEASE_VERSION
	@echo "Missing RELEASE_VERSION, is this being run from CI?"
	@exit 1
endif
	cp -R chart/smi-metrics tmp/smi-metrics
	sed -i.bak 's/CHART_VERSION/${VERSION}/g' tmp/smi-metrics/Chart.yaml
	for fname in $$(grep -rnl '$*' tmp/smi-metrics); do \
		sed -i.bak 's/VERSION/${RELEASE_VERSION}/g' $$fname; \
	done
	helm package tmp/smi-metrics -d tmp --save=false

.PHONY: dev
dev: bootstrap
	tilt up

.PHONY: release
release: release-bootstrap build-chart
ifndef RELEASE_VERSION
	@echo "Missing RELEASE_VERSION, is this being run from CI?"
	@exit 1
endif
ifndef GITHUB_TOKEN
	@echo "Requires a GITHUB_TOKEN with edit permissions for releases."
	@exit 1
endif
	ghr -u servicemeshinterface \
		${RELEASE_VERSION} \
		tmp/smi-metrics-*

.PHONY: push
push: build
	docker push $(IMAGE)

.PHONY: build
build:
	docker build -t $(IMAGE) .
