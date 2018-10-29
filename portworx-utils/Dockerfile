FROM ubuntu:18.10

RUN apt-get update
RUN apt-get install -y python3.6 python3-pip git jq apt-transport-https curl wget gzip tar
RUN pip3 install SoftLayer pyyaml 
RUN pip3 install --upgrade "urllib3==1.22" 
RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update
RUN apt-get install -y kubectl

ENV PYTHONUNBUFFERED=1

COPY ./mkpvyaml.py /
COPY ./px_iks_utils.py /
COPY ./vls /
COPY ./dvol /
COPY ./docker-entrypoint.sh /

RUN chmod +x /docker-entrypoint.sh
RUN chmod +x /mkpvyaml.py
RUN chmod +x /vls
RUN chmod +x /dvol

ENTRYPOINT [ "/docker-entrypoint.sh" ]
