package wavefront

import "context"

// Interface defining Alert CRUD operations

type Interface interface {
	CreateOrUpdateWavefrontAlert(ctx context.Context, input AlertInput) (output AlertOutput, err error)
	DeleteWavefrontAlert(ctx context.Context, alertID string)error
}