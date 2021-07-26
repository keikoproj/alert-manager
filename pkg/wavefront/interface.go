package wavefront

import (
	"context"

	wf "github.com/WavefrontHQ/go-wavefront-management-api"
)

// Interface defining Alert CRUD operations

type Interface interface {
	CreateAlert(ctx context.Context, input *wf.Alert) error
	ReadAlert(ctx context.Context, alertID string) (output *wf.Alert, err error)
	UpdateAlert(ctx context.Context, input *wf.Alert) error
	DeleteAlert(ctx context.Context, alertID string) error
}
