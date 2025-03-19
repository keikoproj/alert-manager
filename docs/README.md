# Alert Manager Documentation

Welcome to the Alert Manager documentation. This directory contains resources to help you install, configure, and use Alert Manager.

## Getting Started

* [Quick Start Guide](quickstart.md) - Follow this guide to quickly set up Alert Manager and create your first alert
* [Installation Guide](install.md) - Detailed installation instructions for production environments

## Core Concepts

* [Architecture Documentation](architecture.md) - Learn about Alert Manager's architecture and components
* [Design Documentation](design.md) - Understand the design principles and decisions behind Alert Manager

## Configuration

* [Configuration Options](configmap-properties.md) - Detailed explanation of all available configuration options

## Advanced Topics

* [Troubleshooting and Debugging Guide](troubleshooting-and-debugging.md) - Comprehensive guide for diagnosing and fixing issues
* [Security Guide](security.md) - Best practices for securing Alert Manager deployments
* [Uninstallation](../hack/uninstall.sh) - Script to cleanly uninstall Alert Manager

## For Developers

* [Developer Guide](developer-guide.md) - Information for developers who want to contribute to Alert Manager
* [Contributing Guidelines](../CONTRIBUTING.md) - How to contribute to the Alert Manager project
* [Image Update Script](../hack/update-image.sh) - Script to update the controller image tag

## Examples

* [Sample Secret Configuration](sample-secret.yaml) - Example of how to configure secrets for Alert Manager
* [Comprehensive Alert Examples](../examples/comprehensive-alerts.yaml) - Various alert examples for different use cases
* [Sample WavefrontAlert](../hack/sample-wavefront-alert.yaml) - Basic example of a WavefrontAlert resource
* [Sample AlertsConfig](../hack/sample-alerts-config.yaml) - Example of using AlertsConfig for templated alerts

## Images

The [images](./images) directory contains diagrams illustrating Alert Manager architecture and workflows.

## Additional Resources

* [GitHub Repository](https://github.com/keikoproj/alert-manager) - Main repository for Alert Manager
* [Releases](https://github.com/keikoproj/alert-manager/releases) - Release notes and version information
* [Issues](https://github.com/keikoproj/alert-manager/issues) - Report bugs or request features
