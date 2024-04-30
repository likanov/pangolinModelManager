#!/bin/sh
VERSION=`git rev-parse --short HEAD`
GOOS=windows GOARCH=amd64 go build -o bin/pangolinModelManager-amd64.exe -ldflags "-X main.version=$VERSION" main.go
