#!/usr/bin/env bash

go build -ldflags="-s -w" -o varys.bin
upx --brute varys.bin
