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
	"fmt"
	wf "github.com/WavefrontHQ/go-wavefront-management-api"
	"github.com/keikoproj/alert-manager/pkg/log"
)

type Client struct {
	client *wf.Client
}

var ApiToken string

// NewClient returns new client instance for wavefront api with given configuration
func NewClient(ctx context.Context, config *wf.Config) (*Client, error) {
	wFClient, err := wf.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to configure Wavefront Client %s", err)
	}
	return &Client{client: wFClient}, nil
}

// CreateOrUpdateWavefrontAlert creates/update a wavefront alert
func (w *Client) CreateAlert(ctx context.Context, alert *wf.Alert) error {
	log := log.Logger(ctx, "pkg.wavefront", "CreateAlert")
	log.V(1).Info("create wavefront alert request")
	if err := ValidateAlertInput(ctx, alert); err != nil {
		log.Error(err, "unable to create the alert due to validation failed")
		return err
	}

	if err := w.client.Alerts().Create(alert); err != nil {
		log.Error(err, "unable to create the alert")
		return err
	}

	log.Info("wavefront response", "alert", *alert)

	//if err := w.client.Alerts().SetACL(*alert.ID, alert.ACL.CanView, []string{}); err != nil {
	//	log.Error(err, "unable to set the ACL")
	//	return err
	//}
	log.V(1).Info("successfully created alert", "alertID", alert.ID)
	return nil
}

func (w *Client) ReadAlert(ctx context.Context, alertID string) (alert *wf.Alert, err error) {
	log := log.Logger(ctx, "pkg.wavefront", "ReadAlert")
	log = log.WithValues("alertID", alertID)

	log.V(1).Info("Retrieving alert from Wavefront")

	alert = &wf.Alert{
		ID: &alertID,
	}
	if err := w.client.Alerts().Get(alert); err != nil {
		log.Error(err, "unable to retrieve the alert from wavefront")
		return alert, err
	}

	return alert, nil
}

func (w *Client) UpdateAlert(ctx context.Context, alert *wf.Alert) error {
	log := log.Logger(ctx, "pkg.wavefront", "UpdateAlert")
	log = log.WithValues("alertID", *alert.ID)

	log.V(1).Info("Updating an alert")

	// lets get the alert first
	_, err := w.ReadAlert(ctx, *alert.ID)

	if err != nil {
		log.Error(err, "unable to find the alert in wavefront", "alertID", *alert.ID)
		return err
	}
	if err := w.client.Alerts().Update(alert); err != nil {
		log.Error(err, "unable to retrieve the alert from wavefront")
		return err
	}
	log.Info("wavefront response", "alert", *alert)
	log.V(1).Info("successfully updated alert", "alertID", alert.ID)
	return nil
}

// DeleteWavefrontAlert deletes a specific alert from Wavefront
func (w *Client) DeleteAlert(ctx context.Context, alertID string) error {
	log := log.Logger(ctx, "pkg.wavefront", "DeleteWavefrontAlert")
	log = log.WithValues("alertID", alertID)
	log.V(1).Info("Removing an alert")

	// lets get the alert first
	alert, err := w.ReadAlert(ctx, alertID)

	if err != nil {
		log.Error(err, "unable to find the alert in wavefront. assuming alert already got deleted")
		return nil
	}
	if err := w.client.Alerts().Delete(alert, false); err != nil {
		log.Error(err, "unable to delete the alert from wavefront")
		return err
	}
	log.V(1).Info("successfully deleted the wavefront alert")
	return nil
}
