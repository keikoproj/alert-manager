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

//go:generate mockgen -destination=mocks/mock_wavefrontiface.go -package=mock_wavefront github.com/keikoproj/alert-manager/pkg/wavefront Interface

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/keikoproj/alert-manager/internal/config"
	"github.com/keikoproj/alert-manager/internal/utils"
	"github.com/keikoproj/alert-manager/pkg/log"
	"github.com/keikoproj/alert-manager/pkg/wavefront"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wf "github.com/WavefrontHQ/go-wavefront-management-api"
	_ "github.com/golang/mock/mockgen/model"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	controllercommon "github.com/keikoproj/alert-manager/internal/controllers/common"
)

const (
	wavefrontAlertFinalizerName = "wavefrontalert.finalizers.alertmanager.keikoproj.io"
	requestId                   = "request_id"
	//30 seconds
	errRequeueTime = 30000
)

// WavefrontAlertReconciler reconciles a WavefrontAlert object
type WavefrontAlertReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	CommonClient    *controllercommon.Client
	WavefrontClient wavefront.Interface
}

//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=alertmanager.keikoproj.io,resources=wavefrontalerts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=alertmanager.keikoproj.io,resources=wavefrontalerts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=alertmanager.keikoproj.io,resources=wavefrontalerts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WavefrontAlert object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *WavefrontAlertReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	ctx = context.WithValue(ctx, requestId, uuid.New())
	log := log.Logger(ctx, "controllers", "wavefrontalert_controller", "Reconcile")
	log = log.WithValues("wavefrontalert_cr", req.NamespacedName)
	log.Info("Start of the request")

	var wfAlert alertmanagerv1alpha1.WavefrontAlert
	if err := r.Get(ctx, req.NamespacedName, &wfAlert); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//var status alertmanagerv1alpha1.WavefrontAlertStatus
	//Main responsibilities of the Wavefront Alert Controller
	// Check if it is delete request
	if !wfAlert.ObjectMeta.DeletionTimestamp.IsZero() {
		requeueFlag := false
		// Delete use case
		if err := r.HandleDelete(ctx, &wfAlert); err != nil {
			log.Error(err, "unable to delete the alert")
			requeueFlag = true
		}
		return ctrl.Result{Requeue: requeueFlag}, nil
	}

	//First time use case
	if !utils.ContainsString(wfAlert.ObjectMeta.Finalizers, wavefrontAlertFinalizerName) {
		log.Info("New wavefront alert resource. Adding the finalizer", "finalizer", wavefrontAlertFinalizerName)
		wfAlert.ObjectMeta.Finalizers = append(wfAlert.ObjectMeta.Finalizers, wavefrontAlertFinalizerName)
		r.CommonClient.UpdateMeta(ctx, &wfAlert)
		//That's fine- Let it come for requeue and we can create the alert
		return ctrl.Result{}, nil
	}
	// Calculate the checksum
	data, err := json.Marshal(wfAlert.Spec)
	if err != nil {
		return r.UpdateIndividualWavefrontAlertStatusError(ctx, &wfAlert, alertmanagerv1alpha1.Error, err, errRequeueTime)
	}
	lastChangeChecksum := utils.CalculateChecksum(ctx, string(data))
	wfAlert.Status.LastChangeChecksum = lastChangeChecksum
	// Check for exportedParams length
	exportedParamslength := 0
	proceed := true
	if wfAlert.Spec.ExportedParams != nil {
		exportedParamslength = len(wfAlert.Spec.ExportedParams)
	}

	if wfAlert.Status.ObservedGeneration == wfAlert.ObjectMeta.Generation && wfAlert.Status.State != alertmanagerv1alpha1.Error {
		proceed = false
	}

	if exportedParamslength > 0 {
		// Check for exportedParams checksum
		exist, reqChecksum := utils.ExportParamsChecksum(ctx, wfAlert.Spec.ExportedParams)
		// if status doesn't have checksum- it means its very first request
		if exist && wfAlert.Status.ExportParamsChecksum != "" {
			log.Info("checksum difference", "reqChecksum", reqChecksum, "checksum", wfAlert.Status.ExportParamsChecksum)
			if reqChecksum != wfAlert.Status.ExportParamsChecksum {
				proceed = false
				wfAlert.Status.State = alertmanagerv1alpha1.ReadyToBeUsed
				//TODO: Think about this again
				wfAlert.Status.ExportParamsChecksum = reqChecksum
			}
		}
	}
	// If user changed exported params and wavefront alert spec (ObservedGeneration) at the same time, we cannot perform any
	// change because we need the substituted value of that exported param
	// If there is a change in wavefront alert spec (ObservedGeneration) and NO CHANGE in exported Params,
	// we need to loop through all the alertsconfig CR under status
	// and fill up with the substituted params value then apply the change in wavefront for individual alert
	if exportedParamslength > 0 && proceed && wfAlert.Status.ObservedGeneration != wfAlert.ObjectMeta.Generation {
		// wavefrontalerts spec change
		log.Info("wavefrontalerts spec was changed, processing to update individual alert that associated with it")
		wfAlert.Status.LastChangeChecksum = lastChangeChecksum
		for _, c := range wfAlert.Status.AlertsStatus {
			// Get alertsconfig CR
			alertConfigNamespacedName := types.NamespacedName{Namespace: req.Namespace, Name: c.AssociatedAlertsConfig.CR}
			var alertsConfig alertmanagerv1alpha1.AlertsConfig
			if err := r.Get(ctx, alertConfigNamespacedName, &alertsConfig); err != nil {
				return r.UpdateIndividualWavefrontAlertStatusError(ctx, &wfAlert, alertmanagerv1alpha1.Error, err, errRequeueTime)
			}
			err := r.UpdateIndividualWavefrontAlert(ctx, req, c, alertsConfig, wfAlert)
			if err != nil {
				state := alertmanagerv1alpha1.Error
				if strings.Contains(err.Error(), "Exceeded limit setting") {
					state = alertmanagerv1alpha1.ClientExceededLimit
				}
				// Update the state to be error for each alert
				c.State = state
				if err := r.CommonClient.PatchWfAlertAndAlertsConfigStatus(ctx, c.State, &wfAlert, &alertsConfig, c, errRequeueTime); err != nil {
					log.Error(err, "unable to patch wfalert and alertsconfig status objects")
					return r.UpdateIndividualWavefrontAlertStatusError(ctx, &wfAlert, state, err, errRequeueTime)
				}
				return r.UpdateIndividualWavefrontAlertStatusError(ctx, &wfAlert, state, err, errRequeueTime)
			}
			// Update the state to be ready for each alert
			c.State = alertmanagerv1alpha1.Ready
			if err := r.CommonClient.PatchWfAlertAndAlertsConfigStatus(ctx, c.State, &wfAlert, &alertsConfig, c); err != nil {
				log.Error(err, "unable to patch wfalert and alertsconfig status objects")
				return r.UpdateIndividualWavefrontAlertStatusError(ctx, &wfAlert, alertmanagerv1alpha1.Error, err, errRequeueTime)
			}
		}
		wfAlert.Status.ObservedGeneration = wfAlert.ObjectMeta.Generation
		// We are going to stop here because we already updated individual alerts that associate with
		// this wavefront alert template
		return r.CommonClient.UpdateStatus(ctx, &wfAlert, alertmanagerv1alpha1.Ready, errRequeueTime)
	}

	wfAlert.Status.ObservedGeneration = wfAlert.ObjectMeta.Generation
	if !proceed {
		// do nothing
		log.Info("There is no change in the spec.. skipping")
		//wfAlert.Status = status
		return ctrl.Result{}, r.Status().Update(ctx, &wfAlert)
	}

	// Validate the alert request
	var alert wf.Alert
	r.convertAlertCR(ctx, &wfAlert, &alert)
	if err := wavefront.ValidateAlertInput(ctx, &alert); err != nil {
		log.Error(err, "Failed to validate wavefront alert input")
		wfAlert.Status.LastChangeChecksum = lastChangeChecksum
		wfAlert.Status.State = alertmanagerv1alpha1.Error
		wfAlert.Status.ErrorDescription = err.Error()

		// First update status directly to ensure it's set immediately
		if updateErr := r.Status().Update(ctx, &wfAlert); updateErr != nil {
			log.Error(updateErr, "Failed to update status immediately after validation error")
		}

		// Then use the standard error handler which may requeue
		return r.UpdateIndividualWavefrontAlertStatusError(ctx, &wfAlert, alertmanagerv1alpha1.Error, err, errRequeueTime)
	}

	// so simple validation is done so lets Handle reconcile
	// Keep it simple - If already exist- call updateAlert with existing alertID
	// If alert doesn't exist- create Alert
	// TODO: In future, check if we need to do GET API call to get the existing alert and
	//  compare it with the request to see if changes are really needed

	if len(wfAlert.Status.AlertsStatus) == 0 {
		// New alert
		// First time use case
		// Lets create an alert
		var alert wf.Alert
		r.convertAlertCR(ctx, &wfAlert, &alert)
		log.V(1).Info("alert values", "alertObj", alert)
		if err := r.WavefrontClient.CreateAlert(ctx, &alert); err != nil {
			state := alertmanagerv1alpha1.Error
			if strings.Contains(err.Error(), "Exceeded limit setting") {
				// For ex: error is "Exceeded limit setting: 100 alerts allowed per customer"
				state = alertmanagerv1alpha1.ClientExceededLimit
			}
			log.Error(err, "unable to create the alert")
			wfAlert.Status.LastChangeChecksum = lastChangeChecksum
			return r.UpdateIndividualWavefrontAlertStatusError(ctx, &wfAlert, state, err, errRequeueTime)
		}

		alertResponse := alertmanagerv1alpha1.AlertStatus{
			ID:                 *alert.ID,
			Name:               alert.Name,
			Link:               fmt.Sprintf("https://%s/alerts/%s", config.Props.WavefrontAPIUrl(), *alert.ID),
			LastChangeChecksum: lastChangeChecksum,
		}
		alertsStatus := make(map[string]alertmanagerv1alpha1.AlertStatus)
		alertsStatus[alertResponse.Name] = alertResponse
		wfAlert.Status.RetryCount = 0
		wfAlert.Status.AlertsStatus = alertsStatus
		wfAlert.Status.ObservedGeneration = wfAlert.ObjectMeta.Generation
		return r.CommonClient.UpdateStatus(ctx, &wfAlert, alertmanagerv1alpha1.Ready, errRequeueTime)
	}

	// existing alert - Perform the updateAlert one by one
	// this is for standalone alerts not alertsconfig scenario
	currStatus := wfAlert.Status.AlertsStatus
	for _, a := range wfAlert.Status.AlertsStatus {
		// Create a local copy of ID to avoid memory aliasing in loop
		id := a.ID
		alert := wf.Alert{
			ID: &id,
		}
		r.convertAlertCR(ctx, &wfAlert, &alert)
		state := alertmanagerv1alpha1.Ready
		respAlert := a
		// TODO: Only do the UpdateAlert if there is a difference between parent lastChangeChecksum and child lastChangeChecksum- This could be in a scenario
		//  where it updated 99 out of 100 child alerts and 1 got failed and it got requeued. so instead of trying to update 100 again lets just do only 1 api
		// call update api
		if err := r.WavefrontClient.UpdateAlert(ctx, &alert); err != nil {
			r.Recorder.Event(&wfAlert, v1.EventTypeWarning, err.Error(), "unable to update the alert")
			state = alertmanagerv1alpha1.Error
			if strings.Contains(err.Error(), "Exceeded limit setting") {
				// For ex: error is "Exceeded limit setting: 100 alerts allowed per customer"
				state = alertmanagerv1alpha1.ClientExceededLimit
			}
			log.Error(err, "unable to update the alert")
			respAlert.State = state
			respAlert.LastChangeChecksum = a.LastChangeChecksum
			// if even one of the child got failed, make parent status as error
			wfAlert.Status.State = state
			wfAlert.Status.RetryCount = wfAlert.Status.RetryCount + 1
		}
		log.Info("alert ids before and after", "before", a.ID, "after", alert.ID)
		respAlert.State = state
		//TODO: Figure out a better way to handle this in future when we have multiple
		wfAlert.Status.State = state
		currStatus[respAlert.Name] = respAlert
	}

	if wfAlert.Status.State == alertmanagerv1alpha1.Ready {
		wfAlert.Status.RetryCount = 0
		wfAlert.Status.ErrorDescription = ""
	}
	wfAlert.Status.AlertsStatus = currStatus
	wfAlert.Status.ObservedGeneration = wfAlert.ObjectMeta.Generation
	return r.CommonClient.UpdateStatus(ctx, &wfAlert, wfAlert.Status.State, errRequeueTime)
}

// PatchIndividualAlertsStatusError function is a utility function to patch the error status
// We use status patch instead of status update to avoid any overwrite between two threads when alertsConfig CR has multiple alert configs
func (r *WavefrontAlertReconciler) PatchIndividualAlertsStatusError(ctx context.Context, wfAlert *alertmanagerv1alpha1.WavefrontAlert, alertName string, state alertmanagerv1alpha1.State, err error, requeueTime ...float64) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "alertsconfig_controller", "PatchIndividualAlertsStatusError")
	log = log.WithValues("alertsConfig_cr", wfAlert.Name, "namespace", wfAlert.Namespace)
	alertStatus := wfAlert.Status.AlertsStatus[alertName]
	alertStatus.State = state
	alertStatus.ErrorDescription = err.Error()
	alertStatus.LastUpdatedTimestamp = metav1.Now()
	alertStatusBytes, _ := json.Marshal(alertStatus)
	retryCount := wfAlert.Status.RetryCount + 1
	log.Error(err, "error occured in alerts config for alert name", "alertName", alertName)
	r.Recorder.Event(wfAlert, v1.EventTypeWarning, err.Error(), fmt.Sprintf("error occured in wavefront controller for alert name %s", alertName))

	patch := []byte(fmt.Sprintf("{\"status\":{\"state\": \"%s\", \"retryCount\": %d, \"alertStatus\":{\"%s\":%s}}}", state, retryCount, alertName, string(alertStatusBytes)))
	return r.CommonClient.PatchStatus(ctx, wfAlert, client.RawPatch(types.MergePatchType, patch), alertmanagerv1alpha1.Error, errRequeueTime)
}

// HandleDelete function handles the deleting wavefront alerts
func (r *WavefrontAlertReconciler) HandleDelete(ctx context.Context, wfAlert *alertmanagerv1alpha1.WavefrontAlert) error {
	log := log.Logger(ctx, "controllers", "wavefrontalert_controller", "HandleDelete")
	log = log.WithValues("wavefrontalert_cr", wfAlert.Name, "namespace", wfAlert.Namespace)
	// Lets check the status of the CR and
	// retrieve all the alerts associated with this CR and delete it
	//Check if any alerts were created with this config
	if len(wfAlert.Status.AlertsStatus) > 0 {
		//Call wavefront api and delete the alerts one by one
		for _, alert := range wfAlert.Status.AlertsStatus {
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
	wfAlert.ObjectMeta.Finalizers = utils.RemoveString(wfAlert.ObjectMeta.Finalizers, wavefrontAlertFinalizerName)
	r.CommonClient.UpdateMeta(ctx, wfAlert)
	log.Info("Successfully deleted wfAlert")
	r.Recorder.Event(wfAlert, v1.EventTypeNormal, "Deleted", "Successfully deleted WavefrontAlert")
	return nil
}

// convertAlertCR converts alert CR to wf.Alert
func (r *WavefrontAlertReconciler) convertAlertCR(ctx context.Context, wfAlert *alertmanagerv1alpha1.WavefrontAlert, alert *wf.Alert) {
	log := log.Logger(ctx, "controllers", "wavefrontalert_controller", "convertAlertCR")
	log = log.WithValues("wavefrontalert_cr", wfAlert.Name, "namespace", wfAlert.Namespace)
	if err := wavefront.ConvertAlertCRToWavefrontRequest(ctx, wfAlert.Spec, alert); err != nil {
		errMsg := "unable to convert the wavefront spec to Alert API request. will not be retried"
		log.Error(err, errMsg)
		r.Recorder.Event(wfAlert, v1.EventTypeWarning, "MalformedSpec", errMsg)
		wfAlert.Status = alertmanagerv1alpha1.WavefrontAlertStatus{
			RetryCount:       wfAlert.Status.RetryCount + 1,
			ErrorDescription: errMsg,
			State:            alertmanagerv1alpha1.MalformedSpec,
		}
		// There is no use of requeue in this case
		if _, updateErr := r.CommonClient.UpdateStatus(ctx, wfAlert, alertmanagerv1alpha1.MalformedSpec); updateErr != nil {
			log.Error(updateErr, "Failed to update WavefrontAlert status")
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *WavefrontAlertReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alertmanagerv1alpha1.WavefrontAlert{}).
		WithEventFilter(controllercommon.StatusUpdatePredicate{}).
		Complete(r)
}

func (r *WavefrontAlertReconciler) UpdateIndividualWavefrontAlertStatusError(
	ctx context.Context,
	wfAlert *alertmanagerv1alpha1.WavefrontAlert,
	state alertmanagerv1alpha1.State,
	err error,
	requeueTime ...float64,
) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "wavefront_alerts_controller", "UpdateIndividualWavefrontAlertStatusError")
	if len(requeueTime) == 0 {
		requeueTime = []float64{errRequeueTime}
	}

	log.Error(err, "error occurred in wavefront alert", "wavefrontAlert", wfAlert.Name)
	r.Recorder.Event(wfAlert, v1.EventTypeWarning, err.Error(), fmt.Sprintf("error occurred in wavefront alert %s", wfAlert.Name))
	return r.CommonClient.UpdateStatus(ctx, wfAlert, state, requeueTime[0])
}

func (r *WavefrontAlertReconciler) UpdateIndividualWavefrontAlert(
	ctx context.Context,
	req ctrl.Request,
	alertStatus alertmanagerv1alpha1.AlertStatus,
	alertsConfig alertmanagerv1alpha1.AlertsConfig,
	wfAlert alertmanagerv1alpha1.WavefrontAlert,
) error {
	log := log.Logger(ctx, "controllers", "wavefrontalert_controller", "UpdateIndividualWavefrontAlert")
	// Get the corresponding alert in alertsConfig
	config := alertsConfig.Spec.Alerts[alertStatus.AssociatedAlert.CR]
	var alert wf.Alert
	globalMap := alertsConfig.Spec.GlobalParams
	//merge the alerts config global params and individual params
	params := utils.MergeMaps(ctx, globalMap, config.Params)
	// Create wavefront alert with proper substituted value of that exported param
	if err := controllercommon.GetProcessedWFAlert(ctx, &wfAlert, params, &alert); err != nil {
		return err
	}
	// Update alert in wavefront
	alert.ID = &alertStatus.ID
	// Validate the alert request
	if err := wavefront.ValidateAlertInput(ctx, &alert); err != nil {
		return err
	}

	if err := r.WavefrontClient.UpdateAlert(ctx, &alert); err != nil {
		return err
	}
	log.Info("alert successfully got updated", "alertID", alert.ID)
	return nil
}
