package wavefront

import "context"

type Wavefront struct {

}

//CreateOrUpdateWavefrontAlert creates/update a wavefront alert
func (w *Wavefront)	CreateOrUpdateWavefrontAlert(ctx context.Context, input AlertInput) (output AlertOutput, err error) {

	return AlertOutput{}, nil
}

// DeleteWavefrontAlert deletes a specific alert from Wavefront
func (w *Wavefront) DeleteWavefrontAlert(ctx context.Context, alertID string)error {

	return nil
}