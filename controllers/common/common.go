package common

import (
	"context"
	"encoding/json"
	"fmt"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/internal/template"
	"github.com/keikoproj/alert-manager/pkg/log"
	"github.com/keikoproj/alert-manager/pkg/wavefront"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	wf "github.com/WavefrontHQ/go-wavefront-management-api"
	ctrl "sigs.k8s.io/controller-runtime"
)

type StatusUpdatePredicate struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating generation change
func (StatusUpdatePredicate) Update(e event.UpdateEvent) bool {
	log := log.Logger(context.Background(), "controllers.common", "Update")

	if e.ObjectOld == nil {
		log.Error(nil, "Update event has no old runtime object to update", "event", e)
		return false
	}
	if e.ObjectNew == nil {
		log.Error(nil, "Update event has no new runtime object for update", "event", e)
		return false
	}

	//Better way to do it is to get GVK from ObjectKind but Kind is dropped during decode.
	//For more details, check the status of the issue here
	//https://github.com/kubernetes/kubernetes/issues/80609

	// Try to type caste to WavefrontAlert first if it doesn't work move to namespace type casting
	if oldWFAlertObj, ok := e.ObjectOld.(*alertmanagerv1alpha1.WavefrontAlert); ok {
		newWFAlertObj := e.ObjectNew.(*alertmanagerv1alpha1.WavefrontAlert)
		if !reflect.DeepEqual(oldWFAlertObj.Status, newWFAlertObj.Status) {
			return false
		}
	}

	if oldAlertsConfigObj, ok := e.ObjectOld.(*alertmanagerv1alpha1.AlertsConfig); ok {
		newAlertsConfigObj := e.ObjectNew.(*alertmanagerv1alpha1.AlertsConfig)
		if !reflect.DeepEqual(oldAlertsConfigObj.Status, newAlertsConfigObj.Status) {
			return false
		}
	}

	return true
}

// Client is a manager client to get the common stuff for all the controllers
type Client struct {
	client.Client
	Recorder record.EventRecorder
}

//UpdateMeta function updates the metadata (mostly finalizers in this case)
//This function accepts runtime.Object which can be either cluster type or namespace type
func (r *Client) UpdateMeta(ctx context.Context, object client.Object) {
	log := log.Logger(ctx, "controllers.common", "UpdateMeta")
	if err := r.Update(ctx, object); err != nil {
		log.Error(err, "Unable to update object metadata (finalizer)")
		panic(err)
	}
	log.Info("successfully updated the meta")
}

//UpdateStatus function updates the status based on the process step
func (r *Client) UpdateStatus(ctx context.Context, obj client.Object, state alertmanagerv1alpha1.State, requeueTime ...float64) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers.common", "common", "UpdateStatus")

	if err := r.Status().Update(ctx, obj); err != nil {
		log.Error(err, "Unable to update status", "status", state)
		r.Recorder.Event(obj, v1.EventTypeWarning, string(alertmanagerv1alpha1.Error), "Unable to create/update status due to error "+err.Error())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	if state != alertmanagerv1alpha1.Error {
		return ctrl.Result{}, nil
	}

	//if wait time is specified, requeue it after provided time
	if len(requeueTime) == 0 {
		requeueTime[0] = 0
	}

	log.Info("Requeue time", "time", requeueTime[0])
	return ctrl.Result{RequeueAfter: time.Duration(requeueTime[0]) * time.Millisecond}, nil
}

//PatchStatus function patches the status based on the process step
func (r *Client) PatchStatus(ctx context.Context, obj client.Object, patch client.Patch, state alertmanagerv1alpha1.State, requeueTime ...float64) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers.common", "common", "PatchStatus")

	if err := r.Status().Patch(ctx, obj, patch); err != nil {
		log.Error(err, "Unable to patch the status", "status", state)
		r.Recorder.Event(obj, v1.EventTypeWarning, string(alertmanagerv1alpha1.Error), "Unable to patch status due to error "+err.Error())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	if state != alertmanagerv1alpha1.Error {
		return ctrl.Result{}, nil
	}

	//if wait time is specified, requeue it after provided time
	if len(requeueTime) == 0 {
		requeueTime[0] = 0
	}

	log.Info("Requeue time", "time", requeueTime[0])
	return ctrl.Result{RequeueAfter: time.Duration(requeueTime[0]) * time.Millisecond}, nil
}

//ConvertAlertCR converts alert CR to wf.Alert
func (r *Client) ConvertAlertCR(ctx context.Context, wfAlert *alertmanagerv1alpha1.WavefrontAlert, alert *wf.Alert) {
	log := log.Logger(ctx, "controllers", "wavefrontalert_controller", "convertAlertCR")
	//log = log.WithValues("wavefrontalert_cr", wfAlert.Name, "namespace", wfAlert.Namespace)
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
		r.UpdateStatus(ctx, wfAlert, alertmanagerv1alpha1.MalformedSpec)
	}
}

//GetProcessedWFAlert function converts wavefront alert spec to wavefront api request by processing template with the values provided in alerts config
func GetProcessedWFAlert(ctx context.Context, wfAlert *alertmanagerv1alpha1.WavefrontAlert, config *alertmanagerv1alpha1.Config, alert *wf.Alert) error {
	log := log.Logger(ctx, "controllers", "common", "GetProcessedWFAlert")
	log = log.WithValues("alertsConfig_cr", wfAlert.Name)

	wfAlertBytes, err := json.Marshal(wfAlert.Spec)
	if err != nil {
		// update the status and retry it
		return err
	}

	if err := wavefront.ValidateTemplateParams(ctx, wfAlert.Spec.ExportedParams, config.Params); err != nil {
		return err
	}

	// execute Golang Template
	wfAlertTemplate, err := template.ProcessTemplate(ctx, string(wfAlertBytes), config.Params)
	if err != nil {
		//update the status and retry it
		return err
	}
	log.Info("Template process is successful", "here", wfAlertTemplate)

	// Unmarshal back to wavefront alert
	if err := json.Unmarshal([]byte(wfAlertTemplate), &wfAlert.Spec); err != nil {
		// update the wfAlert status and retry it
		return err
	}
	// Convert to Alert
	if err := wavefront.ConvertAlertCRToWavefrontRequest(ctx, wfAlert.Spec, alert); err != nil {
		errMsg := "unable to convert the wavefront spec to Alert API request. will not be retried"
		log.Error(err, errMsg)
		return err
	}

	// Validate the alert- just make sure severity and other required fields are properly replaced/substituted
	if err := wavefront.ValidateAlertInput(ctx, alert); err != nil {
		return err
	}
	return nil
}

//PatchWfAlertAndAlertsConfigStatus function patches the individual alert status for both wavefront alert and alerts config
func (r *Client) PatchWfAlertAndAlertsConfigStatus(ctx context.Context, wfAlert *alertmanagerv1alpha1.WavefrontAlert, alertsConfig *alertmanagerv1alpha1.AlertsConfig, alertStatus alertmanagerv1alpha1.AlertStatus) error {
	log := log.Logger(ctx, "controllers", "common", "PatchWfAlertAndAlertsConfigStatus")
	log = log.WithValues("wfAlertCR", wfAlert.Name, "alertsConfigCR", alertsConfig.Name)

	alertStatusBytes, _ := json.Marshal(alertStatus)
	patch := []byte(fmt.Sprintf("{\"status\":{\"state\": \"%s\", \"alertsStatus\":{\"%s\":%s}}}", alertmanagerv1alpha1.Ready, wfAlert.Name, string(alertStatusBytes)))
	_, err := r.PatchStatus(ctx, alertsConfig, client.RawPatch(types.MergePatchType, patch), alertmanagerv1alpha1.Ready)
	if err != nil {
		log.Error(err, "unable to patch the status for alerts config object")
		return err
	}
	wfAlertStatusPatch := []byte(fmt.Sprintf("{\"status\":{\"state\": \"%s\",\"alertsStatus\":{\"%s\":%s}}}", alertmanagerv1alpha1.Ready, alertsConfig.Name, string(alertStatusBytes)))
	if err != nil {
		log.Error(err, "unable to patch the status for wavefront alert object")
		return err
	}
	_, err = r.PatchStatus(ctx, wfAlert, client.RawPatch(types.MergePatchType, wfAlertStatusPatch), alertmanagerv1alpha1.Ready)
	log.Info("alert successfully got updated for both wavefront alert and alerts config objects")

	return nil
}
