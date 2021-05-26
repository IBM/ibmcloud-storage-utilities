# What's new
This chart has below new features added
- UBI image update for VA issue

# Breaking Changes
None

# Fixes
Please refer v2.0.7 changelog section.

# Documentation
https://cloud.ibm.com/docs/containers?topic=containers-utilities#block_storage_attacher

# Prerequisites
1. IKS cluster with kube version 1.10 or higher

# Version History

| Chart  | Date       | Kubernetes Required | Breaking Changes | Details                     |
| -----  | ---------- | ------------------- | ---------------- | --------------------------- |
| v2.0.7 | 2021-05-26 | >=1.10              | None             | Refer Changelog v2.0.7      |
| v2.0.6 | 2021-04-27 | >=1.10              | None             | Refer Changelog v2.0.6      |
| v2.0.5 | 2021-04-15 | >=1.10              | None             | Refer Changelog v2.0.5      |
| v2.0.4 | 2021-03-22 | >=1.10              | None             | Refer Changelog v2.0.4      |
| v2.0.3 | 2020-12-21 | >=1.10              | None             | Refer Changelog v2.0.3      |
| v2.0.1 | 2020-12-09 | >=1.10              | None             | Refer Changelog v2.0.1      |
| 2.0.0  | 2020-11-13 | >=1.10              | None             | Refer Changelog v2.0.0      |
| 1.1.4  | 2020-02-19 | >=1.10              | None             | Refer Changelog v1.1.4      |
| 1.1.3  | 2019-11-25 | >=1.10              | None             | Refer Changelog v1.1.3      |
| 1.1.2  | 2019-09-12 | >=1.10              | None             | Refer Changelog v1.1.2      |
| 1.1.1  | 2019-08-28 | >=1.10              | None             | Refer Changelog v1.1.1      |
| 1.1.0  | 2019-06-20 | >=1.10              | None             | Refer Changelog v1.1.0      |
| 1.0.2  | 2019-03-19 | >=1.10              | None             | Refer Changelog v1.0.2      |
| 1.0.1  | 2019-01-23 | >=1.10              | None             | Refer Changelog v1.0.1      |
| 1.0.0  | 2018-12-05 | >=1.10              | None             | Initial chart version       |

## [v2.0.7] - 2021-05-26
UBI image update for VA issue

### Changelog
- UBI image update for VA issue

## [v2.0.6] - 2021-04-27
UBI image patch update for VA issue

### Changelog
- UBI image patch update for VA issue

## [v2.0.5] - 2021-04-15
UBI image update for VA issue

### Changelog
- UBI image update for VA issue

## [v2.0.4] - 2021-03-22
Updated Golang

### Changelog
- Use GO Lang 1.15.9 version
- Upgraded x/text version

## [v2.0.3] - 2020-12-21
UBI image update for VA issue

### Changelog
- UBI image update for VA issue

## [v2.0.1] - 2020-12-09
FS Cloud changes and update GO Lang

### Changelog
- Artifactory, image signing and linking resource
- Use GO Lang 1.15.5 version

## [v2.0.0] - 2020-11-13
Upgrading to use UBI base image and security changes for non-root user

### Changelog
- We use UBI base image
- We made changes for security context for non-root user_name
- Use GO Lang 1.15.2 version

## [v1.1.4] - 2020-02-19
Changing the API version of Daemon set to support kubernetes 1.16.5

### Changelog
- Updated the API version of Daemon set to apps/v1

## [v1.1.3] - 2019-11-25
Adding infinite retry for attach and improving error handling

### Changelog
- Adding infinite retry for attach flow and improving error handling.
- Added queues to accept multiple attach volume requests.
- Added mutex locks for attach and detach.

## [v1.1.2] - 2019-09-12
Fix for detach PV impacting other attacher pods

### Changelog
- Changes to pv_watcher to ignore detach requests of other nodes.

## [v1.1.1] - 2019-08-28
Update the GO lang version of block attacher image to 1.12.9

### Changelog
- Changes to Dockerfile.builder file

## [v1.1.0] - 2019-06-20
This is the release for OpenShift support

### Changelog
- Change image registry to use icr.io
- Changes to support OpenShift

## [v1.0.2] - 2019-03-19
This is a bug fix release.

### Changelog
- Update the GO lang version of block attacher image to 1.12.1
- CVE-2019-6486 is fixed in this update.
- CVE-2019-9741 is fixed in this update.

## [v1.0.1] - 2019-01-23
This is a bug fix release.

### Changelog
- Update the GO lang version of block attacher image to 1.11.4

## [v1.0.0] - 2018-12-05
This is the initial release.

### Changelog
- Attacher component.
- Helm chart.
