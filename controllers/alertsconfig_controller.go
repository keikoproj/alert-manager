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
	"github.com/google/uuid"
	internalconfig "github.com/keikoproj/alert-manager/internal/config"
	"github.com/keikoproj/alert-manager/internal/template"
	"github.com/keikoproj/alert-manager/internal/utils"
	"github.com/keikoproj/alert-manager/pkg/log"
	"github.com/keikoproj/alert-manager/pkg/wavefront"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

	// Convert list of alerts from status into Map for easy access since
	// we need to compare alert hash from status to newly calculated hash to find the difference
	alertHashMap := alertsConfig.Status.Alerts

	// if len(configs) > 0
	for _, config := range alertsConfig.Spec.Alerts {
		// Calculate checksum and compare it with the status checksum
		exist, reqChecksum := utils.CalculateAlertConfigChecksum(ctx, config)
		// if request and status checksum matches then there is change in this specific alert config
		if exist && alertHashMap[config.AlertName].LastChangeChecksum == reqChecksum {
			log.Info("checksum is equal. skipping")
			//skip it
			continue
		}
		// if there is a diff
		// Get Alert CR

		var wfAlert alertmanagerv1alpha1.WavefrontAlert
		wfAlertNamespacedName := types.NamespacedName{Namespace: req.Namespace, Name: config.AlertName}
		if err := r.Get(ctx, wfAlertNamespacedName, &wfAlert); err != nil {
			log.Error(err, "unable to get the wavefront alert details for the requested name", "wfAlertName", config.AlertName)

			// Update the status and retry it
			return r.PatchIndividualAlertsConfigError(ctx, &alertsConfig, config.AlertName, err)
		}
		wfAlertBytes, err := json.Marshal(wfAlert.Spec)
		if err != nil {
			// update the status and retry it
			return r.PatchIndividualAlertsConfigError(ctx, &alertsConfig, config.AlertName, err)
		}

		// execute Golang Template
		wfAlertTemplate, err := template.ProcessTemplate(ctx, string(wfAlertBytes), config.Params)
		if err != nil {
			//update the status and retry it
			return r.PatchIndividualAlertsConfigError(ctx, &alertsConfig, config.AlertName, err)
		}
		log.Info("Template process is successful", "here", wfAlertTemplate)

		// Unmarshal back to wavefront alert
		if err := json.Unmarshal([]byte(wfAlertTemplate), &wfAlert.Spec); err != nil {
			// update the wfAlert status and retry it
			return r.PatchIndividualAlertsConfigError(ctx, &alertsConfig, config.AlertName, err)
		}
		// Convert to Alert
		var alert wf.Alert

		if err := wavefront.ConvertAlertCRToWavefrontRequest(ctx, wfAlert.Spec, &alert); err != nil {
			errMsg := "unable to convert the wavefront spec to Alert API request. will not be retried"
			log.Error(err, errMsg)
			return r.PatchIndividualAlertsConfigError(ctx, &alertsConfig, config.AlertName, err)
		}

		// Create/Update Alert
		if alertHashMap[config.AlertName].ID == "" {
			// Create use case
			if err := r.WavefrontClient.CreateAlert(ctx, &alert); err != nil {
				r.Recorder.Event(&alertsConfig, v1.EventTypeWarning, err.Error(), "unable to create the alert")
				state := alertmanagerv1alpha1.Error
				if strings.Contains(err.Error(), "Exceeded limit setting") {
					// For ex: error is "Exceeded limit setting: 100 alerts allowed per customer"
					state = alertmanagerv1alpha1.Error
				}
				log.Error(err, "unable to create the alert")

				return r.CommonClient.UpdateStatus(ctx, &alertsConfig, state, errRequeueTime)
			}
			alertStatus := alertmanagerv1alpha1.AlertStatus{
				ID:                 *alert.ID,
				Name:               alert.Name,
				LastChangeChecksum: reqChecksum,
				Link:               fmt.Sprintf("https://%s/alerts/%s", internalconfig.Props.WavefrontAPIUrl(), *alert.ID),
				State:              alertmanagerv1alpha1.Ready,
				AssociatedAlert: alertmanagerv1alpha1.AssociatedAlert{
					CR: config.AlertName,
				},
			}

			alertStatusBytes, _ := json.Marshal(alertStatus)
			patch := []byte(fmt.Sprintf("{\"status\":{\"state\": \"%s\", \"alertsCount\": 0, \"retryCount\": 0, \"alertStatus\":{\"%s\":%s}}}", alertmanagerv1alpha1.Ready, config.AlertName, string(alertStatusBytes)))
			r.CommonClient.PatchStatus(ctx, &alertsConfig, client.RawPatch(types.MergePatchType, patch), alertmanagerv1alpha1.Ready)
			log.Info("alert successfully got created", "alertID", alert.ID)

		} else {

			alertID := alertHashMap[config.AlertName].ID
			alert.ID = &alertID
			//Update use case
			if err := r.WavefrontClient.UpdateAlert(ctx, &alert); err != nil {
				r.Recorder.Event(&alertsConfig, v1.EventTypeWarning, err.Error(), "unable to update the alert")
				state := alertmanagerv1alpha1.Error
				if strings.Contains(err.Error(), "Exceeded limit setting") {
					// For ex: error is "Exceeded limit setting: 100 alerts allowed per customer"
					state = alertmanagerv1alpha1.Error
				}
				log.Error(err, "unable to create the alert")

				return r.CommonClient.UpdateStatus(ctx, &alertsConfig, state, errRequeueTime)
			}

			alertStatus := alertHashMap[config.AlertName]
			alertStatus.LastChangeChecksum = reqChecksum

			alertStatusBytes, _ := json.Marshal(alertStatus)
			patch := []byte(fmt.Sprintf("{\"status\":{\"state\": \"%s\", \"alertsCount\": 0, \"retryCount\": 0, \"alertStatus\":{\"%s\":%s}}}", alertmanagerv1alpha1.Ready, config.AlertName, string(alertStatusBytes)))
			r.CommonClient.PatchStatus(ctx, &alertsConfig, client.RawPatch(types.MergePatchType, patch), alertmanagerv1alpha1.Ready)
			log.Info("alert successfully got updated", "alertID", alert.ID)
		}

		// Update the Alert status and also Config Status
		// No diff
		// Continue

	}

	return ctrl.Result{}, nil
}

//PatchIndividualAlertsConfigError function is a utility function to patch the error status
func (r *AlertsConfigReconciler) PatchIndividualAlertsConfigError(ctx context.Context, alertsConfig *alertmanagerv1alpha1.AlertsConfig, alertName string, err error, requeueTime ...float64) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "alertsconfig_controller", "PatchIndividualAlertsConfigError")
	log = log.WithValues("alertsConfig_cr", alertsConfig.Name, "namespace", alertsConfig.Namespace)
	alertStatus := alertsConfig.Status.Alerts[alertName]
	alertStatus.State = alertmanagerv1alpha1.Error
	alertStatusBytes, _ := json.Marshal(alertStatus)
	retryCount := alertsConfig.Status.RetryCount + 1
	log.Error(err, "error occured in alerts config for alert name", "alertName", alertName)
	r.Recorder.Event(alertsConfig, v1.EventTypeWarning, err.Error(), fmt.Sprintf("error occured in alerts config for alert name %s", alertName))

	patch := []byte(fmt.Sprintf("{\"status\":{\"state\": \"%s\", \"alertsCount\": %d, \"retryCount\": %d, \"alertStatus\":{\"%s\":%s}}}", alertmanagerv1alpha1.Error, alertsConfig.Status.AlertsCount, retryCount, alertName, string(alertStatusBytes)))
	return r.CommonClient.PatchStatus(ctx, alertsConfig, client.RawPatch(types.MergePatchType, patch), alertmanagerv1alpha1.Error, 30)
}

//HandleDelete function handles the deleting wavefront alerts
func (r *AlertsConfigReconciler) HandleDelete(ctx context.Context, alertsConfig *alertmanagerv1alpha1.AlertsConfig) error {
	log := log.Logger(ctx, "controllers", "alertsconfig_controller", "HandleDelete")
	log = log.WithValues("alertsConfig_cr", alertsConfig.Name, "namespace", alertsConfig.Namespace)
	// Lets check the status of the CR and
	// retrieve all the alerts associated with this CR and delete it
	//Check if any alerts were created with this config
	if len(alertsConfig.Status.Alerts) > 0 {
		//Call wavefront api and delete the alerts one by one
		for _, alert := range alertsConfig.Status.Alerts {
			if alert.ID != "" {
				if err := r.WavefrontClient.DeleteAlert(ctx, alert.ID); err != nil {
					log.Error(err, "skipping alert deletion", "alertID", alert.ID)
					// Just skip it for now
					// this is too opinionated but we don't want to stop the delete execution for other alerts as well
					// if there is any valid reasons not to skip it, we can look into it in future
				}

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
