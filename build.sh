#!/bin/sh
BUILDPATH="build/nomnomlog"

set -e
mkdir -p $BUILDPATH
VERSION=`cat VERSION`
echo "Building nomnomlog ${VERSION} local-release, use Makefile for distro-releases."
go build -o $BUILDPATH/nomnomlog -ldflags="-X main.Version=${VERSION}" .
cp README.md LICENSE example_config.yml $BUILDPATH

cd $BUILDPATH/..
rm -f nomnomlog.tar.gz
tar -czf nomnomlog.tar.gz `basename $BUILDPATH`
#rm -r nomnomlog
