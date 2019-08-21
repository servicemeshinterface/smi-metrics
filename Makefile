ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

HAS_TILT := $(shell command -v tilt;)
HAS_GOX := $(shell command -v gox;)
HAS_GHR := $(shell command -v ghr;)
HAS_SEMANTICS := $(shell command -v semantics;)

# Check for already defined environment variables
TAG ?= $(shell git describe --exact-match --tags $(git log -n1 --pretty='%h'))
IMAGE ?= thomasr/smi-metrics

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
	@#Check for semantics
ifndef HAS_SEMANTICS
	@echo "Installing semantics"
	go get -u github.com/stevenmatthewt/semantics
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

.PHONY: release
release: release-bootstrap
	set +e \
    $(eval RELEASE_TAG=$(shell semantics --output-tag))
	@if [ $(RELEASE_TAG) ]; then \
	  echo "Tagging the latest commit image with tag: ${RELEASE_TAG}"; \
	  docker tag ${IMAGE}:${COMMIT_TAG} ${IMAGE}:${RELEASE_TAG}; \
	  echo "Pusing the docker image: ${IMAGE}:${RELEASE_TAG} "; \
	  docker push ${IMAGE}:${RELEASE_TAG}; \
	  echo "Generating binaries for Github Release"; \
	  env GO111MODULE=on make vendor; \
	  make binaries; \
	  cd dist/ && gzip *; \
	  echo "Pushing the Github Release"; \
	  ghr -t ${GITHUB_TOKEN} -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} --replace ${RELEASE_TAG} ./; \
	else \
	  echo "The commit message(s) did not indicate a major/minor/patch version."; \
	fi 

.PHONY: push
push: build
	docker push $(IMAGE):$(TAG)

.PHONY: build
build:
	docker build -t $(IMAGE):$(TAG) .
