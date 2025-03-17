# Alert Manager ConfigMap Properties

This document describes all the configuration options available in the alert-manager ConfigMaps.

## Required ConfigMaps

Alert Manager requires two ConfigMaps to function properly:

1. `alert-manager-controller-manager-config` - Used by the Kubernetes controller-runtime framework
2. `alert-manager-configmap` - Used directly by the alert-manager application code

> **Important:** Both ConfigMaps are required and serve different purposes. Missing either one will cause the controller to crash on startup.

## Why Two ConfigMaps?

The alert-manager controller uses two separate ConfigMaps for different purposes:

1. **Controller Framework ConfigMap**: `alert-manager-controller-manager-config`
   - This ConfigMap is used by the underlying Kubernetes controller-runtime framework
   - It configures system-level features like health probes, metrics, leader election
   - It also contains important environment variable values for the controller deployment

2. **Application ConfigMap**: `alert-manager-configmap`
   - This ConfigMap is explicitly loaded by the application code
   - It's defined as a constant (`AlertManagerConfigMapName = "alert-manager-configmap"`) in the source code
   - It contains application-specific configuration like backend type and URL

## Controller Manager ConfigMap

This ConfigMap is used by controller-runtime and contains configuration for health probes, metrics, webhook, and leader election:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alert-manager-controller-manager-config
  namespace: alert-manager-system
data:
  controller_manager_config.yaml: |
    apiVersion: controller-runtime.sigs.k8s.io/v1alpha1
    kind: ControllerManagerConfig
    health:
      healthProbeBindAddress: :8081
    metrics:
      bindAddress: 127.0.0.1:8080
    webhook:
      port: 9443
    leaderElection:
      leaderElect: true
      resourceName: 5eb85e31.keikoproj.io
  MONITORING_BACKEND_URL: "https://YOUR_WAVEFRONT_INSTANCE.wavefront.com"
  MONITORING_BACKEND_TYPE: "wavefront"
```

## Application ConfigMap

This ConfigMap contains application-specific configuration and is required by the controller:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alert-manager-configmap
  namespace: alert-manager-system
data:
  app.mode: "dev"
  base.url: "https://YOUR_WAVEFRONT_INSTANCE.wavefront.com"
  backend.type: "wavefront"
```

## Secret Configuration

In addition to ConfigMaps, alert-manager requires a secret for API token authentication:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: wavefront-api-token
  namespace: alert-manager-system
type: Opaque
stringData:
  wavefront-api-token: "YOUR_API_TOKEN_HERE"
```

> **IMPORTANT:** The secret key name (`wavefront-api-token`) must exactly match the secret name for the controller to correctly read the token.

## Available Configuration Properties

### Application ConfigMap Properties

| Property | Description | Example |
|----------|-------------|---------|
| `app.mode` | Application mode (dev/prod) | `"dev"` |
| `base.url` | URL of your Wavefront instance | `"https://example.wavefront.com"` |
| `backend.type` | Type of monitoring backend | `"wavefront"` |

### Controller Manager ConfigMap Properties

| Property | Description | Example |
|----------|-------------|---------|
| `MONITORING_BACKEND_URL` | URL of your monitoring backend | `"https://example.wavefront.com"` |
| `MONITORING_BACKEND_TYPE` | Type of monitoring backend | `"wavefront"` |

## Troubleshooting ConfigMap Issues

If you encounter issues with ConfigMaps:

1. **Check if both ConfigMaps exist**:
   ```bash
   kubectl get configmaps -n alert-manager-system
   ```
   
2. **Verify the content of each ConfigMap**:
   ```bash
   kubectl get configmap alert-manager-configmap -n alert-manager-system -o yaml
   kubectl get configmap alert-manager-controller-manager-config -n alert-manager-system -o yaml
   ```

3. **Check controller logs for ConfigMap-related errors**:
   ```bash
   kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager -c manager | grep -i configmap
   ```

4. **Common Error**: `configmaps "alert-manager-configmap" not found` indicates the application ConfigMap is missing.

For more detailed troubleshooting steps, see the [Troubleshooting Guide](troubleshooting.md#missing-configmap).
