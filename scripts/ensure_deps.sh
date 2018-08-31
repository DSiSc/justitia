#!/bin/bash

# Change directory to project root folder
PROJ_FOLDER=$(cd "$(dirname "$0")/..";pwd)
cd $PROJ_FOLDER

# Read  "dependencies.txt" under project root
DEPS=$(grep -v "^#" dependencies.txt | grep -v "^$")

# Go get all the imported packages (except the ones under "vendor" folder) to $GOPATH
for dep in $DEPS; do
  dep_repo=$(echo ${dep} | awk -F ':' '{print $1}')
  if [ -d "${GOPATH}/src/${dep_repo}" ]; then
    cd ${GOPATH}/src/${dep_repo}
    git checkout master &> /dev/null
  fi
  go get -v -u ${dep_repo}
done

# Check out to desired version
for dep in $DEPS; do
  dep_repo=$(echo ${dep} | awk -F ':' '{print $1}')
  dep_ver=$(echo ${dep} | awk -F ':' '{print $2}')
  if [ -d "${GOPATH}/src/${dep_repo}" ]; then

    echo "[INFO] Ensuring ${dep_repo} on ${dep_ver} ..."

    cd ${GOPATH}/src/${dep_repo}

    git fetch origin > /dev/null

    # Try checkout to ${dep_ver}
    git checkout ${dep_ver} > /dev/null && (git pull &> /dev/null | true)

    if [ $? != 0 ]; then
      # If failed, checkout to origin/${dep_ver}
      git checkout origin/${dep_ver} > /dev/null
      if [ $? != 0 ]; then
        echo "[ERROR] Got error when checking out ${dep_ver} under ${dep_repo}, please check."
        exit 1
      else
        echo "[INFO] ${dep_repo} is now on ${dep_ver}"
      fi
    else
      echo "[INFO] ${dep_repo} is now on ${dep_ver}"
    fi
  else
    echo "[WARN] ${GOPATH}/src/${dep_repo} not exist, do nothing, please check dependencies.txt."
  fi
done

