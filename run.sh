#!/bin/bash

# add option for optional compilation
MAINDIR=$PWD
cd goapp/
rm main
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
cd $MAINDIR
sudo docker-compose up
