# Alert Manager ConfigMap Properties

This document describes all the configuration options available in the alert-manager ConfigMap.

## ConfigMap Structure

The alert-manager ConfigMap consists of several sections that control different aspects of the controller's behavior:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alert-manager-config
  namespace: alert-manager-system
data:
  # Monitoring system endpoints
  wavefront.url: "https://try.wavefront.com"
  
  # Controller configuration
  controller.reconcile.interval: "5m"
  controller.max-concurrent-reconciles: "5"
  
  # Alert defaults
  defaults.severity: "INFO"
  defaults.resolve-minutes: "5"
  
  # Retry configuration
  retry.initial-interval: "5s"
  retry.max-interval: "1m"
  retry.max-attempts: "3"
  
  # Logging configuration
  logging.level: "info"
```

## Available Configuration Options

### Wavefront Settings

| Property | Default | Description |
|----------|---------|-------------|
| `wavefront.url` | `https://try.wavefront.com` | The URL of your Wavefront instance |
| `wavefront.timeout` | `30s` | HTTP request timeout for Wavefront API calls |
| `wavefront.batch-size` | `100` | Maximum number of alerts to process in a single batch |

### Controller Settings

| Property | Default | Description |
|----------|---------|-------------|
| `controller.reconcile.interval` | `5m` | How often to run full reconciliation |
| `controller.max-concurrent-reconciles` | `5` | Maximum number of concurrent reconciles |
| `controller.status-update-interval` | `1m` | How often to update status for resources |
| `controller.manager-workers` | `10` | Number of worker threads in the controller manager |

### Alert Defaults

| Property | Default | Description |
|----------|---------|-------------|
| `defaults.severity` | `INFO` | Default severity for alerts if not specified |
| `defaults.resolve-minutes` | `5` | Default resolve after minutes if not specified |
| `defaults.minutes` | `5` | Default alert trigger time if not specified |
| `defaults.alert-type` | `CLASSIC` | Default alert type if not specified |

### Retry Configuration

| Property | Default | Description |
|----------|---------|-------------|
| `retry.initial-interval` | `5s` | Initial retry interval for failed API calls |
| `retry.max-interval` | `1m` | Maximum retry interval (with exponential backoff) |
| `retry.max-attempts` | `3` | Maximum retry attempts before giving up |
| `retry.jitter-factor` | `0.1` | Jitter factor to add randomness to retry intervals |

### Logging Configuration

| Property | Default | Description |
|----------|---------|-------------|
| `logging.level` | `info` | Logging level (`debug`, `info`, `warn`, `error`) |
| `logging.format` | `json` | Log format (`json` or `text`) |
| `logging.development` | `false` | Enable development mode logging |
| `logging.stacktrace-level` | `error` | Level at which to capture stacktraces |

## Applying Configuration Changes

To update the configuration:

1. Edit the ConfigMap:
   ```bash
   kubectl edit configmap alert-manager-config -n alert-manager-system
   ```

2. Restart the controller to apply changes:
   ```bash
   kubectl rollout restart deployment alert-manager-controller-manager -n alert-manager-system
   ```

## Environment Variables

The following environment variables can be used to override ConfigMap settings:

| Environment Variable | Corresponding ConfigMap Key | Description |
|----------------------|----------------------------|-------------|
| `WAVEFRONT_URL` | `wavefront.url` | Wavefront instance URL |
| `RECONCILE_INTERVAL` | `controller.reconcile.interval` | Reconciliation interval |
| `LOG_LEVEL` | `logging.level` | Logging level |

These can be set in the controller Deployment specification.
