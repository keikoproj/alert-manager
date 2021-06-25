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
	"fmt"
	"github.com/google/uuid"
	"github.com/keikoproj/alert-manager/internal/utils"
	"github.com/keikoproj/alert-manager/pkg/k8s"
	"github.com/keikoproj/alert-manager/pkg/log"
	"github.com/keikoproj/alert-manager/pkg/wavefront"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	controllercommon "github.com/keikoproj/alert-manager/controllers/common"
)

const (
	wavefrontAlertFinalizerName = "wavefrontalert.finalizers.alertmanager.keikoproj.io"
	requestId                   = "request_id"
	//2 minutes
	maxWaitTime = 120000
	//30 seconds
	errRequeueTime = 300000
)

// WavefrontAlertReconciler reconciles a WavefrontAlert object
type WavefrontAlertReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	K8sSelfClient *k8s.Client
	Recorder      record.EventRecorder
	CommonClient  *controllercommon.Client
	wavefrontClient wavefront.Interface
}

//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create
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
	//Main responsibilities of the Wavefront Alert Controller
	// Check if it is delete request
	if !wfAlert.ObjectMeta.DeletionTimestamp.IsZero() {
		// Delete use case
		return ctrl.Result{}, r.HandleDelete(ctx, &wfAlert)
	}
	// Check for exportedParams length
	length := 0
	proceed := true
	if wfAlert.Spec.ExportedParams != nil {
		length = len(wfAlert.Spec.ExportedParams)
	}
	if length > 0 {
		// Check for exportedParams checksum
		exist, reqChecksum := utils.ExportParamsChecksum(ctx, wfAlert.Spec.ExportedParams)
		if exist && &wfAlert.Status.ExportParamsChecksum != nil {
			if reqChecksum != wfAlert.Status.ExportParamsChecksum {
				proceed = false
			}
		}
	}

	if !proceed {
		// do nothing
		log.Info("exportedParams checksum changed. Not proceeding further..")
		return ctrl.Result{}, nil
	}

	switch wfAlert.Status.State {
	case "":
		//Brand new alert creation
	case alertmanagerv1alpha1.Ready:
		//Probably an update
	case alertmanagerv1alpha1.Error:
		// Its an Error- so retry

	}

	// Call wavefront apis and create/update the alerts

	// look for the status and call wavefront APIs to delete alert/s
	// if its create/update
	// if exportedParams is empty, proceed with wavefront apis to create the alert
	// exportedParams is not empty and no status available- This is considered as brand new request
	// don't do anything
	// if there is a diff in exportedParams checksum compared to status
	// don't do anything
	// else
	// look for the status and proceed with wavefront apis to update the alert

	return ctrl.Result{}, nil
}

//HandleDelete function handles the deleting wavefront alerts
func (r *WavefrontAlertReconciler) HandleDelete(ctx context.Context, wfAlert *alertmanagerv1alpha1.WavefrontAlert) error {
	log := log.Logger(ctx, "controllers", "wavefrontalert_controller", "HandleDelete")
	log = log.WithValues("wavefrontalert_cr", wfAlert.Name, "namespace", wfAlert.Namespace)
	// Lets check the status of the CR and
	// retrieve all the alerts associated with this CR and delete it
	//Check if any alerts were created with this config
	if len(wfAlert.Status.Alerts) > 0 {
		//Call wavefront api and delete the alerts one by one
	}

	// Ok. Lets delete the finalizer so controller can delete the custom object
	log.Info("Removing finalizer from WavefrontAlert")
	wfAlert.ObjectMeta.Finalizers = utils.RemoveString(wfAlert.ObjectMeta.Finalizers, wavefrontAlertFinalizerName)
	r.CommonClient.UpdateMeta(ctx, wfAlert)
	log.Info("Successfully deleted wfAlert")
	r.Recorder.Event(wfAlert, v1.EventTypeNormal, "Deleted", "Successfully deleted WavefrontAlert")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WavefrontAlertReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&alertmanagerv1alpha1.WavefrontAlert{}).
		Complete(r)
}
