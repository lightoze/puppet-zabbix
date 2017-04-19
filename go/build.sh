#!/bin/sh

GOARCH=386 go build -ldflags="-s -w" -o ../files/jolokia-discovery.x86 jolokia-common.go jolokia-discovery.go
GOARCH=amd64  go build -ldflags="-s -w" -o ../files/jolokia-discovery.x86_64 jolokia-common.go jolokia-discovery.go
GOARCH=386 go build -ldflags="-s -w" -o ../files/jolokia-read.x86 jolokia-common.go jolokia-read.go
GOARCH=amd64  go build -ldflags="-s -w" -o ../files/jolokia-read.x86_64 jolokia-common.go jolokia-read.go
