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
	"fmt"
	"github.com/WavefrontHQ/go-wavefront-management-api"
	"github.com/keikoproj/alert-manager/internal/utils"
	"github.com/keikoproj/alert-manager/pkg/log"
)

// ValidateAlertInput validates alert inputs
func ValidateAlertInput(ctx context.Context, input *wavefront.Alert) error {
	log := log.Logger(ctx, "pkg.wavefront", "validateAlertInput")
	log.V(1).Info("validating input request")

	// Validate common required fields first
	if input.Name == "" {
		err := errors.New("validation failed: alertName must not be empty")
		log.Error(err, "validation failed: alertName must not be empty")
		return err
	}

	if input.DisplayExpression == "" {
		err := errors.New("validation failed: displayExpression must not be empty")
		log.Error(err, "validation failed: displayExpression must not be empty")
		return err
	}

	// Validate Minutes and ResolveAfterMinutes
	if input.Minutes <= 0 {
		err := errors.New("validation failed: minutes must be greater than 0")
		log.Error(err, "validation failed: minutes must be greater than 0")
		return err
	}

	if input.ResolveAfterMinutes <= 0 {
		err := errors.New("validation failed: resolveAfterMinutes must be greater than 0")
		log.Error(err, "validation failed: resolveAfterMinutes must be greater than 0")
		return err
	}

	// Validate alert type specific conditions
	if err := validateAlertConditions(ctx, input); err != nil {
		return err
	}

	return nil
}

func validateAlertConditions(ctx context.Context, input *wavefront.Alert) error {
	log := log.Logger(ctx, "pkg.wavefront", "validateAlertConditions")
	log.V(1).Info("validating condition/s from input request", "alert", input)
	if input.AlertType == wavefront.AlertTypeThreshold {
		if len(input.Conditions) != 0 {
			if err := validateThresholdLevels(ctx, utils.TrimSpacesMap(input.Conditions)); err != nil {
				log.Error(err, "validation failed: invalid severity mentioned in conditions")
				return err
			}
		} else {
			err := errors.New("validation failed: conditions must not be empty")
			log.Error(err, "validation failed: conditions must not be empty")
			return err
		}

	} else if input.AlertType == wavefront.AlertTypeClassic {
		if input.Condition == "" {
			err := errors.New("validation failed: condition must not be empty")
			log.Error(err, "validation failed: condition must not be empty")
			return err
		}

		if input.Severity != "" {
			if err := validateSeverity(ctx, input.Severity); err != nil {
				log.Error(err, "validation failed: invalid severity mentioned in the request")
				return err
			}
		} else {
			err := errors.New("validation failed: severity must not be empty")
			log.Error(err, "validation failed: severity must not be empty")
			return err
		}
	} else {
		err := fmt.Errorf("validation failed: invalid alert type: %s", input.AlertType)
		log.Error(err, "validation failed: invalid alert type")
		return err
	}

	return nil
}

// validateThresholdLevels validates threshold values included in the request
func validateThresholdLevels(ctx context.Context, m map[string]string) error {
	log := log.Logger(ctx, "pkg.wavefront", "validateThresholdLevels")
	log.V(1).Info("validating threshold values")
	for key := range m {
		return validateSeverity(ctx, key)
	}
	return nil
}

func validateSeverity(ctx context.Context, key string) error {
	log := log.Logger(ctx, "pkg.wavefront", "validateSeverity")
	log.V(1).Info("validating severity values")
	ok := false
	for _, level := range []string{"severe", "warn", "info", "smoke"} {
		if key == level {
			ok = true
			break
		}
	}
	if !ok {
		err := fmt.Errorf("invalid severity: %s", key)
		log.Error(err, "invalid severity found")
		return err
	}
	return nil
}

// ValidateTemplateParams function validates whether all the required template exported params been supplied in alert config
func ValidateTemplateParams(ctx context.Context, exportParams []string, configValues map[string]string) error {
	log := log.Logger(ctx, "pkg.wavefront", "validateTemplateParams")
	log.V(1).Info("validating export params with config params")
	for _, param := range exportParams {
		if _, ok := configValues[param]; !ok {
			return fmt.Errorf("Required exported param %s is not supplied. ", param)
		}
	}
	return nil
}
