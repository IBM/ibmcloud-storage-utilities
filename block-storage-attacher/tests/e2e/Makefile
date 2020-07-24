# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

VERSION :=
TAG := $(shell git describe --abbrev=0 --tags HEAD 2>/dev/null)
COMMIT := $(shell git rev-parse HEAD)
GOPACKAGES=$(shell go list ./... | grep -v /vendor/)

IMAGE_NAME?=armada-master/armada-storage-e2e
.PHONY: all
all: verify-gofmt vet

test-integration: verify-gofmt
	go test `go list ./... | grep -v 'vendor\|e2e'`
.PHONY: test-integration

# Build PVG Testing files that will go in the side car container
.PHONY: pvg-build-e2e-tests-image
pvg-build-e2e-tests-image: verify-gofmt
	go test  ./e2e-tests -c -o pvg/armada-storage-e2e
	sudo docker build \
        -f Dockerfile \
        -t ${IMAGE_NAME}:latest .; \
        
armada-portworx-e2e-test: verify-gofmt
	@echo $(PVG_PHASE)
ifeq ("$(PVG_PHASE)","armada-prestage")
	go test ./e2e-tests -v -p 1 -ginkgo.progress -ginkgo.v -ginkgo.trace -ginkgo.failFast true -ginkgo.focus "PRESTAGE" --kubeconfig=${KUBECONFIGPATH} -timeout 90m 
else
	go test ./e2e-tests -v -p 1 -ginkgo.progress -ginkgo.v -ginkgo.trace -ginkgo.failFast true --kubeconfig=${KUBECONFIGPATH}  -timeout 90m
endif
.PHONY: armada-portworx-e2e-test

verify-gofmt:
	gofmt -s -w -l `find . -type f -name "*.go" | grep -v vendor` || exit 1
.PHONY: verify-gofmt

.PHONY: vet
vet:
	go vet ${GOPACKAGES}

deps:
	echo "Installing dependencies ..."
	glide install --strip-vendor
.PHONY: deps

.PHONY: clean 
clean:
	rm -f pvg/armada-storage-e2e
