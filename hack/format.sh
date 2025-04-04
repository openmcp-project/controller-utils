#!/bin/bash

set -euo pipefail
PROJECT_ROOT="$(realpath $(dirname $0)/..)"

if [[ -z ${LOCALBIN:-} ]]; then
  LOCALBIN="$PROJECT_ROOT/bin"
fi
if [[ -z ${FORMATTER:-} ]]; then
  FORMATTER="$LOCALBIN/goimports"
fi

write_mode="-w"
if [[ ${1:-} == "--verify" ]]; then
  write_mode=""
  shift
fi

tmp=$("${FORMATTER}" -l $write_mode -local=github.com/openmcp-project/controller-utils $("$PROJECT_ROOT/hack/unfold.sh" --clean --no-unfold "$@"))

if [[ -z ${write_mode} ]] && [[ ${tmp} ]]; then
  echo "unformatted files detected, please run 'make format'" 1>&2
  echo "$tmp" 1>&2
  exit 1
fi

if [[ ${tmp} ]]; then
  echo "> Formatting imports ..."
  echo "$tmp"
fi
