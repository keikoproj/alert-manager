package common

import (
	"context"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/pkg/log"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

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
