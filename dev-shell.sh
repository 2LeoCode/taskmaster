#!/bin/bash

docker build $(dirname "$0") -f Dockerfile.dev -t taskmaster-dev
docker run -it -v /dev:/dev -v /proc:/proc -v .:/usr/src/taskmaster taskmaster-dev bash 
