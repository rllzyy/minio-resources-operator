#!/bin/sh

kubectl delete -f deploy/crds/minio.robotinfra.com_miniobuckets_crd.yaml
kubectl delete -f deploy/crds/minio.robotinfra.com_miniousers_crd.yaml
