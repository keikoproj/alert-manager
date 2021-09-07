/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	internalconfig "github.com/keikoproj/alert-manager/internal/config"
	"github.com/keikoproj/alert-manager/internal/utils"
	"github.com/keikoproj/alert-manager/pkg/log"
	"github.com/keikoproj/alert-manager/pkg/wavefront"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	wf "github.com/WavefrontHQ/go-wavefront-management-api"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	controllercommon "github.com/keikoproj/alert-manager/controllers/common"
)

const (
	alertsConfigFinalizerName = "alertsconfig.finalizers.alertmanager.keikoproj.io"
)

// AlertsConfigReconciler reconciles a AlertsConfig object
type AlertsConfigReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	CommonClient    *controllercommon.Client
	WavefrontClient wavefront.Interface
}

//+kubebuilder:rbac:groups=alertmanager.keikoproj.io,resources=alertsconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=alertmanager.keikoproj.io,resources=alertsconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=alertmanager.keikoproj.io,resources=alertsconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AlertsConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *AlertsConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("alertsconfig", req.NamespacedName)

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	ctx = context.WithValue(ctx, requestId, uuid.New())
	log := log.Logger(ctx, "controllers", "alertconfig_controller", "Reconcile")
	log = log.WithValues("alertconfig_cr", req.NamespacedName)
	log.Info("Start of the request")

	// Get the CR
	var alertsConfig alertmanagerv1alpha1.AlertsConfig
	if err := r.Get(ctx, req.NamespacedName, &alertsConfig); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if it is delete request
	if !alertsConfig.ObjectMeta.DeletionTimestamp.IsZero() {
		requeueFlag := false
		// Delete use case
		if err := r.HandleDelete(ctx, &alertsConfig); err != nil {
			log.Error(err, "unable to delete the alert")
			requeueFlag = true
		}
		return ctrl.Result{Requeue: requeueFlag}, nil
	}

	//First time use case
	if !utils.ContainsString(alertsConfig.ObjectMeta.Finalizers, alertsConfigFinalizerName) {
		log.Info("New alerts config resource. Adding the finalizer", "finalizer", alertsConfigFinalizerName)

		alertsConfig.ObjectMeta.Finalizers = append(alertsConfig.ObjectMeta.Finalizers, alertsConfigFinalizerName)
		r.CommonClient.UpdateMeta(ctx, &alertsConfig)
		//That's fine- Let it come for requeue and we can create the alert
		return ctrl.Result{}, nil
	}

	alertHashMap := alertsConfig.Status.AlertsStatus

	// Handle create/update here
	for alertName, config := range alertsConfig.Spec.Alerts {
		// Calculate checksum and compare it with the status checksum
		exist, reqChecksum := utils.CalculateAlertConfigChecksum(ctx, config)
		// if request and status checksum matches then there is NO change in this specific alert config
		if exist && alertHashMap[alertName].LastChangeChecksum == reqChecksum {
			log.V(1).Info("checksum is equal so there is no change. skipping", "alertName", alertName)
			//skip it
			continue
		}
		// if there is a diff
		// Get Alert CR

		var wfAlert alertmanagerv1alpha1.WavefrontAlert
		wfAlertNamespacedName := types.NamespacedName{Namespace: req.Namespace, Name: alertName}
		if err := r.Get(ctx, wfAlertNamespacedName, &wfAlert); err != nil {
			log.Error(err, "unable to get the wavefront alert details for the requested name", "wfAlertName", alertName)
			// This means wavefront alert itself is not created.
			// There could be 2 use cases
			// 1. There was a race condition if wavefrontalert and alerts config got created 'almost at the same time'
			// 2. Wrong alert name and user is going to correct
			// Ideal way to handle this is to make the alert config status to error and requeue it once in 5 mins or so instead of standard kube builder requeue time
			// Update the status and retry it
			return r.PatchIndividualAlertsConfigError(ctx, &alertsConfig, alertName, alertmanagerv1alpha1.Error, err)
		}
		var alert wf.Alert
		//Get the processed wf alert
		if err := controllercommon.GetProcessedWFAlert(ctx, &wfAlert, &config, &alert); err != nil {
			return r.PatchIndividualAlertsConfigError(ctx, &alertsConfig, alertName, alertmanagerv1alpha1.Error, err)
		}
		// Create/Update Alert
		if alertHashMap[alertName].ID == "" {
			// Create use case
			if err := r.WavefrontClient.CreateAlert(ctx, &alert); err != nil {
				r.Recorder.Event(&alertsConfig, v1.EventTypeWarning, err.Error(), "unable to create the alert")
				state := alertmanagerv1alpha1.Error
				if strings.Contains(err.Error(), "Exceeded limit setting") {
					// For ex: error is "Exceeded limit setting: 100 alerts allowed per customer"
					state = alertmanagerv1alpha1.ClientExceededLimit
				}
				log.Error(err, "unable to create the alert")

				return r.PatchIndividualAlertsConfigError(ctx, &alertsConfig, alertName, state, err)
			}
			alertStatus := alertmanagerv1alpha1.AlertStatus{
				ID:                 *alert.ID,
				Name:               alert.Name,
				LastChangeChecksum: reqChecksum,
				Link:               fmt.Sprintf("https://%s/alerts/%s", internalconfig.Props.WavefrontAPIUrl(), *alert.ID),
				State:              alertmanagerv1alpha1.Ready,
				AssociatedAlert: alertmanagerv1alpha1.AssociatedAlert{
					CR: alertName,
				},
				AssociatedAlertsConfig: alertmanagerv1alpha1.AssociatedAlertsConfig{
					CR: alertsConfig.Name,
				},
			}
			if err := r.CommonClient.PatchWfAlertAndAlertsConfigStatus(ctx, &wfAlert, &alertsConfig, alertStatus); err != nil {
				log.Error(err, "unable to patch wfalert and alertsconfig status objects")
				return r.PatchIndividualAlertsConfigError(ctx, &alertsConfig, alertName, alertmanagerv1alpha1.Error, err)
			}
			log.Info("alert successfully got created", "alertID", alert.ID)

		} else {
			alertID := alertHashMap[alertName].ID
			alert.ID = &alertID
			//TODO: Move this to common so it can be used for both wavefront and alerts config
			//Update use case
			// This can be changed to a common function that used by alertconfig and wavefrontalerts, because it is
			// updating the wavefront alert
			if err := r.WavefrontClient.UpdateAlert(ctx, &alert); err != nil {
				r.Recorder.Event(&alertsConfig, v1.EventTypeWarning, err.Error(), "unable to update the alert")
				state := alertmanagerv1alpha1.Error
				if strings.Contains(err.Error(), "Exceeded limit setting") {
					// For ex: error is "Exceeded limit setting: 100 alerts allowed per customer"
					state = alertmanagerv1alpha1.ClientExceededLimit
				}
				log.Error(err, "unable to create the alert")

				return r.PatchIndividualAlertsConfigError(ctx, &alertsConfig, alertName, state, err)
			}

			alertStatus := alertHashMap[alertName]
			alertStatus.LastChangeChecksum = reqChecksum

			if err := r.CommonClient.PatchWfAlertAndAlertsConfigStatus(ctx, &wfAlert, &alertsConfig, alertStatus); err != nil {
				log.Error(err, "unable to patch wfalert and alertsconfig status objects")
				return r.PatchIndividualAlertsConfigError(ctx, &alertsConfig, alertName, alertmanagerv1alpha1.Error, err)
			}
			log.Info("alert successfully got updated", "alertID", alert.ID)
		}
	}

	// Now - lets see if there is any config is removed compared to the status
	// If there is any, we need to make a call to delete the alert
	return r.HandleIndividalAlertConfigRemoval(ctx, req.NamespacedName)
}

//HandleIndividalAlertConfigRemoval function handles if there is any config got removed from the spec, if so- delete that alert in wavefront and also update the status
func (r *AlertsConfigReconciler) HandleIndividalAlertConfigRemoval(ctx context.Context, namespacedName types.NamespacedName) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "alertsconfig_controller", "HandleIndividalAlertConfigRemoval")
	log = log.WithValues("alertsConfig_cr", namespacedName)
	// Get the alerts config again

	var updatedAlertsConfig alertmanagerv1alpha1.AlertsConfig
	if err := r.Get(ctx, namespacedName, &updatedAlertsConfig); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	tempStatusConfig := updatedAlertsConfig.Status.AlertsStatus
	tempState := updatedAlertsConfig.Status.State
	var toBeDeleted []string

	for key, status := range updatedAlertsConfig.Status.AlertsStatus {
		// This is for sure delete use case
		if _, ok := updatedAlertsConfig.Spec.Alerts[key]; !ok {
			//This means we didn't find this in spec anymore
			// Lets delete that then
			if err := r.DeleteIndividualAlert(ctx, updatedAlertsConfig.Name, status, updatedAlertsConfig.Namespace); err != nil {
				//Ignore if errors since we can consider it as already deleted
				log.Error(err, "unable to delete the alert, assuming alerts doesn't exist anymore- proceeding further")
			}
			toBeDeleted = append(toBeDeleted, key)
		} else {
			if status.State == alertmanagerv1alpha1.Error {
				tempState = alertmanagerv1alpha1.Error
			}
		}
	}

	for _, key := range toBeDeleted {
		delete(tempStatusConfig, key)
	}
	// update the count
	updatedAlertsConfig.Status.AlertsCount = len(updatedAlertsConfig.Spec.Alerts)
	updatedAlertsConfig.Status.AlertsStatus = tempStatusConfig
	// update the status
	return r.CommonClient.UpdateStatus(ctx, &updatedAlertsConfig, tempState, errRequeueTime)
}

//DeleteIndividualAlert function deletes individual alert and also patches the status on both wavefront alert and also alerts config status
func (r *AlertsConfigReconciler) DeleteIndividualAlert(ctx context.Context, alertName string, alertStatus alertmanagerv1alpha1.AlertStatus, namespace string) error {
	log := log.Logger(ctx, "controllers", "alertsconfig_controller", "DeleteIndividualAlert")
	log = log.WithValues("alertsConfig_cr", alertName)

	if alertStatus.ID != "" {
		if err := r.WavefrontClient.DeleteAlert(ctx, alertStatus.ID); err != nil {
			log.Error(err, "skipping alert deletion", "alertID", alertStatus.ID)
			// Just skip it for now
			// this is too opinionated but we don't want to stop the delete execution for other alerts as well
			// if there is any valid reasons not to skip it, we can look into it in future
		}
	}

	// Update the wavefront alert status

	var wfAlert alertmanagerv1alpha1.WavefrontAlert
	wfAlertNamespacedName := types.NamespacedName{Namespace: namespace, Name: alertStatus.AssociatedAlert.CR}
	if err := r.Get(ctx, wfAlertNamespacedName, &wfAlert); err != nil {
		log.Error(err, "unable to get the wavefront alert details for the requested name", "wfAlertName", alertName)
		// This means wavefront alert itself is not there so we can ignore.
		return nil
	}

	// Wavefront alert is present - lets update the status

	currStatus := wfAlert.Status.AlertsStatus
	delete(currStatus, alertName)
	wfAlert.Status.AlertsStatus = currStatus
	r.CommonClient.UpdateStatus(ctx, &wfAlert, wfAlert.Status.State)
	return nil
}

//PatchIndividualAlertsConfigError function is a utility function to patch the error status
// We use status patch instead of status update to avoid any overwrite between two threads when alertsConfig CR has multiple alert configs
func (r *AlertsConfigReconciler) PatchIndividualAlertsConfigError(ctx context.Context, alertsConfig *alertmanagerv1alpha1.AlertsConfig, alertName string, state alertmanagerv1alpha1.State, err error, requeueTime ...float64) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "alertsconfig_controller", "PatchIndividualAlertsConfigError")
	log = log.WithValues("alertsConfig_cr", alertsConfig.Name, "namespace", alertsConfig.Namespace)
	alertStatus := alertsConfig.Status.AlertsStatus[alertName]
	alertStatus.State = state
	alertStatusBytes, _ := json.Marshal(alertStatus)
	retryCount := alertsConfig.Status.RetryCount + 1
	log.Error(err, "error occured in alerts config for alert name", "alertName", alertName)
	r.Recorder.Event(alertsConfig, v1.EventTypeWarning, err.Error(), fmt.Sprintf("error occured in alerts config for alert name %s", alertName))

	patch := []byte(fmt.Sprintf("{\"status\":{\"state\": \"%s\", \"alertsCount\": %d, \"retryCount\": %d, \"alertsStatus\":{\"%s\":%s}}}", state, alertsConfig.Status.AlertsCount, retryCount, alertName, string(alertStatusBytes)))
	return r.CommonClient.PatchStatus(ctx, alertsConfig, client.RawPatch(types.MergePatchType, patch), alertmanagerv1alpha1.Error, errRequeueTime)
}

//HandleDelete function handles the deleting wavefront alerts
func (r *AlertsConfigReconciler) HandleDelete(ctx context.Context, alertsConfig *alertmanagerv1alpha1.AlertsConfig) error {
	log := log.Logger(ctx, "controllers", "alertsconfig_controller", "HandleDelete")
	log = log.WithValues("alertsConfig_cr", alertsConfig.Name, "namespace", alertsConfig.Namespace)
	// Lets check the status of the CR and
	// retrieve all the alerts associated with this CR and delete it
	//Check if any alerts were created with this config
	if len(alertsConfig.Status.AlertsStatus) > 0 {
		//Call wavefront api and delete the alerts one by one
		for _, alert := range alertsConfig.Status.AlertsStatus {
			if err := r.DeleteIndividualAlert(ctx, alertsConfig.Name, alert, alertsConfig.Namespace); err != nil {
				log.Error(err, "skipping alert deletion", "alertID", alert.ID)
				// Just skip it for now
				// this is too opinionated but we don't want to stop the delete execution for other alerts as well
				// if there is any valid reasons not to skip it, we can look into it in future
			}
		}
	}

	// Ok. Lets delete the finalizer so controller can delete the custom object
	log.Info("Removing finalizer from WavefrontAlert")
	alertsConfig.ObjectMeta.Finalizers = utils.RemoveString(alertsConfig.ObjectMeta.Finalizers, alertsConfigFinalizerName)
	r.CommonClient.UpdateMeta(ctx, alertsConfig)
	log.Info("Successfully deleted wfAlert")
	r.Recorder.Event(alertsConfig, v1.EventTypeNormal, "Deleted", "Successfully deleted WavefrontAlert")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlertsConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alertmanagerv1alpha1.AlertsConfig{}).
		WithEventFilter(controllercommon.StatusUpdatePredicate{}).
		Complete(r)
}
