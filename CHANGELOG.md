# Changelog

## Unreleased

### Added

### Changed

### Deprecated

### Removed

### Bug Fixes

## 0.3.3

### Bug Fixes

- Fix panic

## 0.3.2

### Added

- Set k8s controller reference for user and bucket
- Publish helm chart

### Bug Fixes

- Fix helm chart

## 0.3.1

### Bug Fixes

- Fix metrics initialization.

## 0.3

### Added

- Helm chart in `deploy/`

### Changed

- `MinioServer` CRD is now cluster scoped.
- Can manage `MinioUser` and `MinioBucket` in all namespaces.

### Bug Fixes

- Work around [Operator SDK #1858](https://github.com/operator-framework/operator-sdk/issues/1858)
- Restrict cluster permissions.

## v0.2

Main change: support multiple Minio servers.

### Added

- Add MinioServer CRD.
- Bucket and User CRD now depends on MinioServer.

### Removed

- CRDs status are now empty.

## v0.1

Initial release.
