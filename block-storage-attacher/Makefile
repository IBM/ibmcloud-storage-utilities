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
OSS_FILES := Dockerfile go.mod

PROXY_IMAGE_URL_DOCKERHUB:="docker.io"

ifdef ARTIFACTORY_API_KEY
PROXY_IMAGE_URL_DOCKERHUB="wcp-alchemy-containers-team-dockerhub-docker-remote.artifactory.swg-devops.com"
endif

.PHONY: all
all: deps gosec fmt vet buildgo test buildimage

.PHONY: driver
driver: deps gosec buildimage

.PHONY: deps
deps:
	echo "Installing dependencies ..."
#	glide install --strip-vendor
	go get github.com/pierrre/gotestcover
	go get -v github.com/securego/gosec/cmd/gosec

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
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

.PHONY: test
test:
	$(GOPATH)/bin/gotestcover -v -race -coverprofile=cover.out ${GOPACKAGES}

.PHONY: coverage
coverage:
	go tool cover -html=cover.out -o=cover.html


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
	go run github.ibm.com/alchemy-containers/armada-opensource-lib/cmd/makeoss ${OSS_FILES}
	go mod tidy
	sed -i '/armada-opensource-lib/d' ./OPENSOURCE

.PHONY: runanalyzedeps
runanalyzedeps:
	@docker build --rm --build-arg ARTIFACTORY_API_KEY="${ARTIFACTORY_API_KEY}"  -t armada/analyze-deps -f Dockerfile.dependencycheck .
	docker run -v `pwd`/dependency-check:/results armada/analyze-deps

.PHONY: analyzedeps
analyzedeps:
	/tmp/dependency-check/bin/dependency-check.sh --enableExperimental --log /results/logfile --out /results --disableAssembly \
		--suppress /src/dependency-check/suppression-file.xml --format JSON --prettyPrint --failOnCVSS 0 --scan /src

.PHONY: showanalyzedeps
showanalyzedeps:
	grep "VULNERABILITY FOUND" dependency-check/logfile;
	cat dependency-check/dependency-check-report.json |jq '.dependencies[] | select(.vulnerabilities | length>0)';
