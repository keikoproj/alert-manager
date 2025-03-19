package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/keikoproj/alert-manager/internal/config/common"
	"github.com/keikoproj/alert-manager/pkg/k8s"
	"github.com/keikoproj/alert-manager/pkg/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

var (
	Props *Properties
)

type Properties struct {
	wavefrontAPITokenSecretName string
	wavefrontAPIUrl             string
}

func init() {
	logger := log.Logger(context.Background(), "config", "properties", "init")
	// bypass Kubernetes client initialization when running tests
	if strings.ToLower(os.Getenv("TEST_MODE")) == "true" {
		logger.Info("TEST_MODE=true, skipping Kubernetes client initialization")
		// Set default test properties
		if err := LoadProperties("test", nil); err != nil {
			logger.Error(err, "failed to load test properties")
			return
		}
		logger.Info("Loaded properties in init func for tests")
		return
	}

	// Check if KUBECONFIG is set and bypass client creation if running tests
	if len(os.Getenv("KUBECONFIG")) == 0 && strings.ToLower(os.Getenv("LOCAL")) != "true" {
		logger.Info("KUBECONFIG not set, skipping Kubernetes client initialization")
		// Set default test properties
		if err := LoadProperties("test", nil); err != nil {
			logger.Error(err, "failed to load test properties")
			return
		}
		logger.Info("Loaded properties in init func for tests")
		return
	}

	res := k8s.NewK8sSelfClientDoOrDie().GetConfigMap(context.Background(), common.AlertManagerNamespaceName, common.AlertManagerConfigMapName)

	// load properties into a global variable
	if err := LoadProperties("", res); err != nil {
		logger.Error(err, "failed to load properties")
		panic(err)
	}
	logger.Info("Loaded properties in init func")
}

func LoadProperties(env string, cm ...*v1.ConfigMap) error {
	logger := log.Logger(context.Background(), "internal.config.properties", "LoadProperties")
	Props = &Properties{}
	// for local testing
	if env != "" {
		return nil
	}

	if len(cm) == 0 || cm[0] == nil {
		logger.Error(fmt.Errorf("config map cannot be nil"), "config map cannot be nil")
		return fmt.Errorf("config map cannot be nil")
	}

	WavefrontAPITokenSecretName := cm[0].Data[common.WavefrontAPITokenK8sSecretName]
	if WavefrontAPITokenSecretName == "" {
		WavefrontAPITokenSecretName = "wavefront-api-token"
	}
	Props.wavefrontAPITokenSecretName = WavefrontAPITokenSecretName

	WavefrontAPIUrl := cm[0].Data[common.WavefrontAPIUrl]
	if WavefrontAPIUrl == "" {
		msg := "wavefront api url must be provided and should be in format "
		err := errors.New(msg)
		logger.Error(err, "unable to find wavefront api url in config map")
		return err
	}
	Props.wavefrontAPIUrl = WavefrontAPIUrl

	return nil
}

func (p *Properties) WavefrontAPITokenSecretName() string {
	return p.wavefrontAPITokenSecretName
}

func (p *Properties) WavefrontAPIUrl() string {
	return p.wavefrontAPIUrl
}

func RunConfigMapInformer(ctx context.Context) {
	logger := log.Logger(context.Background(), "internal.config.properties", "RunConfigMapInformer")
	cmInformer := k8s.GetConfigMapInformer(ctx, common.AlertManagerNamespaceName, common.AlertManagerConfigMapName)
	// AddEventHandler returns a handle and a registration error.
	// We don't need to use the handle as we run the informer for the lifetime of the context
	_, regErr := cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: updateProperties,
	})
	if regErr != nil {
		logger.Error(regErr, "Failed to register event handler")
		return
	}
	logger.Info("Starting config map informer")
	cmInformer.Run(ctx.Done())
	logger.Info("Cancelling config map informer")
}

func updateProperties(old, new interface{}) {
	logger := log.Logger(context.Background(), "internal.config.properties", "updateProperties")
	oldCM := old.(*v1.ConfigMap)
	newCM := new.(*v1.ConfigMap)
	if oldCM.ResourceVersion == newCM.ResourceVersion {
		return
	}
	logger.Info("Updating config map", "new revision ", newCM.ResourceVersion)
	if err := LoadProperties("", newCM); err != nil {
		logger.Error(err, "failed to update config map")
	}
}
