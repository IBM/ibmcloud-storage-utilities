GOPRIVATE := github.ibm.com
export

IMAGE = ibmcloud-block-storage-attacher
GOPACKAGES=$(shell go list ./... | grep -v /vendor/ | grep -v /cmd | grep -v /tests )
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./tests/e2e/*")
VERSION := latest

GIT_COMMIT_SHA="$(shell git rev-parse HEAD 2>/dev/null)"
GIT_REMOTE_URL="$(shell git config --get remote.origin.url 2>/dev/null)"
BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")"
# Jenkins vars. Set to `unknown` if the variable is not yet defined
BUILD_NUMBER?=unknown
TRAVIS_BUILD_NUMBER ?= unknown
OSS_FILES := Dockerfile.fips go.mod

LINT_VERSION=1.55.2
GOSEC_VERSION="v2.19.0"
NANCY_VERSION := 1.0.42
PROXY_IMAGE_URL=docker-na.artifactory.swg-devops.com/wcp-alchemy-containers-team-icr-docker-remote
PROXY_IMAGE_URL_DOCKERHUB:="docker.io"

ifdef ARTIFACTORY_API_KEY
PROXY_IMAGE_URL_DOCKERHUB="docker-na.artifactory.swg-devops.com/wcp-alchemy-containers-team-dockerhub-docker-remote"
endif

.PHONY: all
all: deps gosec fmt vet buildgo test buildimage-fips

install-bt: ## Install 'bt' via git clone + installer script. Set BT_VERSION to install a custom version.
@V=$${BT_VERSION:-master} C=~/.cache/bt/repo && if ! git -C "$$C" status >/dev/null; then mkdir -p "$$C" && rm -rf "$$C" && git clone https://github.ibm.com/alchemy-containers/go-build-tools.git "$$C"; fi && pushd "$$C" && git fetch && { git reset --hard "$$V" || git reset --hard origin/"$$V"; } && popd && "$$C"/scripts/install.sh "$$V"

.PHONY: driver
driver: deps gosec buildimage-fips

.PHONY: deps
deps: install-bt
	echo "Installing dependencies ..."
#	glide install --strip-vendor
	go get github.com/pierrre/gotestcover
	go install github.com/pierrre/gotestcover
	go get -v github.com/securego/gosec/cmd/gosec
	go install -v github.com/securego/gosec/cmd/gosec

.PHONY: fmt
fmt:
	gofmt -l ${GOFILES}
	@if [ -n "$$(gofmt -l ${GOFILES})" ]; then echo 'Above Files needs gofmt fixes. Please run gofmt -l -w on your code.' && exit 1; fi

.PHONY: vet
vet:
	go vet ${GOPACKAGES}
.PHONY: gosec
gosec:
	$(GOPATH)/bin/gosec -exclude-dir=tests ./...

.PHONY: buildgo
buildgo:
	GOOS=linux GOARCH=amd64 go build ${GOTAGS}

.PHONY: buildbsa
buildbsa:
	set -ex; go build ${GOTAGS} -v github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher

.PHONY: buildsu
buildsu:
	set -ex; go build ${GOTAGS} -o /go/bin/systemutil utils/systemutil.go
	chmod 755 /go/bin/systemutil

.PHONY: test
test:
	$(GOPATH)/bin/gotestcover -v -race -coverprofile=cover.out ${GOPACKAGES}

.PHONY: coverage
coverage:
	go tool cover -html=cover.out -o=cover.html

.PHONY: buildimage-fips
buildimage-fips: login-registry deps
	@ set -x; \
	ASSETURL="$$(curl -X GET -H "Authorization: token ${GHE_TOKEN}" "https://github.ibm.com/api/v3/repos/alchemy-registry/base/releases/latest" | jq -r '.assets[0].url')"; \
	curl -H "Authorization: token ${GHE_TOKEN}" -H "Accept: application/octet-stream" -L $$ASSETURL > base.image
	set -a
	. ./base.image && \
	IMAGE_TAG="$$(echo $$COMMON_ARMADA_golang_builder | cut -d':' -f2)" && \
	BUILDER="${PROXY_IMAGE_URL}/armada-master/ibm-storage-golang:$$IMAGE_TAG" && \
	bt build image -goproxy -- docker build --rm \
			--build-arg git_commit_id=${GIT_COMMIT_SHA} \
        	--build-arg git_remote_url=${GIT_REMOTE_URL} \
			--build-arg build_date=${BUILD_DATE} \
        	--build-arg travis_build_number=${TRAVIS_BUILD_NUMBER} \
			--build-arg REPO_SOURCE_URL=${REPO_SOURCE_URL} \
			--build-arg BUILD_URL="${BUILD_URL}" \
			--build-arg IMAGE_TAG=$$IMAGE_TAG \
			--build-arg BUILDER=$$COMMON_ARMADA_golang_builder \
			--build-arg GOOS="linux" \
			--build-arg GOARCH="amd64" \
			-t $(IMAGE):$(TRAVIS_COMMIT) -f Dockerfile.fips .
	set +a

.PHONY: buildimage
buildimage: build-systemutil
	#go run github.ibm.com/alchemy-containers/go-build-tools/cmd/goproxy -docker-build --
		docker build \
			--build-arg git_commit_id=${GIT_COMMIT_SHA} \
			--build-arg git_remote_url=${GIT_REMOTE_URL} \
			--build-arg build_date=${BUILD_DATE} \
			--build-arg travis_build_number=${TRAVIS_BUILD_NUMBER} \
			--build-arg REPO_SOURCE_URL=${REPO_SOURCE_URL} \
			--build-arg BUILD_URL=${BUILD_URL} \
			-t $(IMAGE):$(VERSION) -f Dockerfile .

.PHONY: build-systemutil
build-systemutil:
	#Build executables like block-attacher and system util GO binaries
	#go mod download
	#go run github.ibm.com/alchemy-containers/go-build-tools/cmd/goproxy -docker-build -- \
	docker build \
	    --build-arg PROXY_IMAGE_URL_DOCKERHUB=${PROXY_IMAGE_URL_DOCKERHUB} \
	-t common-lib-builder --pull -f Dockerfile.builder .
	docker run common-lib-builder /bin/true
	docker cp `docker ps -q -n=1`:/go/bin/systemutil ./systemutil
	docker cp `docker ps -q -n=1`:/go/bin/block-storage-attacher ./block-storage-attacher
	chmod 755 ./systemutil
	mv systemutil images/

.PHONY: oss
oss:
	bt lint opensource -generate ${OSS_FILES}
	go mod tidy
	grep -v "armada-opensource-lib" ./OPENSOURCE > temp && mv temp ./OPENSOURCE

.PHONY: lint
 lint: install-bt
	bt lint golangci --version="v${LINT_VERSION}" -- --timeout 5m
	bt lint opensource ${OSS_FILES}
	bt lint semver-deps -prefix github.ibm.com/alchemy-containers
	bt lint nancy --version="v${NANCY_VERSION}"
	bt lint gosec -config .gosec.json -version ${GOSEC_VERSION}

.PHONY: lint-root-repo
 lint-root-repo:
	bt lint copyright
	bt lint shellcheck

.PHONY: lint-fix
 lint-fix: install-bt
	bt lint golangci --version="v${LINT_VERSION}" -- --fix
	bt lint opensource -generate ${OSS_FILES}
	bt lint copyright -generate

.PHONY: login-registry
login-registry: install-bt
	bt build login-registry -artifactory-username "${ARTIFACTORY_USER}" -all-artifactory -iam-key "${registry_iam_api_key}" -all-icr