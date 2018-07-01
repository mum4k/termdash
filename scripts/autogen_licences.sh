#!/bin/bash -eu
#
# Copyright 2018 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

BIN_DIR="${HOME}/bin"
INSTALL_DIR="${BIN_DIR}/autogen"
AUTOGEN="${INSTALL_DIR}/autogen"

if [ "$#" -eq 0 ]; then
  echo "Usage $0 <directory> [WRITE]"
  echo
  echo -n "Starts searching at <directory> for all Go files and adds licences "
  echo "by editing them in place."
  echo "Doesn't make any changes unless WRITE is the second argument."
  exit 1
fi

DIRECTORY="$1"

WRITE=""
if [ "$#" -ge 3 ]; then
  WRITE="$2"
fi

if [ ! -d "${BIN_DIR}" ]; then
  echo "Directory ${BIN_DIR} doesn't exist."
  exit 1
fi



if [ ! -d "${INSTALL_DIR}" ]; then
  git clone git@github.com:mbrukman/autogen.git "${BIN_DIR}/autogen"
  if [ $? -ne 0 ]; then
    echo "Failed to run git clone."
    exit 1
  fi
fi

if [ "${WRITE}" == "WRITE" ]; then
  DRY_RUN=""
else
  DRY_RUN="echo "
fi

ADD_LICENCE="${DRY_RUN}${AUTOGEN} -i --no-top-level-comment"
FIND_FILES="find ${DIRECTORY} -type f -name \*.go"
LICENCE="Licensed under the Apache License"

MISSING=0

for FILE in `eval ${FIND_FILES}`; do
  if ! grep -q "${LICENCE}" "${FILE}"; then
    MISSING=1
    eval "${ADD_LICENCE} ${FILE}"
  fi
done

if [[ ! -z "$DRY_RUN" ]] && [ $MISSING -eq 1 ]; then
      echo -e "\nFound files with missing licences. To fix, run the commands above."
      echo "Or just execute:"
      echo "$0 . WRITE"
fi
