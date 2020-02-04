#!/bin/sh

docker run --rm -ti -v `pwd`:/workdir --workdir=/workdir alpine/helm:3.0.3 package .
