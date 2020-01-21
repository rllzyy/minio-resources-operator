#!/bin/sh

set -e

kubectl create -f deploy/crds/minio.robotinfra.com_miniousers_crd.yaml
kubectl create -f deploy/crds/minio.robotinfra.com_miniobuckets_crd.yaml
