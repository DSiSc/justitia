#! /bin/bash

set -e

# Change directory to project root folder
PROJ_FOLDER=$(cd "$(dirname "$0")/..";pwd)
cd $PROJ_FOLDER

echo "" > coverage.txt

for pkg in $(go list ./... | grep -v vendor); do
  go test -timeout 5m -race -coverprofile=profile.cov -covermode=atomic "$pkg"
  if [ -f profile.cov ]; then
    cat profile.cov >> coverage.txt
    rm profile.cov
  fi
done
