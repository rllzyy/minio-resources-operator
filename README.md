# Minio Resources Operator

Kubernetes Operator that manage buckets and users on a Minio server.

## Usage

Create a `MinioServer`:

```yaml 
apiVersion: minio.walkbase.com/v1alpha1
kind: MinioServer
metadata:
  name: test
spec:
  hostname: myserver.example.com
  port: 9000
  ssl: false
```

Create a `MinioBucket`:

```yaml
apiVersion: minio.walkbase.com/v1alpha1
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
apiVersion: minio.walkbase.com/v1alpha1
kind: MinioUser
metadata:
  name: test
spec:
  server: test
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