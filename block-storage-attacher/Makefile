#/*******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Container Service, 5737-D43
# * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************/

IMAGE = ibmcloud-block-storage-attacher
GOPACKAGES=$(shell go list ./... | grep -v /vendor/ | grep -v /cmd | grep -v /tests )
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./tests/e2e/*")
VERSION := latest

.PHONY: all
all: deps fmt vet buildgo test buildimage

.PHONY: driver
driver: deps buildgo buildimage

.PHONY: deps
deps:
	echo "Installing dependencies ..."
	glide install --strip-vendor
	go get github.com/pierrre/gotestcover

.PHONY: fmt
fmt:
	gofmt -l ${GOFILES}
	@if [ -n "$$(gofmt -l ${GOFILES})" ]; then echo 'Above Files needs gofmt fixes. Please run gofmt -l -w on your code.' && exit 1; fi

.PHONY: vet
vet:
	go vet ${GOPACKAGES}

.PHONY: buildgo
buildgo:
	GOOS=linux GOARCH=amd64 go build

.PHONY: test
test:
	#$(GOPATH)/bin/gotestcover -v -race -coverprofile=cover.out ${GOPACKAGES}
	$(GOPATH)/bin/gotestcover -v -race -coverprofile=cover.out ${GOPACKAGES}

.PHONY: coverage
coverage:
	go tool cover -html=cover.out -o=cover.html


.PHONY: buildimage
buildimage: build-systemutil
	docker build -t $(IMAGE):$(VERSION) -f Dockerfile .
#	cd images/ ;\
#	docker build -t $(IMAGE):$(VERSION) -f Dockerfile .
#        --build-arg git_commit_id=${GIT_COMMIT_SHA} \
#        --build-arg git_remote_url=${GIT_REMOTE_URL} \
#        --build-arg build_date=${BUILD_DATE} \
#        --build-arg jenkins_build_id=${BUILD_ID} \
#        --build-arg jenkins_build_number=${BUILD_NUMBER} \
#				--build-arg this_build_id=https://travis.ibm.com/${TRAVIS_REPO_SLUG}/builds/${TRAVIS_BUILD_ID} \
#	 -t $(IMAGE):$(VERSION) -f ./images/Dockerfile .


.PHONY: build-systemutil
build-systemutil:
	#Build and copy executables
	docker build \
		-t common-lib-builder --pull -f Dockerfile.builder .
	docker run common-lib-builder /bin/true
	docker cp `docker ps -q -n=1`:/go/bin/systemutil ./systemutil
	chmod 755 ./systemutil
	mv systemutil images/

