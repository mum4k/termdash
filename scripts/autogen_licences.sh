#!/bin/bash

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
WRITE="$2"

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
  echo "The WRITE argument not specified, dry run mode."
  echo "Would have executed:"
fi

ADD_LICENCE="${DRY_RUN}${AUTOGEN} -i --no-top-level-comment"
FIND_FILES="find ${DIRECTORY} -type f -name \*.go"
LICENCE="Licensed under the Apache License"

for FILE in `eval ${FIND_FILES}`; do
  if ! grep -q "${LICENCE}" "${FILE}"; then
    eval "${ADD_LICENCE} ${FILE}"
  fi
done
