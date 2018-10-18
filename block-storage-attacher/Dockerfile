FROM alpine:3.7

# Default values
ARG git_commit_id=unknown
ARG git_remote_url=unknown
ARG build_date=unknown
ARG jenkins_build_number=unknown

# Add Labels to image to show build details
LABEL git-commit-id=${git_commit_id}
LABEL git-remote-url=${git_remote_url}
LABEL build-date=${build_date}
LABEL jenkins-build-number=${jenkins_build_number}

RUN mkdir -p /home/armada-storage/
RUN mkdir -p /host
ADD images/systemutil /home/armada-storage
ADD images/iscsi-attach.sh /home/armada-storage
ADD images/iscsi-portworx-volume.conf /home/armada-storage
ADD images/ibmc-portworx.service /home/armada-storage
ADD images/run.sh /home/armada-storage
RUN chmod 775 /home/armada-storage/systemutil
RUN chmod 775 /home/armada-storage/run.sh
RUN chmod 775 /home/armada-storage/iscsi-attach.sh

COPY block-storage-attacher /home/armada-storage
RUN chmod +x /home/armada-storage/block-storage-attacher

CMD ./home/armada-storage/run.sh
#ENTRYPOINT ["/home/armada-storage/block-storage-attacher"]
