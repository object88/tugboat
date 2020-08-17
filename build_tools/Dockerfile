FROM golang:1.15-buster

ARG BUILD_AND_RELEASE
ARG DO_TEST
ARG DO_VERIFY
ARG DO_VET
ARG TARGET

ENV DO_LOCAL_INSTALL="false"
ENV GO111MODULE="on"
ENV GOFLAGS="-mod=vendor"

# Set up some common elements for docker images that build from here
# 1) Create a `appuser` user.
# 2) Create a `passwd` and `group` entry for that user in `/opt/appuser`, so that
#    it can be copied into other docker images
# These are done in separate layers for clarity; this docker image is only used
# locally, so reducing the number of layers is not an interesting concern.
RUN apt-get update && \
  apt-get install --no-install-recommends -y bash ca-certificates curl gcc git jq libdevmapper-dev libgpgme-dev libssl-dev libbtrfs-dev openssl zip
RUN mkdir -p /home/appuser && \
  useradd appuser && \
  chown -R appuser:appuser /home/appuser
RUN mkdir -p /opt/appuser && \
  touch /opt/appuser/group && \
  touch /opt/appuser/passwd && \
  touch /opt/appuser/subgid && \
  touch /opt/appuser/subuid && \
  echo "appuser:x:1000:" >> /opt/appuser/group && \
  echo "appuser:x:1000:1000::/home/appuser:" >> /opt/appuser/passwd && \
  echo "appuser:1000:1" >> /opt/appuser/subgid && \
  echo "appuser:1000:1" >> /opt/appuser/subuid

WORKDIR /go/src/github.com/object88/tugboat

COPY . .

# Run the build.
SHELL ["/bin/bash", "-c"]
RUN ./build.sh --no-local-install ${TARGET//,/ }