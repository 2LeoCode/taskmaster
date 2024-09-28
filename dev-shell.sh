#!/bin/bash

docker build $(dirname "$0") -f Dockerfile.dev -t taskmaster-dev
docker run -it -v .:/usr/src/taskmaster taskmaster-dev bash 
