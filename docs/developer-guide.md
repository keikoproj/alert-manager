# Developer Guide for Alert Manager

This guide provides instructions for developers who want to contribute to the alert-manager project.

## Project Overview

Alert Manager is built using [Kubebuilder](https://book.kubebuilder.io/), a framework for building Kubernetes APIs using custom resource definitions (CRDs). Kubebuilder provides scaffolding tools to quickly create new APIs, controllers, and webhook components. Understanding Kubebuilder will greatly help in comprehending the alert-manager codebase structure and development workflow.

## Development Environment Setup

### Prerequisites

- Go 1.19+ (check the current version in go.mod)
- Kubernetes cluster for testing (minikube, kind, or a remote cluster)
- Docker for building images
- kubectl CLI
- kustomize
- controller-gen

### Clone the Repository

```bash
git clone https://github.com/keikoproj/alert-manager.git
cd alert-manager
```

### Install Required Tools

The Makefile can help install required development tools:

```bash
# Install controller-gen
make controller-gen

# Install kustomize
make kustomize

# Install mockgen (for tests)
make mockgen
```

### Build the Project

```bash
# Build the manager binary
make

# Run the manager locally (outside the cluster)
make run
```

## Running Tests

### Unit Tests

```bash
# Run unit tests
make test
```

### Integration Tests

For integration tests, you need a Kubernetes cluster and Wavefront access:

```bash
# Set up environment variables for integration tests
export WAVEFRONT_URL=https://your-wavefront-instance.wavefront.com
export WAVEFRONT_TOKEN=your-api-token

# Run integration tests
make integration-test
```

## Creating and Deploying Custom Builds

### Building Docker Images

To build a custom Docker image:

```bash
# Build the controller image
make docker-build IMG=your-registry/alert-manager:your-tag

# Push the image to your registry
make docker-push IMG=your-registry/alert-manager:your-tag
```

### Deploying Custom Builds

Deploy your custom build to a cluster:

```bash
# Deploy with your custom image
make deploy IMG=your-registry/alert-manager:your-tag
```

## Code Structure

Here's an overview of the project structure:

```
.
├── api/                    # API definitions (CRDs)
│   └── v1alpha1/           # API version
├── cmd/                    # Entry points
├── config/                 # Kubernetes YAML manifests
├── controllers/            # Reconciliation logic
├── pkg/                    # Shared packages
│   ├── wavefront/          # Wavefront client
│   └── splunk/             # Splunk client (future)
└── hack/                   # Development scripts
```

### Key Components

- **api/v1alpha1**: Contains the CRD definitions, including the WavefrontAlert and AlertsConfig types.
- **controllers**: Contains the controllers that reconcile the custom resources.
- **pkg/wavefront**: Implements the Wavefront API client.

## Making Changes

### Adding a New Feature

1. Create a new branch: `git checkout -b feature/your-feature-name`
2. Make your changes
3. Add tests for your changes
4. Run tests: `make test`
5. Build and verify: `make`
6. Commit changes with DCO signature: `git commit -s -m "Your commit message"`
7. Push changes: `git push origin feature/your-feature-name`
8. Create a pull request

### Adding a New Monitoring System

To add support for a new monitoring system:

1. Create a new client package in `pkg/`
2. Define a new CRD in `api/v1alpha1/`
3. Create a new controller in `controllers/`
4. Update the controller manager to include your new controller
5. Add appropriate tests
6. Update documentation

## Debugging

### Running the Controller Locally

For easier debugging, you can run the controller outside the cluster:

```bash
# Run the controller locally
make run
```

### Remote Debugging

You can use Delve for remote debugging:

```bash
# Install Delve if you don't have it
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with Delve
dlv debug ./cmd/main.go -- --kubeconfig=$HOME/.kube/config
```

### Verbose Logging

To enable debug logs:

```bash
# When running locally
make run LOG_LEVEL=debug

# In a deployed controller
kubectl edit deployment alert-manager-controller-manager -n alert-manager-system
# Add environment variable LOG_LEVEL=debug
```

## Code Generation

alert-manager uses kubebuilder and controller-gen for code generation.

### Kubebuilder and controller-gen

Alert Manager follows the [Kubebuilder](https://book.kubebuilder.io/) project structure and conventions. The project was initially scaffolded using Kubebuilder, which set up:

- API types in `api/v1alpha1/`
- Controller logic in `controllers/`
- Configuration files in `config/`
- Main entry point in `cmd/manager/main.go`

When you make changes to the API types, you need to regenerate various files:

```bash
# Generate CRDs
make manifests

# Generate code (deepcopy methods, etc.)
make generate
```

### Adding New API Types

To add a new Custom Resource Definition (e.g., for a new monitoring system):

```bash
# Use kubebuilder to scaffold a new API
kubebuilder create api --group alertmanager --version v1alpha1 --kind YourNewAlert

# This will create:
# - api/v1alpha1/yournewealert_types.go
# - controllers/yournewealert_controller.go
# - And update main.go to include the new controller
```

After scaffolding, you'll need to:
1. Define your API schema in the `_types.go` file
2. Implement the reconciliation logic in the controller
3. Regenerate the manifests and code as described above

## Continuous Integration

The project uses GitHub Actions for CI. When you submit a PR, the CI will:

1. Run unit tests
2. Build the controller image
3. Verify code generation is up-to-date
4. Check code style

Make sure all CI checks pass before requesting a review.

## Releasing

To create a new release:

1. Update version tags in all relevant files
2. Run tests and ensure they pass
3. Create a git tag: `git tag -a v0.x.y -m "Release v0.x.y"`
4. Push the tag: `git push origin v0.x.y`
5. Create a release on GitHub with release notes
