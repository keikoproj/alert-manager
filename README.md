# alert-manager

Generic alerts management inside k8s cluster using CRD. The idea is to allow users to create monitoring alerts along with application deployment.

Following monitoring solutions are supported at the moment and will be adding support for other monitoring solutions depends on the requirement

1. Wavefront
2. Splunk (Phase 2)

### High Level Architecture

![Alert Manager High Architecture](docs/images/alert-manager-arch.png)

Modules involved:

1. WavefrontAlert CRD (Specific to monitoring software)
2. AlertsConfig CRD (Generic for all the alerts)
3. Alert-Manager Controller

##### WavefrontAlert CRD
Monitoring software specific CRDs. For example, WavefrontAlert CRD represents all the fields which are needed to create
alerts in wavefront.


##### AlertsConfig CRD
Above WavefrontAlert CRD can be used for small scale alert setup but if you have a use case where same wavefront alert type needs to be
created for multiple applications (for example: api request count) with slight changes, we might be looking at large number of CRs which might
cause space issues in etcd and also other latency issues. For example, 100 alerts for 450 applications/clusters could result in 45000 CRs
if we want to maintain 1:1 relationship for an alert.

To avoid this problem, introducing new CRD “AlertsConfig” which can represent multiple alert configurations in a single CR

AlertsConfig CRD represents the alert configuration and can be used to represent multiple alerts in a single CR. 
This can be used per application/cluster to represent all the alerts specific to that cluster. 
This changes the total number of CRs to represent cluster alert management to 100 + 450 = 550 in above example.

Simplest usage is, we parameterize application name using go template in WavefrontAlert and pass that value in AlertsConfig CR. 
For more specific and complex usages, please check here

##### Alert-Manager Controller
Alert Controller will handle reconcile of alert and alert config CRs and manage the alerts directly in the target system(ex: Wavefront).

Alert and AlertsConfig association
Alert CR (for ex: WavefrontAlert) status will include AlertsConfig CR name along with the target name which could help represent the targets to which that alert has been applied to.

Alert controller will update the Alert CR status as soon as there is a change in AlertsConfig and also Alert controller will take care of applying the modified alert if there is a change in the Alert Spec.


### Usage

Sample Wavefront alert

```yaml
apiVersion: alertmanager.keikoproj.io/v1alpha1
kind: WavefrontAlert
metadata:
  name: wavefrontalert-sample
spec:
  # Add fields here
  alertType: CLASSIC
  alertName: test-alert2
  condition: ts(status.health)
  displayExpression: ts(status.health)
  minutes: 50
  resolveAfterMinutes: 5
  severity: severe
  tags:
    - test-alert
    - something-weird
```

and you should see the wavefront alert info in the status if the alert got successfully created

```yaml
apiVersion: alertmanager.keikoproj.io/v1alpha1
kind: WavefrontAlert
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"alertmanager.keikoproj.io/v1alpha1","kind":"WavefrontAlert","metadata":{"annotations":{},"name":"wavefrontalert-sample","namespace":"alert-manager-system"},"spec":{"alertName":"test-alert2","alertType":"CLASSIC","condition":"ts(status.health)","displayExpression":"ts(status.health)","minutes":50,"resolveAfterMinutes":5,"severity":"severe","tags":["test-alert","something-weird"]}}
  creationTimestamp: "2021-08-26T19:20:27Z"
  finalizers:
  - wavefrontalert.finalizers.alertmanager.keikoproj.io
  generation: 8
  name: wavefrontalert-sample
  namespace: alert-manager-system
  resourceVersion: "293921"
  uid: 7d108d5c-30a6-4f80-8acc-cdb6ef5fb785
spec:
  alertName: test-alert2
  alertType: CLASSIC
  condition: ts(status.health)
  displayExpression: ts(status.health)
  minutes: 50
  resolveAfterMinutes: 5
  severity: severe
  tags:
  - test-alert
  - something-weird
status:
  alertsStatus:
    test-alert2:
      alertName: test-alert2
      associatedAlert: {}
      associatedAlertsConfig: {}
      id: "1630005627450"
      lastChangeChecksum: 3a86ae56b46c66d51cf270dde6c469b7
      link: https://try.wavefront.com/alerts/1630005627450
      state: Ready
  lastChangeChecksum: 3a86ae56b46c66d51cf270dde6c469b7
  observedGeneration: 8
  state: Ready
```
For more template alerts user case and alerts config usage please refer here

### Installation
1. Update the config map with the target urls. For example: For WavefrontAlerts, it is pointed to try.wavefront.com and you must change it to your target url
2. Create a Secret with wavefront Api Key (for WavefrontAlerts). Please refer Sample Secret [here](docs/Sample-Secret.yaml)
   - **Important**: You must replace the placeholder value in the Sample-Secret.yaml with your actual base64 encoded Wavefront API token
3. Point your kubeconfig file to the k8s cluster where you want to create. ex: export KUBECONFIG=~/.kube/config
4. make deploy

### ❤ Contributing ❤

Please see [CONTRIBUTING.md](.github/CONTRIBUTING.md).

### Developer Guide

Please see [DEVELOPER.md](.github/DEVELOPER.md).
