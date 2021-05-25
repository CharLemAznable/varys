#!/usr/bin/env bash

go build -ldflags="-s -w" -o varys.bin cmd/all/main.go
upx --brute varys.bin
