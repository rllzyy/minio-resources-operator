# Minio Resources Operator

Kubernetes Operator that manage resource on a Minio server.

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

Soon a Helm chart is coming.

## Usage

Create bucket and user using CR. Look at:

- `deploy/crds/minio.robotinfra.com_v1alpha1_miniobucket_cr.yaml`
- `deploy/crds/minio.robotinfra.com_v1alpha1_miniouser_cr.yaml`

for example.
