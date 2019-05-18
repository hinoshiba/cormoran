#!/bin/sh
export GOPATH
GOPATH="`pwd`"
cd src/cormoran
dep ensure
dep status
