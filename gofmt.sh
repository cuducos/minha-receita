#!/bin/sh -l

GOFMT="$(gofmt -d .)"
if [ -n "$GOFMT" ]; then
	echo "Please, run go fmt in the following files:"
    gofmt -l .
	exit 1
fi
