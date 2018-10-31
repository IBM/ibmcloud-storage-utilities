IMAGE = ibmcloud-block-storage-provisioner
VERSION := latest

.PHONY: all
all: provisioner

.PHONY: provisioner
provisioner:
	docker build -t $(IMAGE):$(VERSION) -f Dockerfile .
