ARG PROXY_IMAGE_URL_DOCKERHUB=blank
FROM "${PROXY_IMAGE_URL_DOCKERHUB}"/golang:1.18.6

#ARG GOPROXY=off

WORKDIR /go/src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher
ADD . /go/src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher
RUN set -ex; cd /go/src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/ && \
CC=$(which musl-gcc) go build -o /go/bin/systemutil --ldflags '-w -linkmode external -extldflags "-static"' utils/systemutil.go

RUN set -ex; cd /go/src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/ && CGO_ENABLED=0 go install -v github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher
#CMD ["/bin/bash"]
