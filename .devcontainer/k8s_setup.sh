#!/bin/sh

set -e

kubectl create -f deploy/crds/minio.walkbase.com_miniousers_crd.yaml
kubectl create -f deploy/crds/minio.walkbase.com_miniobuckets_crd.yaml
kubectl create -f deploy/crds/minio.walkbase.com_minioservers_crd.yaml