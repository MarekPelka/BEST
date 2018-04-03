#!/bin/bash

if [ -z ${1+x} ]; then 
    EXR="Exited"
else 
    EXR=$1
fi

sudo docker ps -a | grep $EXR | cut -d ' ' -f 1 | xargs sudo docker rm -v
