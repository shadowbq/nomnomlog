#!/bin/sh

BUILDPATH="build/nomnomlog"

set -e

mkdir -p $BUILDPATH

go build -o $BUILDPATH/nomnomlog .
cp README.md LICENSE example_config.yml $BUILDPATH

cd $BUILDPATH/..
rm -f nomnomlog.tar.gz
tar -czf nomnomlog.tar.gz `basename $BUILDPATH`
#rm -r nomnomlog
