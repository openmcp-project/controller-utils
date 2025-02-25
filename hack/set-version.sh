#!/bin/bash

set -euo pipefail
PROJECT_ROOT="$(realpath $(dirname $0)/..)"

HACK_DIR="$PROJECT_ROOT/hack"

VERSION=$1

# update VERSION file
echo "$VERSION" > "$PROJECT_ROOT/VERSION"
