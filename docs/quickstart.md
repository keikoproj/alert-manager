# Alert Manager Quickstart Guide

This guide will help you quickly get started with alert-manager by walking through the installation process and creating your first alert.

## Prerequisites

Before you begin, ensure you have:

- A Kubernetes cluster (v1.16+)
- `kubectl` configured with admin access to your cluster
- If using Wavefront, a Wavefront account and API token

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/keikoproj/alert-manager.git
cd alert-manager
```

### 2. Configure the Controller

Create a Wavefront API token secret:

```bash
kubectl create namespace alert-manager-system

# Replace YOUR_WAVEFRONT_TOKEN with your actual Wavefront API token
kubectl create secret generic wavefront-api-key \
  --from-literal=token=YOUR_WAVEFRONT_TOKEN \
  -n alert-manager-system
```

### 3. Update the ConfigMap

Edit the ConfigMap to specify your Wavefront URL:

```bash
# Open the ConfigMap YAML file
vim config/manager/alertmanager_config.yaml

# Update the URL to your Wavefront instance
# Change from try.wavefront.com to your actual Wavefront URL
```

### 4. Deploy the Controller

```bash
make deploy
```

### 5. Verify the Installation

```bash
kubectl get pods -n alert-manager-system
```

You should see the alert-manager-controller-manager pod running.

## Creating Your First Alert

### Basic Wavefront Alert

Create a file named `my-first-alert.yaml`:

```yaml
apiVersion: alertmanager.keikoproj.io/v1alpha1
kind: WavefrontAlert
metadata:
  name: my-first-alert
  namespace: default
spec:
  alertType: CLASSIC
  alertName: "High CPU Usage"
  condition: "ts(kubernetes.node.cpu.usage.total.percent) > 80"
  displayExpression: "ts(kubernetes.node.cpu.usage.total.percent)"
  minutes: 5
  resolveAfterMinutes: 5
  severity: WARN
  tags:
    - cpu
    - node
    - quickstart
  target:
    - "team@example.com"  # Replace with your notification target
```

Apply the alert to your cluster:

```bash
kubectl apply -f my-first-alert.yaml
```

### Check Alert Status

```bash
kubectl get wavefrontalert my-first-alert -n default -o yaml
```

You should see the status field populated with information about your alert, including its ID and a link to view it in Wavefront.

## Creating Alert Templates with AlertsConfig

For more advanced usage, you can create alert templates that can be applied to multiple services.

### 1. Create an Alert Template

Create a file named `cpu-alert-template.yaml`:

```yaml
apiVersion: alertmanager.keikoproj.io/v1alpha1
kind: WavefrontAlert
metadata:
  name: cpu-alert-template
  namespace: alert-templates
spec:
  alertType: CLASSIC
  alertName: "High CPU Usage - {{ .appName }}"
  condition: "ts(kubernetes.pod.cpu.usage.total.percent, app={{ .appName }}) > {{ .threshold }}"
  displayExpression: "ts(kubernetes.pod.cpu.usage.total.percent, app={{ .appName }})"
  minutes: {{ .duration }}
  resolveAfterMinutes: 5
  severity: {{ .severity }}
  tags:
    - cpu
    - {{ .environment }}
    - {{ .appName }}
  target:
    - "{{ .teamEmail }}"
```

Apply the template:

```bash
kubectl create namespace alert-templates
kubectl apply -f cpu-alert-template.yaml
```

### 2. Create an AlertsConfig

Create a file named `my-alerts-config.yaml`:

```yaml
apiVersion: alertmanager.keikoproj.io/v1alpha1
kind: AlertsConfig
metadata:
  name: my-app-alerts
  namespace: default
spec:
  provider: wavefront
  parameters:
    appName: my-application
    threshold: 75
    duration: 10
    severity: WARN
    environment: production
    teamEmail: team@example.com
  alerts:
    - type: cpu-alert-template
      enabled: true
      namespaceSelector: alert-templates
```

Apply the AlertsConfig:

```bash
kubectl apply -f my-alerts-config.yaml
```

## Next Steps

- Explore more [Alert Manager Examples](examples/)
- Learn about [Configuration Options](docs/ConfigMap_Properties.md)
- Read the [Architecture Documentation](architecture.md)
- Set up [IRSA Integration](docs/aws-integration.md) if running on AWS

## Troubleshooting

If you encounter issues, check the [Troubleshooting Guide](troubleshooting.md) or view the controller logs:

```bash
kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager
```
