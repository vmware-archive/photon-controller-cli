#!/bin/bash -x

# Terminate script on error
set -e

# Fail if there is any go fmt error.
if [[ -n "$(gofmt -l photon)" ]]; then
  echo Fix gofmt errors
  gofmt -d photon
  exit 1
fi
