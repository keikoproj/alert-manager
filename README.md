# alert-manager

[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)][GithubMaintainedUrl]
[![PR](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)][GithubPrsUrl]
[![slack](https://img.shields.io/badge/slack-join%20the%20conversation-ff69b4.svg)][SlackUrl]

[![Release][ReleaseImg]][ReleaseUrl]
[![Build Status][BuildStatusImg]][BuildMasterUrl]
[![codecov][CodecovImg]][CodecovUrl]
[![Go Report Card][GoReportImg]][GoReportUrl]

A Kubernetes operator that enables management of monitoring alerts as custom resources within your Kubernetes clusters.

## Table of Contents
- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Components](#components)
  - [WavefrontAlert CRD](#wavefrontalert-crd)
  - [AlertsConfig CRD](#alertsconfig-crd)
  - [Alert-Manager Controller](#alert-manager-controller)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Standard Installation](#standard-installation)
  - [Configuration](#configuration)
- [Usage](#usage)
  - [WavefrontAlert Example](#wavefrontalert-example)
  - [AlertsConfig Example](#alertsconfig-example)
- [Contributing](#contributing)
- [Troubleshooting](#troubleshooting)
- [Version Compatibility](#version-compatibility)

## Overview

alert-manager enables you to define and manage monitoring alerts as Kubernetes resources, allowing you to:

- Create alerts alongside your application deployments
- Version control your alert definitions
- Apply GitOps practices to your monitoring configuration
- Scale efficiently with templated alerts

Currently supported monitoring backends include:
- Wavefront
- Splunk (Phase 2)

## Architecture

![Alert Manager High Architecture](docs/images/alert-manager-arch.png)

alert-manager follows a Kubernetes operator pattern that watches for custom resources and reconciles them with the target monitoring systems.

For a more detailed view of the architecture including component interactions and workflows, see the [Architecture Documentation](docs/architecture.md).

## Features

- **Declarative alert management** - Define alerts using Kubernetes custom resources
- **Multiple monitoring systems** - Support for different monitoring backends
- **Templating** - Create reusable alert templates across applications
- **Scalable** - AlertsConfig allows efficient alert management without etcd bloat
- **GitOps compatible** - Manage alerts through the same pipeline as your applications

## Components

### WavefrontAlert CRD

A Kubernetes custom resource that defines a specific alert in Wavefront. This resource includes all the fields necessary to create an alert in Wavefront, including:

- Alert name and description
- Alert conditions
- Notification targets
- Severity levels
- Tags

### AlertsConfig CRD

A scalable solution for managing multiple similar alerts. Instead of creating thousands of individual alert CRs (which could cause etcd storage issues), AlertsConfig allows you to:

- Define a template for alerts once
- Apply that template to many applications/services with custom parameters
- Manage alerts at scale (e.g., 100 alerts for 450 applications can be managed with just 550 CRs instead of 45,000)

### Alert-Manager Controller

The controller that reconciles the custom resources and manages the actual alerts in the target systems (e.g., Wavefront). It handles:

- Creation of new alerts
- Updates to existing alerts
- Deletion of removed alerts
- Error handling and status reporting

## Installation

### Prerequisites

- Kubernetes cluster v1.16+
- kubectl configured with admin access
- For Wavefront alerts: A Wavefront account and API token

### Standard Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/keikoproj/alert-manager.git
   cd alert-manager
   ```

2. Update the ConfigMap with your target monitoring system URLs:
   ```bash
   # For Wavefront, modify the URL from try.wavefront.com to your Wavefront instance
   vim config/default/iammanager.keikoproj.io_iamroles-configmap.yaml
   ```

3. Create a Secret with your monitoring system credentials:
   - For Wavefront, create a Secret with your API key
   - Sample Secret template is available at [docs/Sample-Secret.yaml](docs/Sample-Secret.yaml)

4. Deploy the controller and CRDs:
   ```bash
   make deploy
   ```

### Configuration

For detailed configuration options, see the ConfigMap documentation [here](docs/configmap-properties.md).

## Usage

### WavefrontAlert Example

```yaml
apiVersion: alertmanager.keikoproj.io/v1alpha1
kind: WavefrontAlert
metadata:
  name: alert-sample
spec:
  # Add fields here
  alertType: CLASSIC
  alertName: test-alert2
  condition: ts(status.health)
  displayExpression: ts(status.health)
  minutes: 50
  resolveAfterMinutes: 5
  severity: severe
  tags:
    - test-alert
    - something-weird
```

and you should see the wavefront alert info in the status if the alert got successfully created

```yaml
apiVersion: alertmanager.keikoproj.io/v1alpha1
kind: WavefrontAlert
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"alertmanager.keikoproj.io/v1alpha1","kind":"WavefrontAlert","metadata":{"annotations":{},"name":"wavefrontalert-sample","namespace":"alert-manager-system"},"spec":{"alertName":"test-alert2","alertType":"CLASSIC","condition":"ts(status.health)","displayExpression":"ts(status.health)","minutes":50,"resolveAfterMinutes":5,"severity":"severe","tags":["test-alert","something-weird"]}}
  creationTimestamp: "2021-08-26T19:20:27Z"
  finalizers:
  - wavefrontalert.finalizers.alertmanager.keikoproj.io
  generation: 8
  name: wavefrontalert-sample
  namespace: alert-manager-system
  resourceVersion: "293921"
  uid: 7d108d5c-30a6-4f80-8acc-cdb6ef5fb785
spec:
  alertName: test-alert2
  alertType: CLASSIC
  condition: ts(status.health)
  displayExpression: ts(status.health)
  minutes: 50
  resolveAfterMinutes: 5
  severity: severe
  tags:
  - test-alert
  - something-weird
status:
  alertsStatus:
    test-alert2:
      alertName: test-alert2
      associatedAlert: {}
      associatedAlertsConfig: {}
      id: "1630005627450"
      lastChangeChecksum: 3a86ae56b46c66d51cf270dde6c469b7
      link: https://try.wavefront.com/alerts/1630005627450
      state: Ready
  lastChangeChecksum: 3a86ae56b46c66d51cf270dde6c469b7
  observedGeneration: 8
  state: Ready
```
For more template alerts user case and alerts config usage please refer here

### Installation
1. Update the config map with the target urls. For example: For WavefrontAlerts, it is pointed to try.wavefront.com and you must change it to your target url
2. Create a Secret with wavefront Api Key (for WavefrontAlerts). Please refer Sample Secret [here](docs/Sample-Secret.yaml)
3. Point your kubeconfig file to the k8s cluster where you want to create. ex: export KUBECONFIG=~/.kube/config
4. make deploy

### ❤ Contributing ❤

Please see [CONTRIBUTING.md](CONTRIBUTING.md).

### Developer Guide

Please see [DEVELOPER.md](.github/DEVELOPER.md).

## License

Apache License 2.0, see [LICENSE](LICENSE).

## Documentation

- [Architecture Documentation](docs/architecture.md)
- [Quick Start Guide](docs/quickstart.md)
- [Configuration Options](docs/configmap-properties.md)
- [Developer Guide](docs/developer-guide.md)
- [Troubleshooting Guide](docs/troubleshooting.md)

## Version Compatibility

| alert-manager Version | Kubernetes Version | Notable Features | Go Version |
|-----------------------|--------------------|------------------|------------|
| v0.5.0                | 1.22+              | Improved scalability, enhanced status reporting | 1.19+ |
| v0.4.0                | 1.20 - 1.24        | Template alerting, Splunk integration (beta) | 1.18+ | 
| v0.3.0                | 1.18 - 1.22        | Multi-cluster support, alert batching | 1.16+ |
| v0.2.0                | 1.16 - 1.20        | Initial AlertsConfig implementation | 1.15+ |
| v0.1.0                | 1.16+              | Initial release with Wavefront support | 1.13+ |

For detailed information about each release, see the [GitHub Releases page](https://github.com/keikoproj/alert-manager/releases).

<!-- Markdown link -->
[GithubMaintainedUrl]: https://github.com/keikoproj/alert-manager/graphs/commit-activity
[GithubPrsUrl]: https://github.com/keikoproj/alert-manager/pulls
[SlackUrl]: https://keikoproj.slack.com/messages/alert-manager

[ReleaseImg]: https://img.shields.io/github/release/keikoproj/alert-manager.svg
[ReleaseUrl]: https://github.com/keikoproj/alert-manager/releases/latest

[BuildStatusImg]: https://github.com/keikoproj/alert-manager/actions/workflows/unit_test.yaml/badge.svg
[BuildMasterUrl]: https://github.com/keikoproj/alert-manager/actions/workflows/unit_test.yaml

[CodecovImg]: https://codecov.io/gh/keikoproj/alert-manager/branch/master/graph/badge.svg
[CodecovUrl]: https://codecov.io/gh/keikoproj/alert-manager

[GoReportImg]: https://goreportcard.com/badge/github.com/keikoproj/alert-manager
[GoReportUrl]: https://goreportcard.com/report/github.com/keikoproj/alert-manager
