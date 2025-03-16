# Alert Manager Troubleshooting Guide

This guide provides solutions for common issues you might encounter when using alert-manager.

## Table of Contents
- [Installation Issues](#installation-issues)
- [Controller Issues](#controller-issues)
- [Wavefront Alert Issues](#wavefront-alert-issues)
- [AlertsConfig Issues](#alertsconfig-issues)
- [General Kubernetes Issues](#general-kubernetes-issues)
- [Collecting Information for Bug Reports](#collecting-information-for-bug-reports)

## Installation Issues

### Controller Pod Not Starting

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

### CRDs Not Installing

**Symptoms**: Custom Resource Definitions are not created after installation.

**Solution**:
```bash
# Install CRDs manually
kubectl apply -f config/crd/bases/
```

## Controller Issues

### Controller Not Reconciling Resources

**Symptoms**: Custom resources are created but the controller doesn't appear to be processing them.

**Possible Causes and Solutions**:

1. **Controller not watching the namespace**:
   - Ensure you're creating resources in a namespace the controller is configured to watch
   - Check if namespace selector is configured in the controller

2. **Controller crashes or errors**:
   ```bash
   kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager
   ```
   Look for error messages that might indicate why reconciliation is failing.

### High CPU/Memory Usage

**Symptoms**: The controller pod uses excessive CPU or memory.

**Solutions**:
- Reduce the number of concurrent reconciles in the ConfigMap
- Check for reconciliation loops where the same resource is constantly being processed
- Consider implementing pagination if handling large numbers of alerts

## Wavefront Alert Issues

### Alerts Not Appearing in Wavefront

**Symptoms**: Custom resources are created but no alerts appear in Wavefront.

**Possible Causes and Solutions**:

1. **API credentials issue**:
   - Verify that the Wavefront API token is correct
   - Check the secret containing the token:
     ```bash
     kubectl get secret -n alert-manager-system wavefront-api-key -o yaml
     ```

2. **Incorrect Wavefront URL**:
   - Check the ConfigMap for the correct Wavefront URL:
     ```bash
     kubectl get configmap -n alert-manager-system alert-manager-config -o yaml
     ```
   - Ensure the URL includes the protocol (https://) and has no trailing slash

3. **Alert validation errors**:
   - Check the controller logs for validation errors:
     ```bash
     kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager
     ```
   - Verify that all required fields in the WavefrontAlert CR are present and valid

### Alerts Not Updating in Wavefront

**Symptoms**: Changes to alert custom resources are not reflected in Wavefront.

**Solutions**:
- Check if the resource's status shows the alert as "Ready"
- Verify that the controller processed the update (check logs)
- Try adding a small change to force reconciliation
- Check if status.lastChangeChecksum is being updated

## AlertsConfig Issues

### Template Processing Errors

**Symptoms**: AlertsConfig resources show errors related to template processing.

**Possible Causes and Solutions**:

1. **Missing template parameters**:
   - Ensure all parameters referenced in the alert template are provided in the AlertsConfig
   - Check for typos in parameter names

2. **Template not found**:
   - Verify that the referenced alert template exists:
     ```bash
     kubectl get wavefrontalert -A
     ```
   - Check namespace selectors if templates are in different namespaces

3. **Template syntax errors**:
   - Validate Go template syntax
   - Check controller logs for specific syntax errors

## General Kubernetes Issues

### Custom Resource Deletion Hangs

**Symptoms**: Attempts to delete a custom resource hang indefinitely.

**Solution**:
```bash
# Remove finalizers from the resource
kubectl patch wavefrontalert <name> -n <namespace> --type json -p '[{"op":"remove","path":"/metadata/finalizers"}]'
```

### Multiple Identical Alerts

**Symptoms**: The same alert appears multiple times in Wavefront with different IDs.

**Solution**:
- Check for duplicate custom resources
- Verify the alert names are unique
- Check if multiple controller instances are running

## Collecting Information for Bug Reports

When filing a bug report, please include:

1. **Controller logs**:
   ```bash
   kubectl logs -n alert-manager-system deployment/alert-manager-controller-manager --tail=200
   ```

2. **Custom resource definitions**:
   ```bash
   kubectl get crds | grep alertmanager.keikoproj.io
   ```

3. **Custom resource samples**:
   ```bash
   kubectl get wavefrontalert -A -o yaml
   kubectl get alertsconfig -A -o yaml
   ```

4. **Kubernetes version**:
   ```bash
   kubectl version
   ```

5. **Alert manager version**:
   ```bash
   # Get image tag
   kubectl get deployment -n alert-manager-system alert-manager-controller-manager -o jsonpath='{.spec.template.spec.containers[0].image}'
   ```

6. **Reproduction steps**:
   - Clear steps to reproduce the issue
   - Sample resources that demonstrate the problem
