package wavefront

import (
	"context"
	"errors"
	"fmt"
	"github.com/WavefrontHQ/go-wavefront-management-api"
	"github.com/keikoproj/alert-manager/internal/utils"
	"github.com/keikoproj/alert-manager/pkg/log"
)

//ValidateAlertInput validates alert inputs
func ValidateAlertInput(ctx context.Context, input *wavefront.Alert) error {
	log := log.Logger(ctx, "pkg.wavefront", "validateAlertInput")
	log.V(1).Info("validating input request")
	if err := validateAlertConditions(ctx, input); err != nil {
		return err
	}
	return nil
}

func validateAlertConditions(ctx context.Context, input *wavefront.Alert) error {
	log := log.Logger(ctx, "pkg.wavefront", "validateAlertConditions")
	log.V(1).Info("validating condition/s from input request")
	if input.AlertType == wavefront.AlertTypeThreshold {
		if len(input.Conditions) != 0 {
			if err := validateThresholdLevels(ctx, utils.TrimSpacesMap(input.Conditions)); err != nil {
				log.Error(err, "invalid severity mentioned in conditions")
				return err
			}
		} else {
			msg := fmt.Sprintf("conditions must not be empty")
			err := errors.New(msg)
			log.Error(err, msg)
			return err
		}

	} else if input.AlertType == wavefront.AlertTypeClassic {
		if input.Condition == "" {
			msg := fmt.Sprintf("condition must not be empty")
			err := errors.New(msg)
			log.Error(err, msg)
			return err
		}

		if input.Severity != "" {
			if err := validateSeverity(ctx, input.Severity); err != nil {
				log.Error(err, "invalid severity mentioned in the request")
				return err
			}
		} else {
			msg := fmt.Sprintf("severity must not be empty")
			err := errors.New(msg)
			log.Error(err, msg)
			return err
		}
	} else {
		msg := fmt.Sprintf("invalid alert type: %s", input.AlertType)
		err := errors.New(msg)
		log.Error(err, msg)
		return err
	}

	return nil
}

//validateThresholdLevels validates threshold values included in the request
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
		msg := fmt.Sprintf("invalid severity: %s", key)
		err := errors.New(msg)
		log.Error(err, "invalid severity found")
		return err
	}
	return nil
}
