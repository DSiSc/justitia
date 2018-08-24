#! /bin/bash

set -e

echo "" > coverage.txt

for pkg in $(go list ./... | grep -v vendor); do
  go test -timeout 5m -race -coverprofile=profile.cov -covermode=atomic "$pkg"
  if [ -f profile.cov ]; then
    cat profile.cov >> coverage.txt
    rm profile.cov
  fi
done
