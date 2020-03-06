#!/usr/bin/env bash
export GOPATH=$(pwd)
go build -ldflags="-s -w" -o bin/server server/main.go
go build -ldflags="-s -w" -o bin/client client/main.go
