FROM golang:1.15-buster

ARG BUILD_AND_RELEASE
ARG DO_TEST
ARG DO_VERIFY
ARG DO_VET
ARG TARGET

ENV DO_LOCAL_INSTALL="false"
ENV GO111MODULE="on"
ENV GOFLAGS="-mod=vendor"
ENV PROTOC_VERSION="v3.14.0"
ENV PROTOC_ZIP="protoc-3.14.0-linux-x86_64.zip"

# Set up some common elements for docker images that build from here
# 1) Create a `appuser` user.
# 2) Create a `passwd` and `group` entry for that user in `/opt/appuser`, so that
#    it can be copied into other docker images
# These are done in separate layers for clarity; this docker image is only used
# locally, so reducing the number of layers is not an interesting concern.
RUN apt-get update && \
  apt-get install --no-install-recommends -y bash ca-certificates curl gcc git jq libdevmapper-dev libgpgme-dev libssl-dev libbtrfs-dev openssl unzip zip
RUN mkdir -p /home/appuser && \
  useradd appuser && \
  chown -R appuser:appuser /home/appuser
RUN mkdir -p /opt/appuser && \
  touch /opt/appuser/group && \
  touch /opt/appuser/passwd && \
  echo "appuser:x:1000:" >> /opt/appuser/group && \
  echo "appuser:x:1000:1000::/home/appuser:" >> /opt/appuser/passwd
RUN curl -sSL https://github.com/protocolbuffers/protobuf/releases/download/${PROTOC_VERSION}/${PROTOC_ZIP} -o /tmp/${PROTOC_ZIP} && \
  unzip -o /tmp/${PROTOC_ZIP} -d /usr/local bin/protoc && \
  unzip -o /tmp/${PROTOC_ZIP} -d /usr/local 'include/*' && \
  rm -f /tmp/${PROTOC_ZIP}

WORKDIR /go/src/github.com/object88/tugboat

COPY . .

# Run the build.
SHELL ["/bin/bash", "-c"]
RUN ./build.sh ${TARGET//,/ }