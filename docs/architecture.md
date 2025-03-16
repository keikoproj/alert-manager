# Alert Manager Architecture

This document describes the architecture of alert-manager, a Kubernetes operator that enables management of monitoring alerts as custom resources within Kubernetes clusters.

## Overview

alert-manager follows the Kubernetes operator pattern, watching for custom resources that define alerts and reconciling them with monitoring systems like Wavefront. It enables a GitOps approach to alert management, allowing alerts to be version-controlled and deployed alongside applications.

## Architecture Diagram

```mermaid
graph TD
    %% Define styles
    classDef k8s fill:#326ce5,color:white,stroke:white,stroke-width:2px
    classDef wavefront fill:#00ACEE,color:white,stroke:white,stroke-width:2px
    classDef controller fill:#7D559C,color:white,stroke:white,stroke-width:2px
    classDef user fill:#767676,color:white,stroke:white,stroke-width:2px
    
    %% User interactions
    User([DevOps/User]):::user
    User -->|Creates/Updates| WavefrontAlert[WavefrontAlert CR]:::k8s
    User -->|Creates/Updates| AlertsConfig[AlertsConfig CR]:::k8s
    
    %% Kubernetes components
    subgraph Kubernetes Cluster
        WavefrontAlert
        AlertsConfig
        APIServer[Kubernetes API Server]:::k8s
        ControllerManager[Alert Manager Controller]:::controller
        Secret[Credentials Secret]:::k8s
        ConfigMap[AlertManager ConfigMap]:::k8s
        
        WavefrontAlert -->|Submitted to| APIServer
        AlertsConfig -->|Submitted to| APIServer
        APIServer -->|Watched by| ControllerManager
        ControllerManager -->|Updates status| APIServer
        ConfigMap -->|Configuration| ControllerManager
        Secret -->|API Credentials| ControllerManager
    end
    
    %% Monitoring systems
    subgraph Monitoring Systems
        Wavefront[Wavefront]:::wavefront
        Splunk[Splunk]:::wavefront
        
        ControllerManager -->|Creates/Updates Alerts| Wavefront
        ControllerManager -->|Creates/Updates Alerts| Splunk
        Wavefront -->|Alert Status| ControllerManager
    end
    
    %% CR relationships
    AlertsConfig -.->|References| WavefrontAlert
    WavefrontAlert -.->|Status includes| AlertsConfig
```

## Component Interactions

### WavefrontAlert Flow

```mermaid
sequenceDiagram
    actor User
    participant WavefrontAlert as WavefrontAlert CR
    participant Controller as Alert Manager Controller
    participant Wavefront as Wavefront API
    
    User->>WavefrontAlert: Create/Update CR
    WavefrontAlert->>Controller: Notify of change
    Controller->>WavefrontAlert: Process alert definition
    Controller->>Wavefront: Create/Update alert
    Wavefront-->>Controller: Return alert ID & status
    Controller->>WavefrontAlert: Update status with alert ID & link
    
    Note over Controller,Wavefront: Periodic reconciliation
    Controller->>Wavefront: Check alert status
    Wavefront-->>Controller: Return current status
    Controller->>WavefrontAlert: Update CR status
```

### AlertsConfig Flow

```mermaid
sequenceDiagram
    actor User
    participant AlertsConfig as AlertsConfig CR
    participant WavefrontAlert as WavefrontAlert Templates
    participant Controller as Alert Manager Controller
    participant Wavefront as Wavefront API
    
    User->>AlertsConfig: Create/Update CR with parameters
    AlertsConfig->>Controller: Notify of change
    Controller->>WavefrontAlert: Get alert templates
    Controller->>Controller: Process templates with parameters
    
    loop For each alert in config
        Controller->>Wavefront: Create/Update alert
        Wavefront-->>Controller: Return alert ID & status
    end
    
    Controller->>AlertsConfig: Update status with alerts info
    Controller->>WavefrontAlert: Update referenced templates status
```

## Key Components

### 1. Custom Resource Definitions (CRDs)

#### WavefrontAlert CRD
Defines a specific alert in Wavefront with:
- Alert name and conditions
- Notification targets
- Severity
- Display expressions
- Tags

#### AlertsConfig CRD
Allows efficient management of multiple similar alerts by:
- Referencing alert templates (WavefrontAlert CRs)
- Providing parameters to customize the templates
- Enabling/disabling specific alerts
- Overriding default template values

### 2. Alert Manager Controller

The controller:
- Watches for changes to alert-related CRs
- Reconciles the desired state (CRs) with the actual state (monitoring systems)
- Manages the lifecycle of alerts in monitoring systems
- Updates CR status with current alert information
- Handles error conditions and retries

### 3. Monitoring System Integrations

Currently supports:
- **Wavefront**: Complete implementation
- **Splunk**: Planned for future release

## Scalability Design

The AlertsConfig approach addresses a key scalability concern:

Traditional approach: 1 alert type × 100 applications = 100 CRs  
Alert-manager approach: 1 AlertsConfig per application = 1 CR

For large environments (e.g., 450 applications with 100 alert types):
- Traditional approach: 45,000 CRs (risk of etcd overload)
- Alert-manager approach: 450 CRs (manageable)

## Security Considerations

- Controller requires API credentials for monitoring systems
- Credentials stored as Kubernetes Secrets
- RBAC controls who can create/modify alert resources
- Namespace-scoped resources allow isolation between teams

## Configuration

The alert-manager is configured through a ConfigMap which defines:
- Monitoring system endpoints
- Default settings
- Retry parameters
- Logging levels
