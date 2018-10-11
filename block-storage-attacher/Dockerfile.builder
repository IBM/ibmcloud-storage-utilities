FROM golang:1.8.3

WORKDIR /go/src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher
ADD . /go/src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher
RUN set -ex; cd /go/src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/ && \
CC=$(which musl-gcc) go build -o /go/bin/systemutil --ldflags '-w -linkmode external -extldflags "-static"' utils/systemutil.go
CMD ["/bin/bash"]
