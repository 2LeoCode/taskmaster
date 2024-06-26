#!/bin/bash

docker build $(dirname "$0") -f Dockerfile.dev 
docker run -it -v /dev:/dev -v /proc:/proc bash
