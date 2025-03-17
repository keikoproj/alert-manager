# Alert Manager Security Guide

This guide provides security best practices for deploying and using alert-manager in production environments.

## Table of Contents
- [Securing Alert Manager Installation](#securing-alert-manager-installation)
- [Protecting API Credentials](#protecting-api-credentials)
- [Limiting Access to Alert Resources](#limiting-access-to-alert-resources)
- [Namespace Isolation](#namespace-isolation)
- [Securing Alert Definitions](#securing-alert-definitions)
- [Monitoring and Auditing](#monitoring-and-auditing)

## Securing Alert Manager Installation

### Minimum Required Permissions

Alert Manager should run with the minimum required permissions to function:

```yaml
# Example RBAC configuration with restricted permissions
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: alert-manager-restricted-role
rules:
- apiGroups: ["alertmanager.keikoproj.io"]
  resources: ["wavefrontalerts", "alertsconfigs"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["alertmanager.keikoproj.io"]
  resources: ["wavefrontalerts/status", "alertsconfigs/status"]
  verbs: ["get", "update", "patch"]
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
  # Consider restricting to specific resources by name
```

### Secure Network Access

- Enable network policies to restrict who can access the alert-manager controller
- Configure alertmanager service to be accessible only within the cluster
- Use TLS for all communication with external systems

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: alert-manager-network-policy
  namespace: alert-manager-system
spec:
  podSelector:
    matchLabels:
      control-plane: controller-manager
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
  egress:
  - to:
    - ipBlock:
        cidr: 0.0.0.0/0
    ports:
    - protocol: TCP
      port: 443  # HTTPS for Wavefront API
```

## Protecting API Credentials

### Secret Management

- Store Wavefront API tokens in Kubernetes Secrets (not ConfigMaps)
- Consider using an external secret manager (AWS Secrets Manager, HashiCorp Vault, etc.)
- Rotate API tokens regularly (every 30-90 days)

```yaml
# Example using external-secrets operator with AWS Secrets Manager
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: wavefront-api-token
  namespace: alert-manager-system
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secretsmanager
    kind: ClusterSecretStore
  target:
    name: wavefront-api-token
  data:
  - secretKey: wavefront-api-token
    remoteRef:
      key: prod/alert-manager/wavefront-api-token
```

### Secure API Token Transmission

- Always use HTTPS when connecting to Wavefront
- Ensure the controller only accesses the token when needed
- Limit which components can read the secret containing the API token

## Limiting Access to Alert Resources

### RBAC for Alerts

Use namespaced roles to limit who can create and manage alerts:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: team-a
  name: alert-creator
rules:
- apiGroups: ["alertmanager.keikoproj.io"]
  resources: ["wavefrontalerts"]
  verbs: ["create", "update", "delete", "get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: team-a-alert-creators
  namespace: team-a
subjects:
- kind: Group
  name: team-a-developers
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: alert-creator
  apiGroup: rbac.authorization.k8s.io
```

### Admission Controls

Consider implementing admission webhooks to validate and enforce policies on alert definitions:

- Limit which metrics can be queried
- Enforce tagging standards 
- Restrict alert notifications to approved channels
- Prevent creation of alerts that could cause alert fatigue

## Namespace Isolation

Organize alerts by namespace to implement logical separation:

- Create separate namespaces for different teams or applications
- Use namespace-level RBAC controls to limit access
- Consider implementing multi-tenancy with separate alert notification targets

## Securing Alert Definitions

### Setting Appropriate Thresholds

- Implement a threshold review process to prevent alert fatigue
- Document and standardize thresholds for common metrics
- Consider implementing a "burn-in" period for new alerts

### Alert Content Security

- Sanitize user input in alert definitions to prevent injection attacks
- Be cautious with alerts that include query results in notifications 
- Use parameterized templates for alerts

## Monitoring and Auditing

### Auditing Alert Changes

- Use audit logging to track changes to alert definitions
- Consider implementing GitOps workflows to track alert changes
- Regularly review alert changes and implement a change control process

### Monitoring Alert Manager Itself

Monitor the health of the alert-manager controller:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: alert-manager-controller
  namespace: monitoring
spec:
  selector:
    matchLabels:
      control-plane: controller-manager
  namespaceSelector:
    matchNames:
    - alert-manager-system
  podMetricsEndpoints:
  - port: metrics
```

## References

- [Kubernetes RBAC Documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Wavefront Security Best Practices](https://docs.wavefront.com/wavefront_security.html)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/overview/)
