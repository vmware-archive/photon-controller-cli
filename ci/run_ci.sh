#!/bin/bash -x

# Terminate script on error
set -e

export GOPATH=$WORKSPACE
export PATH=$PATH:$WORKSPACE/bin

# $WORKSPACE will be the root of the git repo that is pulled in by Jenkins.
# We need to move its contents into the expected package path inside
# $GOPATH/src (defined by PACKAGESRC) before we can build.
PACKAGESRC=src/github.com/vmware/photon-controller-cli

REPOFILES=(*)
mkdir -p $PACKAGESRC
cp -r ${REPOFILES[*]} $PACKAGESRC/
pushd $PACKAGESRC

# Fail if there is any go fmt error.
if [[ -n "$(gofmt -l photon)" ]]; then
	echo Fix gofmt errors
	gofmt -d photon
	exit 1
fi

make all

rm -rf $WORKSPACE/bin
mv bin $WORKSPACE/.
