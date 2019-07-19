# What's new

This chart has below new features added

- Change image registry to use icr.io
- Changes to support OpenShift


# Breaking Changes
None


# Fixes

Please refer v1.1.0 changelog section.


# Documentation
<Link of IBM block attacher needs to be added here>


# Prerequisites

1. IKS cluster with kube version 1.10 or higher


# Version History

| Chart | Date | Kubernetes Required | Breaking Changes | Details                    |
| ----- | ---------- | ------------ | ---------------- | --------------------------- |
| 1.1.0 | 2019-06-20 | >=1.10       | None             | Refer Changelog v1.1.0      |
| 1.0.2 | 2019-03-19 | >=1.10       | None             | Refer Changelog v1.0.2      |
| 1.0.1 | 2019-01-23 | >=1.10       | None             | Refer Changelog v1.0.1      |
| 1.0.0 | 2018-12-05 | >=1.10       | None             | Initial chart version       |

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
