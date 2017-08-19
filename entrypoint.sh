#! /bin/bash

export GOPATH=/data/go
export PATH=$PATH:/usr/local/go/bin
cd /data/go/src/github.com/creamfinance/route-switcher

rm -rf route-switcher

make "$@"

# ./route-switcher --external-interfaces eth0 --ping-targets 8.8.8.8,195.34.145.210
