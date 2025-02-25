#!/bin/bash -eu

set -euo pipefail

if [[ -n ${EFFECTIVE_VERSION:-""} ]] ; then
  # running in the pipeline use the provided EFFECTIVE_VERSION
  echo "$EFFECTIVE_VERSION"
  exit 0
fi

PROJECT_ROOT="$(realpath $(dirname $0)/..)"
VERSION="$(cat "${PROJECT_ROOT}/VERSION")"

(
  cd "$PROJECT_ROOT"

  if [[ "$VERSION" = *-dev ]] ; then
    VERSION="$VERSION-$(git rev-parse HEAD)"
  fi
  
  echo "$VERSION"
)
