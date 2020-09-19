#!/bin/bash -e
docker build -t replicate .
docker run -it --rm -v `pwd`:/src -w /src replicate bash
