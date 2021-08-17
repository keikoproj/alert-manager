package wavefront

import (
	"context"
	wf "github.com/WavefrontHQ/go-wavefront-management-api"
	"github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/pkg/log"
)

//ConvertAlertCRToWavefrontRequest function converts wavefront alert spec to Alert API input request
func ConvertAlertCRToWavefrontRequest(ctx context.Context, req v1alpha1.WavefrontAlertSpec, alert *wf.Alert) error {
	log := log.Logger(ctx, "pkg.wavefront", "ConvertAlertCRToWavefrontRequest")
	log.V(1).Info("converting alert spec to wavefront api request")

	alert.Name = req.AlertName
	alert.Condition = req.Condition
	alert.Severity = req.Severity
	alert.AlertType = string(req.AlertType)
	alert.Tags = req.Tags
	alert.DisplayExpression = req.DisplayExpression
	alert.Minutes = int(*req.Minutes)
	alert.ResolveAfterMinutes = int(*req.ResolveAfter)
	alert.Target = req.Target
	log.V(1).Info("alert conversion is successful")
	return nil
}
