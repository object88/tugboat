#! /usr/bin/env bash

set -e

cd $(dirname "$0")

docker build . --tag diagrams:local

docker run -t --mount type=bind,src=$PWD/sources,dst=/sources diagrams:local --base-dir /sources