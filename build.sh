#This file makes a tlt binary compiled for 64-Bit Windows, AMD Linux and ARM64 Linux.
#Comment out as needed
#!/bin/bash

export GOARCH=amd64

export GOOS=linux
go build -o tlt.amd64

export GOOS=windows
go build -o tlt.exe

export GOOS=linux
export GOARCH=arm64
go build -o tlt.arm64
