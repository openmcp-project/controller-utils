#!/bin/bash

set -euo pipefail
PROJECT_ROOT="$(realpath $(dirname $0)/..)"

HACK_DIR="$PROJECT_ROOT/hack"

VERSION=$("$HACK_DIR/get-version.sh")

echo "> Finding latest release"
major=${VERSION%%.*}
major=${major#v}
minor=${VERSION#*.}
minor=${minor%%.*}
patch=${VERSION##*.}
patch=${patch%%-*}
echo "v${major}.${minor}.${patch}"
echo

semver=${1:-"minor"}

case "$semver" in
  ("major")
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  ("minor")
    minor=$((minor + 1))
    patch=0
    ;;
  ("patch")
    patch=$((patch + 1))
    ;;
  (*)
    echo "invalid argument: $semver"
    exit 1
    ;;
esac

release_version="v$major.$minor.$patch"

echo "The release version will be $release_version. Please confirm with 'yes' or 'y':"
read confirm

if [[ "$confirm" != "yes" ]] && [[ "$confirm" != "y" ]]; then
  echo "Release not confirmed."
  exit 0
fi
echo

echo "> Updating version to release version"
"$HACK_DIR/set-version.sh" $release_version
echo

git add --all
git commit -m "release $release_version"
git push
echo

echo "> Successfully finished"
