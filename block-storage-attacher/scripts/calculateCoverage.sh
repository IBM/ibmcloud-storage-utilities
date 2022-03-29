#!/bin/bash
# shellcheck disable=SC2002

COVERAGE=$(cat "$TRAVIS_BUILD_DIR"/block-storage-attacher/cover.html | grep "%)"  | sed 's/[][()><%]/ /g' | awk '{ print $4 }' | awk '{s+=$1}END{print s/NR}')

echo "-------------------------------------------------------------------------"
echo "COVERAGE IS ${COVERAGE}%"
echo "-------------------------------------------------------------------------"
