#!/bin/sh
PROJECT="cormoran"
NOW=`date '+%Y%m%d%H%M%S'`
export GOPATH="`pwd`"
ls -1 src/${PROJECT}/exec | while read row ; do
  GOOS=linux   GOARCH=amd64 go install -ldflags "-s -w" ${PROJECT}/exec/$row
done
