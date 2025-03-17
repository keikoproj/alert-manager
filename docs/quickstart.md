# Alert Manager Quickstart Guide

This guide will help you quickly get started with alert-manager by walking through the installation process and creating your first alert.

## Prerequisites

Before you begin, ensure you have:

- A Kubernetes cluster (v1.16+)
- `kubectl` configured with admin access to your cluster
- If using Wavefront, a Wavefront account and API token

## Installation

There are two recommended ways to install alert-manager:

### Installation Using the Script (Recommended)

The easiest way to install alert-manager is using the provided installation script:

1. Clone the repository:
   ```bash
   git clone https://github.com/keikoproj/alert-manager.git
   cd alert-manager
   ```

2. Run the installation script:
   ```bash
   ./hack/install.sh <namespace> <monitoring_backend_url> <api_token>
   
   # Example:
   ./hack/install.sh alert-manager-system wavefront.example.com my-api-token
   ```

The script will:
- Create the namespace if it doesn't exist
- Apply all necessary Kubernetes resources (CRDs, RBAC, ConfigMaps, etc.)
- Create the required secrets with your API token
- Verify the deployment is successful

**Important:** The installation script handles many common issues automatically, such as:
- Creating the correct secret format for the Wavefront API token 
- Setting up proper RBAC permissions for secret access
- Creating the required ConfigMap

### Manual Installation

If you prefer to install manually:

1. Clone the repository:
   ```bash
   git clone https://github.com/keikoproj/alert-manager.git
   cd alert-manager
   ```

2. Create the necessary namespace:
   ```bash
   kubectl create namespace alert-manager-system
   ```

3. Create the Wavefront API token secret (**Note the specific format required**):
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: v1
   kind: Secret
   metadata:
     name: wavefront-api-token
     namespace: alert-manager-system
   type: Opaque
   stringData:
     wavefront-api-token: "YOUR_API_TOKEN_HERE"
   EOF
   ```

4. Create the required configuration ConfigMap:
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: alert-manager-configmap
     namespace: alert-manager-system
   data:
     app.mode: "dev"
     base.url: "YOUR_WAVEFRONT_URL"
     backend.type: "wavefront"
   EOF
   ```

5. Apply the RBAC rules:
   ```bash
   kubectl apply -f config/rbac/
   ```
   
6. Apply the CRDs:
   ```bash
   kubectl apply -f config/crd/bases/
   ```

7. Deploy the controller:
   ```bash
   kubectl apply -f config/manager/manager.yaml
   ```

### Verifying the Installation

To verify the installation was successful:

1. Check if the controller pod is running:
   ```bash
   kubectl get pods -n alert-manager-system
   ```
   You should see the controller pod with status `Running` and 2/2 containers ready.

2. Check the controller logs for any errors:
   ```bash
   kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager -c manager
   ```

If you encounter any issues, refer to the [Troubleshooting Guide](troubleshooting.md) for common problems and solutions.

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
- Configure different alert templates
- Add custom backend implementations
- Use different Kubernetes resources

## Troubleshooting

If you encounter issues, check the [Troubleshooting Guide](troubleshooting.md) or view the controller logs:

```bash
kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager
```
