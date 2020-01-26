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

A Helm chart is in `deploy/`.

## Usage

Create a `MinioServer`:

```yaml 
apiVersion: minio.robotinfra.com/v1alpha1
kind: MinioServer
metadata:
  name: test
spec:
  hostname: myserver.example.com
  port: 9000
  accessKey: admin
  secretKey: testtest
  ssl: false
```

Create a `MinioBucket`:

```yaml
apiVersion: minio.robotinfra.com/v1alpha1
kind: MinioBucket
metadata:
  name: bucket
spec:
  name: mybucket
  server: test
  policy: |
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Action": [
            "s3:GetObject"
          ],
          "Effect": "Allow",
          "Principal": {
            "AWS": ["*"]
          },
          "Resource": [
            "arn:aws:s3:::mybucket/*"
          ],
          "Sid": ""
        }
      ]
    }

```

Create a `MinioUser`:

```yaml
apiVersion: minio.robotinfra.com/v1alpha1
kind: MinioUser
metadata:
  name: test
spec:
  server: test
  accessKey: myUsername
  secretKey: mySecurePassword
  policy: |
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Action": [
            "s3:*"
          ],
          "Effect": "Allow",
          "Resource": [
            "arn:aws:s3:::mybucket/*",
            "arn:aws:s3:::mybucket"
          ],
          "Sid": ""
        }
      ]
    }
```