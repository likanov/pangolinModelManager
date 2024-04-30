#!/bin/sh
VERSION=`git rev-parse --short HEAD`
go build -ldflags "-X main.version=$VERSION"