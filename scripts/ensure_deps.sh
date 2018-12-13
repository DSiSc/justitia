#!/bin/bash

RELEASE=${RELEASE:-master}

# Change directory to project root folder
PROJ_FOLDER=$(cd "$(dirname "$0")/..";pwd)
cd ${PROJ_FOLDER}

if [[ ! -f "release-notes/${RELEASE}.txt" ]]; then
  echo "[ERROR] we don't have release ${RELEASE}, please confirm."
  exit 1
fi

# Read "*.txt" under "release-notes"
DEPS=$(grep -v "^#" release-notes/${RELEASE}.txt | grep -v "^$")

# Go get all the imported packages (except the ones under "vendor" folder) to $GOPATH
# for dep in $DEPS; do
#   dep_repo=$(echo ${dep} | awk -F ':' '{print $1}')
#   if [ -d "${GOPATH}/src/${dep_repo}" ]; then
#     echo "1 checkout"
#     cd ${GOPATH}/src/${dep_repo}
#     git checkout master &> /dev/null
#   fi
#   echo "2 get"
#   go get -v -u ${dep_repo}
# done

for dep in ${DEPS}; do
  dep_repo=$(echo ${dep} | awk -F ':' '{print $1}')
  dep_ver=$(echo ${dep} | awk -F ':' '{print $2}')
  if [[ -d "${GOPATH}/src/${dep_repo}" ]]; then
    cd ${GOPATH}/src/${dep_repo}
    git checkout master &> /dev/null
  fi
done

go get -d -v ./...

# Check out to desired version
for dep in ${DEPS}; do
  dep_repo=$(echo ${dep} | awk -F ':' '{print $1}')
  dep_ver=$(echo ${dep} | awk -F ':' '{print $2}')
  if [[ -d "${GOPATH}/src/${dep_repo}" ]]; then

    echo -e "\n[${dep_repo}] >> [${dep_ver}]"

    cd ${GOPATH}/src/${dep_repo}

    git fetch origin > /dev/null

    # Try checkout to ${dep_ver}
    git checkout ${dep_ver} &> /dev/null && (git pull &> /dev/null | true)

    if [[ $? != 0 ]]; then
      # If failed, checkout to origin/${dep_ver}
      git checkout origin/${dep_ver} &> /dev/null
      if [[ $? != 0 ]]; then
        echo "[ERROR] when checking out ${dep_ver} under ${dep_repo}, please check."
        exit 1
      fi
    fi
#   echo "[INFO] ${dep_repo} is now on [${dep_ver}]"
    git log -n 1 --pretty=oneline
  else
    echo "[WARNING] ${GOPATH}/src/${dep_repo} not exist, do nothing, please check release-notes/$RELEASE.txt."
  fi
done

cd ${PROJ_FOLDER}
git checkout ${RELEASE}