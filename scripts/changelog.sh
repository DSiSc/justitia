#!/bin/sh
#
# Copyright(c) 2018 justitia Group. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -x

SCRIPT_DIR=$(readlink -f "$(dirname $0)")
CHANGELOG_TEMP="CHANGELOG.new"

echo "## $2\n$(date)" >> ${CHANGELOG_TEMP}
echo "" >> ${CHANGELOG_TEMP}
git log $1..HEAD  --oneline | grep -v Merge | sed -e "s/\([0-9|a-z]*\)/* \[\1\](https:\/\/github.com\/DSiSc\/justitia\/commit\/\1)/" >> ${CHANGELOG_TEMP}
echo "" >> ${CHANGELOG_TEMP}
cat ${SCRIPT_DIR}/../CHANGELOG.md >> ${CHANGELOG_TEMP}
mv -f ${CHANGELOG_TEMP} CHANGELOG.md
