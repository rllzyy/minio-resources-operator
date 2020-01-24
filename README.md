# Minio Resources Operator

Kubernetes Operator that manage buckets and users on a Minio server.

[![Docker Pulls](https://img.shields.io/docker/pulls/robotinfra/minio-resources-operator.svg?maxAge=604800)](https://hub.docker.com/r/robotinfra/minio-resources-operator)

## Develop

- Open directory in [VSCode as a container](https://code.visualstudio.com/docs/remote/containers).
- Configure Kubernetes client in the container, such as create a `/root/.kube/config` file.
- Run task `Install CRDs` to create CRD.

You can run operator by running task `Run Operator`.

## Installation

Install operator (look in `deploy` directory):

- `crds/*_crd.yaml`
- `operator.yaml`
- `role_binding.yaml`
- `role.yaml`
- `service_account.yaml`

A Helm chart is coming soon.

## Usage

Create bucket and user using CR. Look at:

- `deploy/crds/minio.robotinfra.com_v1alpha1_miniobucket_cr.yaml`
- `deploy/crds/minio.robotinfra.com_v1alpha1_miniouser_cr.yaml`
- `deploy/crds/minio.robotinfra.com_v1alpha1_minioserver_cr.yaml`

for example.
