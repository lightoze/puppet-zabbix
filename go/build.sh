#!/bin/sh

GOARCH=386 go build -o ../files/jolokia-discovery.x86 jolokia-discovery.go
GOARCH=amd64  go build -o ../files/jolokia-discovery.x86_64 jolokia-discovery.go
GOARCH=386 go build -o ../files/jolokia-read.x86 jolokia-read.go
GOARCH=amd64  go build -o ../files/jolokia-read.x86_64 jolokia-read.go
