#!/usr/bin/env bash
set -e

cd $(dirname "$0")

TARGET=()

export UNIQUE_TAG=${UNIQUE_TAG:-$(git describe --tags)-$(git rev-parse --short HEAD)}
export PATH=$PWD/bin/:$PATH

# Set defaults, allow env val to override
BUILD_AND_RELEASE=${BUILD_AND_RELEASE:-"false"}
DO_PUSH=${DO_PUSH:-"false"}
DO_TEST=${DO_TEST:-"true"}
DO_VERIFY=${DO_VERIFY:-"true"}
DO_VET=${DO_VET:-"true"}

while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --fast)
      DO_TEST="false"
      DO_VERIFY="false"
      DO_VET="false"
      shift
      ;;
    --push)
      DO_PUSH="true"
      shift
      ;;
    --no-test)
      DO_TEST="false"
      shift
      ;;
    --no-verify)
      DO_VERIFY="false"
      shift
      ;;
    --no-vet)
      DO_VET="false"
      shift
      ;;
    *)
      TARGET+=("$1")
      shift
      ;;
  esac
done

# Let this run the `go mod verify` task, so that we don't have to in every
# docker build.
TARGETS=$(IFS=","; echo "${TARGET[*]}")
time docker build --build-arg DO_TEST=$DO_TEST --build-arg DO_VET=$DO_VET --build-arg DO_VERIFY=$DO_VERIFY --build-arg TARGET="$TARGETS" --build-arg BUILD_AND_RELEASE=$BUILD_AND_RELEASE --tag gobuild:local -f build_tools/Dockerfile .

DOCKER_IMAGES=()

echo "Targets length: ${#TARGET[@]}"

if [ ${#TARGET[@]} -eq 0 ]; then
  echo "No targets specified for docker images; searching directory."
  TARGET=($(ls apps))
fi

for D in ${TARGET[@]}; do
  if ! [ -d "./apps/$D" ]; then
    continue
  fi

  if [ ! -z $TARGET ] && [[ ! " ${TARGET[@]} " =~ " ${D} " ]]; then
    continue
  fi

  if [ -f "./apps/$D/docker-targets.ini" ]; then
    echo ""
    echo "Found apps/$D/docker-targets.ini..."

    IFS=$'\n'
    for LINE in $(cat < "./apps/$D/docker-targets.ini"); do
      DOCKER_FILENAME="$(echo -e "${LINE%=*}" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
      DOCKER_IMAGENAME="$(echo -e "${LINE#*=}" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"

      echo ""
      echo "Building targetted docker image for '$D' using '$DOCKER_FILENAME' to build '$DOCKER_IMAGENAME'..."
      time docker build -t "object88/$DOCKER_IMAGENAME:latest" -f apps/$D/$DOCKER_FILENAME .
      docker tag "object88/$DOCKER_IMAGENAME:latest" "object88/$DOCKER_IMAGENAME:$UNIQUE_TAG"
      DOCKER_IMAGES+=("$DOCKER_IMAGENAME:$UNIQUE_TAG")
    done
  elif [ -f "./apps/$D/Dockerfile" ]; then
    DOCKER_IMAGENAME="$D"
    echo ""
    echo "Building docker image for '$D'..."
    time docker build -t "object88/$DOCKER_IMAGENAME:latest" -f apps/$D/Dockerfile .
    docker tag "object88/$DOCKER_IMAGENAME:latest" "object88/$DOCKER_IMAGENAME:$UNIQUE_TAG"
    DOCKER_IMAGES+=("$DOCKER_IMAGENAME:$UNIQUE_TAG")
  fi
done

if [[ $DO_PUSH == "true" ]]; then
  echo "Pushing build images..."
  for I in "${DOCKER_IMAGES[@]}"; do
    echo "Pushing docker image 'object88/$I'"
    time docker push "object88/$I"
  done
  echo "Pushed images."
fi

echo "Finished building the docker images"
