#!/bin/bash

for e in `gb info`; do
	e=`echo $e | sed 's/"//g'`
	export $e
done

mkdir $GB_PROJECT_DIR/bin 2> /dev/null

set -e

for project in `ls $GB_PROJECT_DIR/src`; do
	echo "Building $project"
	GOBIN="$GB_PROJECT_DIR/bin" GOPATH="$GB_PROJECT_DIR:$GB_PROJECT_DIR/vendor" go install -gcflags=-trimpath=$GB_PROJECT_DIR -asmflags=-trimpath=$GB_PROJECT_DIR src/$project/main.go
	strip bin/main
	mv bin/main bin/$project
done

