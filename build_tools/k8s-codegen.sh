#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-./vendor/k8s.io/code-generator}

TEMPDIR=$(mktemp -d)

bash vendor/k8s.io/code-generator/generate-groups.sh all \
  github.com/object88/tugboat/pkg/k8s/client \
  github.com/object88/tugboat/pkg/k8s/apis \
  engineering.tugboat:v1alpha1 \
  --go-header-file ${SCRIPT_ROOT}/build_tools/custom-boilerplate.go.txt \
  --output-base ${TEMPDIR}

echo "generation complete"

rsync -a ${TEMPDIR}/github.com/object88/tugboat/pkg/k8s/apis/ ${SCRIPT_ROOT}/pkg/k8s/apis
rsync -a ${TEMPDIR}/github.com/object88/tugboat/pkg/k8s/client/ ${SCRIPT_ROOT}/pkg/k8s/client
