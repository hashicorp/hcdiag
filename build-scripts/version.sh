#!/usr/bin/env bash

version_file=$1

# We exit the awk execution after the first match to avoid accidentally catching local variables
# named `version` or `prerelease`
version=$(awk '$1 == "version" && $2 == "=" { gsub(/"/, "", $3); print $3; exit; }' < "${version_file}")
prerelease=$(awk '$1 == "prerelease" && $2 == "=" { gsub(/"/, "", $3); print $3; exit; }' < "${version_file}")

if [ -n "$prerelease" ]; then
    echo "${version}-${prerelease}"
else
    echo "${version}"
fi
