#!/bin/sh

docker run --rm -ti -v `pwd`:/workdir --workdir=/workdir alpine/helm:3.1.1 package .
