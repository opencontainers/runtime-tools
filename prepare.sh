#!/bin/bash
set -e

TESTROOT=/testroot/busybox
rm -rf ${TESTROOT}
mkdir -p ${TESTROOT}
tar -xf  rootfs.tar.gz -C ${TESTROOT}

cp runtimetest ${TESTROOT}

pushd $TESTROOT > /dev/null
ocitools generate --args /runtimetest
popd
