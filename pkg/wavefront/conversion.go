/*
Copyright 2025 Keikoproj authors.

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

package wavefront

import (
	"context"
	"errors"
	wf "github.com/WavefrontHQ/go-wavefront-management-api"
	"github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/pkg/log"
)

// ConvertAlertCRToWavefrontRequest function converts wavefront alert spec to Alert API input request
func ConvertAlertCRToWavefrontRequest(ctx context.Context, req v1alpha1.WavefrontAlertSpec, alert *wf.Alert) error {
	log := log.Logger(ctx, "pkg.wavefront", "ConvertAlertCRToWavefrontRequest")
	log.V(1).Info("converting alert spec to wavefront api request")

	alert.Name = req.AlertName
	alert.Condition = req.Condition
	alert.Severity = req.Severity
	alert.AlertType = string(req.AlertType)
	alert.Tags = req.Tags
	alert.DisplayExpression = req.DisplayExpression
	alert.AdditionalInfo = req.AdditionalInformation
	if req.Minutes == nil || req.ResolveAfter == nil {
		err := errors.New("minutes and resolveAfter must be passed")
		log.Error(err, "error occurred in ConvertAlertCRToWavefrontRequest")
		return err
	}
	alert.Minutes = int(*req.Minutes)
	alert.ResolveAfterMinutes = int(*req.ResolveAfter)
	alert.Target = req.Target
	if req.AlertCheckFrequency != 0 {
		alert.CheckingFrequencyInMinutes = req.AlertCheckFrequency
	}
	log.V(1).Info("alert conversion is successful")
	return nil
}
