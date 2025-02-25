#!/bin/bash

set -euo pipefail
PROJECT_ROOT="$(realpath $(dirname $0)/..)"

function tidy() {
  go mod tidy -e
}

echo "Tidy root module ..."
(
  cd "$PROJECT_ROOT"
  tidy
)
