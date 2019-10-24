IMAGE = ibmcloud-block-storage-attacher
GOPACKAGES=$(shell go list ./... | grep -v /vendor/ | grep -v /cmd | grep -v /tests )
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./tests/e2e/*")
VERSION := latest

GIT_COMMIT_SHA="$(shell git rev-parse HEAD 2>/dev/null)"
GIT_REMOTE_URL="$(shell git config --get remote.origin.url 2>/dev/null)"
BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")"
# Jenkins vars. Set to `unknown` if the variable is not yet defined
BUILD_NUMBER?=unknown

.PHONY: all
all: deps gosec fmt vet buildgo test buildimage

.PHONY: driver
driver: deps gosec buildimage

.PHONY: deps
deps:
	echo "Installing dependencies ..."
	glide install --strip-vendor
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
	docker build	\
        --build-arg git_commit_id=${GIT_COMMIT_SHA} \
        --build-arg git_remote_url=${GIT_REMOTE_URL} \
        --build-arg build_date=${BUILD_DATE} \
        --build-arg jenkins_build_number=${BUILD_NUMBER} \
	-t $(IMAGE):$(VERSION) -f Dockerfile .


.PHONY: build-systemutil
build-systemutil:
	#Build executables like block-attacher and system util GO binaries
	docker build -t common-lib-builder --pull -f Dockerfile.builder .
	docker run common-lib-builder /bin/true
	docker cp `docker ps -q -n=1`:/go/bin/systemutil ./systemutil
	docker cp `docker ps -q -n=1`:/go/bin/block-storage-attacher ./block-storage-attacher
	chmod 755 ./systemutil
	mv systemutil images/

