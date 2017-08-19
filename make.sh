#! /bin/bash

NAME=route-switcher-build
docker build -t route-switcher-build:latest .
docker rm -f $NAME
docker run --name=$NAME -v $GOPATH:/data/go route-switcher-build:latest
