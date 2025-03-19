# Alert Manager Troubleshooting and Debugging Guide

This comprehensive guide provides solutions for common issues, debugging techniques, and diagnostic procedures for alert-manager.

## Table of Contents
- [Common Issues and Solutions](#common-issues-and-solutions)
  - [Installation Issues](#installation-issues)
  - [Controller Issues](#controller-issues)
  - [Wavefront Alert Issues](#wavefront-alert-issues)
  - [AlertsConfig Issues](#alertsconfig-issues)
  - [General Kubernetes Issues](#general-kubernetes-issues)
- [Advanced Debugging Techniques](#advanced-debugging-techniques)
  - [Enabling Debug Logs](#enabling-debug-logs)
  - [Common Debug Scenarios](#common-debug-scenarios)
  - [Debugging Alert Creation](#debugging-alert-creation)
- [Collecting Information for Support](#collecting-information-for-support)

## Common Issues and Solutions

This section provides quick solutions for common issues you might encounter when using alert-manager.

### Installation Issues

#### Controller Pod Not Starting

**Symptoms**: The alert-manager-controller-manager pod is not starting or stays in a pending/crash loop state.

**Possible Causes and Solutions**:

1. **Resource constraints**:
   ```bash
   kubectl describe pod -n alert-manager-system alert-manager-controller-manager
   ```
   Look for resource-related errors. If necessary, adjust resource requests/limits in the Deployment.

2. **Image pull issues**:
   ```bash
   kubectl get events -n alert-manager-system
   ```
   Look for image pull errors. Check if the image exists and is accessible from your cluster.

3. **RBAC issues**:
   ```bash
   kubectl describe pod -n alert-manager-system alert-manager-controller-manager
   ```
   Look for RBAC permission errors. Ensure the service account has necessary permissions:
   ```bash
   kubectl describe clusterrole alert-manager-manager-role
   kubectl describe clusterrolebinding alert-manager-manager-rolebinding
   ```

#### Cache Synchronization Timeouts

**Symptoms**: The controller pod crashes with errors like:
```
failed to wait for alertsconfig caches to sync: timed out waiting for cache to be synced for Kind *v1alpha1.AlertsConfig
```

**Solutions**:

1. **Restart the deployment**:
   ```bash
   kubectl rollout restart deployment alert-manager-controller-manager -n alert-manager-system
   ```

2. **Scale down and up**:
   ```bash
   kubectl scale deployment alert-manager-controller-manager --replicas=0 -n alert-manager-system
   # Wait a few seconds
   kubectl scale deployment alert-manager-controller-manager --replicas=1 -n alert-manager-system
   ```

3. **Increase resource limits**:
   ```bash
   kubectl edit deployment alert-manager-controller-manager -n alert-manager-system
   # Increase memory limits in the spec
   ```

#### Secret Format Issues

**Symptoms**: The controller pod starts but immediately crashes with errors about not being able to get the API token from the secret.

**Log Error Example**:
```
unable to get wavefront api token from secret
```

**Solution**:
The secret must be named `wavefront-api-token` and contain a key with the same name:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: wavefront-api-token
  namespace: alert-manager-system
type: Opaque
stringData:
  wavefront-api-token: "YOUR_API_TOKEN_HERE"  # Key name must match the secret name
```

**Verification**:
```bash
# Check the secret structure
kubectl get secret wavefront-api-token -n alert-manager-system -o yaml
```

#### Missing ConfigMap

**Symptoms**: The controller crashes with errors about not finding `alert-manager-configmap`.

**Log Error Example**:
```
configmaps "alert-manager-configmap" not found
```

**Solution**:
Create the required ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alert-manager-configmap
  namespace: alert-manager-system
data:
  app.mode: "dev"
  base.url: "YOUR_WAVEFRONT_URL"
  backend.type: "wavefront"
```

**Verification**:
```bash
# Check if the ConfigMap exists
kubectl get configmap alert-manager-configmap -n alert-manager-system
```

#### Permission Issues with Secrets

**Symptoms**: The controller crashes with permission errors when trying to read the secrets.

**Log Error Example**:
```
secrets "wavefront-api-token" is forbidden: User cannot get resource "secrets" in API group
```

**Solution**:
Ensure the controller service account has permissions to read secrets in its namespace:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: alert-manager-secrets-role
  namespace: alert-manager-system
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: alert-manager-secrets-rolebinding
  namespace: alert-manager-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: alert-manager-secrets-role
subjects:
- kind: ServiceAccount
  name: alert-manager-controller-manager
  namespace: alert-manager-system
```

**Apply the fix**:
```bash
kubectl apply -f rbac-fix.yaml
```

#### CRDs Not Installing

**Symptoms**: Custom Resource Definitions are not created after installation.

**Solutions**:
1. Check if CRDs exist:
   ```bash
   kubectl get crd | grep alertmanager.keikoproj.io
   ```

2. Manually install CRDs:
   ```bash
   kubectl apply -f config/crd/bases/
   ```

3. Check for RBAC issues:
   ```bash
   kubectl get events | grep crd
   ```

### Controller Issues

#### Leader Election Issues

**Symptoms**: Multiple controller instances fighting for leadership or frequent leader changes.

**Solutions**:
1. Check leader election events:
   ```bash
   kubectl get events -n alert-manager-system | grep -i leader
   ```

2. Check lease status:
   ```bash
   kubectl get lease -n alert-manager-system
   ```

3. Ensure only one controller replica is running:
   ```bash
   kubectl get deployment alert-manager-controller-manager -n alert-manager-system
   ```

#### Health Probe Failures

**Symptoms**: Controller pod is marked as not ready due to probe failures.

**Solutions**:
1. Check probe settings:
   ```bash
   kubectl describe deployment alert-manager-controller-manager -n alert-manager-system
   ```

2. Check controller logs:
   ```bash
   kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager -c manager
   ```

3. Check if health endpoints are reachable:
   ```bash
   kubectl exec -it <controller-pod> -n alert-manager-system -c manager -- curl localhost:8081/healthz
   ```

### Wavefront Alert Issues

#### Alerts Show Error State

**Symptoms**: WavefrontAlerts are stuck in an Error state.

**Possible Causes and Solutions**:

1. **API token issues**:
   - Check if the `wavefront-api-token` secret exists and has the correct format
   - Ensure the token has permissions to create/manage alerts

2. **Network connectivity**:
   - Ensure the controller can reach the Wavefront API endpoint
   - Check for network policies that might be blocking outbound traffic

3. **Incorrect alert configuration**:
   - Verify the alert YAML has valid syntax
   - Check for unsupported alert properties

#### Expected Errors in Test Environments

**Symptoms**: Alerts show an Error state with message like "Post 'https:///api/v2/alert': http: no Host in request URL"

**Explanation**: This is expected behavior when using a non-existent Wavefront URL (like example.com). In a test environment without a real Wavefront instance, alerts will show error states as the controller attempts to create them in Wavefront.

**Solution**: In production, use a valid Wavefront URL and API token. For testing purposes, these errors can be safely ignored.

#### Alerts Not Triggering

**Symptoms**: Alerts are not triggering as expected.

**Possible Causes and Solutions**:

1. **Alert configuration issues**:
   - Verify the alert YAML has valid syntax
   - Check for unsupported alert properties

2. **Wavefront API issues**:
   - Check the Wavefront API for errors
   - Verify the Wavefront API token is correct

3. **Controller issues**:
   - Check the controller logs for errors
   - Verify the controller is running and healthy

### AlertsConfig Issues

#### AlertsConfig Not Processing

**Symptoms**: AlertsConfig resources are created but no alerts are generated.

**Solutions**:
1. Check the AlertsConfig status:
   ```bash
   kubectl get alertsconfig <name> -o yaml
   ```

2. Check the controller logs for errors:
   ```bash
   kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager -c manager | grep -i alertsconfig
   ```

3. Ensure the AlertsConfig resource has proper configuration:
   ```yaml
   apiVersion: alertmanager.keikoproj.io/v1alpha1
   kind: AlertsConfig
   metadata:
     name: example-alerts-config
   spec:
     globalGVK:
       group: alertmanager.keikoproj.io
       kind: WavefrontAlert
       version: v1alpha1
     globalParams:
       # Global params here
     alerts:
       # Alert definitions here
   ```

#### Templating Issues

**Symptoms**: Templated placeholders in AlertsConfig are not being replaced.

**Solutions**:
1. Check the template syntax:
   ```yaml
   alertName: "CPU Alert for {{.Namespace}}"  # Correct syntax
   ```

2. Ensure required template variables are available in the context.

3. Check for quoting issues in the YAML:
   ```yaml
   # Correct
   alertName: "CPU Alert for {{.Namespace}}"
   
   # Incorrect - quotes might interfere with templating
   alertName: '{{.Namespace}} CPU Alert'
   ```

### General Kubernetes Issues

#### Namespace Issues

**Symptoms**: Resources can't be found or accessed.

**Solutions**:
1. Ensure resources are in the correct namespace:
   ```bash
   kubectl get all -n alert-manager-system
   kubectl get wavefrontalerts --all-namespaces
   ```

2. Check namespace exists:
   ```bash
   kubectl get namespace alert-manager-system
   ```

#### Multiple Clusters in Context

**Symptoms**: Commands affect the wrong cluster.

**Solutions**:
1. Check current context:
   ```bash
   kubectl config current-context
   ```

2. List all contexts:
   ```bash
   kubectl config get-contexts
   ```

3. Switch to correct context:
   ```bash
   kubectl config use-context <context-name>
   ```

## Advanced Debugging Techniques

When basic troubleshooting doesn't resolve the issue, these advanced debugging techniques can help diagnose more complex problems.

### Enabling Debug Logs

For detailed troubleshooting, you can increase the log verbosity in the alert-manager controller:

#### Temporarily Enable Debug Logs

```bash
# Edit the controller deployment
kubectl edit deployment alert-manager-controller-manager -n alert-manager-system

# Add or modify the command args in the manager container to include verbose logging:
# - name: manager
#   args:
#   - --zap-log-level=debug  # Add this line
```

Or use a one-liner:

```bash
kubectl patch deployment alert-manager-controller-manager -n alert-manager-system --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/1/args/-", "value": "--zap-log-level=debug"}]'
```

#### Permanently Enable Debug Logs

To persistently enable debug logs, you can modify the alert-manager.yaml before installation:

```bash
# In hack/alert-manager.yaml, add the debug flag to the args section:
containers:
- args:
  - --health-probe-bind-address=:8081
  - --metrics-bind-address=127.0.0.1:8080
  - --leader-elect
  - --zap-log-level=debug  # Add this line
```

#### Available Log Levels

You can set different log levels based on your debugging needs:

- `debug`: Most verbose, shows all detailed information
- `info`: Default level, shows important operational information
- `warn`: Only shows warning and error messages
- `error`: Only shows error messages

### Common Debug Scenarios

#### Debugging Controller Startup Issues

If the controller is failing to start properly:

```bash
# Check the controller logs
kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager -c manager

# Check events related to the controller
kubectl get events -n alert-manager-system | grep controller
```

#### Debugging Alert Processing

To debug alert processing issues:

```bash
# Enable debug logs as described above

# Look for log entries with the alert name
kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager -c manager | grep "my-alert-name"

# Or filter by controller name
kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager -c manager | grep "wavefrontalert_controller"
```

#### Debugging API Connectivity

To debug Wavefront API connectivity issues:

```bash
# Check if the controller can reach the Wavefront API
kubectl exec -it -n alert-manager-system deploy/alert-manager-controller-manager -c manager -- curl -I <your-wavefront-url>/api/v2/alert

# Verify the API token secret is correctly mounted
kubectl exec -it -n alert-manager-system deploy/alert-manager-controller-manager -c manager -- ls -la /var/run/secrets/kubernetes.io/serviceaccount/
```

### Debugging Alert Creation

#### Validating Alert Templates

Use the provided validation script to check alert YAML files:

```bash
# Validate a single alert file
./hack/validate-alert.sh path/to/your-alert.yaml

# Validate multiple files
for file in examples/*.yaml; do
  ./hack/validate-alert.sh $file
done
```

#### Tracing Alert Creation Process

When debug logs are enabled, you can trace the complete lifecycle of an alert:

1. Create an alert with a unique name:
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: alertmanager.keikoproj.io/v1alpha1
   kind: WavefrontAlert
   metadata:
     name: debug-test-alert
     namespace: default
   spec:
     alertName: "Debug Test Alert"
     alertType: CLASSIC
     condition: "ts(test.metric) > 0"
     displayExpression: "ts(test.metric)"
     minutes: 5
     resolveAfterMinutes: 5
     severity: severe
     tags:
       - "debug"
       - "test"
   EOF
   ```

2. Observe the logs for this specific alert:
   ```bash
   kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager -c manager | grep "debug-test-alert"
   ```

3. Check the alert status:
   ```bash
   kubectl get wavefrontalert debug-test-alert -o yaml
   ```

## Collecting Information for Support

When requesting support, collect the following information:

### Controller Logs and Configuration

```bash
# Get controller logs
kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager -c manager > controller-logs.txt

# Get controller configuration
kubectl get configmap -n alert-manager-system alert-manager-configmap -o yaml > configmap.yaml
kubectl get configmap -n alert-manager-system alert-manager-controller-manager-config -o yaml > controller-configmap.yaml
```

### Alert Resources

```bash
# Get all alert resources
kubectl get wavefrontalerts -A -o yaml > wavefrontalerts.yaml
kubectl get alertsconfigs -A -o yaml > alertsconfigs.yaml
```

### Kubernetes Environment Information

```bash
# Get Kubernetes version
kubectl version > k8s-version.txt

# Get node information
kubectl get nodes -o wide > nodes.txt

# Get events
kubectl get events -n alert-manager-system > events.txt
```

### System Information

```bash
# Get system information
kubectl top nodes > node-resources.txt
kubectl top pods -n alert-manager-system > pod-resources.txt
```
