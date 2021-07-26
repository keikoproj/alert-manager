package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/keikoproj/alert-manager/internal/config/common"
	"github.com/keikoproj/alert-manager/pkg/k8s"
	"github.com/keikoproj/alert-manager/pkg/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"os"
)

var (
	Props *Properties
)

type Properties struct {
	wavefrontAPITokenSecretName string
	wavefrontAPIUrl             string
}

func init() {
	log := log.Logger(context.Background(), "internal.config.properties", "init")

	if os.Getenv("LOCAL") != "" {
		err := LoadProperties("LOCAL")
		if err != nil {
			log.Error(err, "failed to load properties for local environment")
			return
		}
		log.Info("Loaded properties in init func for tests")
		return
	}

	res := k8s.NewK8sSelfClientDoOrDie().GetConfigMap(context.Background(), common.AlertManagerNamespaceName, common.AlertManagerConfigMapName)

	// load properties into a global variable
	err := LoadProperties("", res)
	if err != nil {
		log.Error(err, "failed to load properties")
		panic(err)
	}
	log.Info("Loaded properties in init func")
}

func LoadProperties(env string, cm ...*v1.ConfigMap) error {
	log := log.Logger(context.Background(), "internal.config.properties", "LoadProperties")
	Props = &Properties{}
	// for local testing
	if env != "" {

		return nil
	}

	if len(cm) == 0 || cm[0] == nil {
		log.Error(fmt.Errorf("config map cannot be nil"), "config map cannot be nil")
		return fmt.Errorf("config map cannot be nil")
	}

	WavefrontAPITokenSecretName := cm[0].Data[common.WavefrontAPITokenK8sSecretName]
	if WavefrontAPITokenSecretName == "" {
		WavefrontAPITokenSecretName = "wavefront-api-token"
	}
	Props.wavefrontAPITokenSecretName = WavefrontAPITokenSecretName

	WavefrontAPIUrl := cm[0].Data[common.WavefrontAPIUrl]
	if WavefrontAPITokenSecretName == "" {
		msg := "wavefront api url must be provided and should be in format "
		err := errors.New(msg)
		log.Error(err, "unable to find wavefront api url in config map")
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
	log := log.Logger(context.Background(), "internal.config.properties", "RunConfigMapInformer")
	cmInformer := k8s.GetConfigMapInformer(ctx, common.AlertManagerNamespaceName, common.AlertManagerConfigMapName)
	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: updateProperties,
	},
	)
	log.Info("Starting config map informer")
	cmInformer.Run(ctx.Done())
	log.Info("Cancelling config map informer")
}

func updateProperties(old, new interface{}) {
	log := log.Logger(context.Background(), "internal.config.properties", "updateProperties")
	oldCM := old.(*v1.ConfigMap)
	newCM := new.(*v1.ConfigMap)
	if oldCM.ResourceVersion == newCM.ResourceVersion {
		return
	}
	log.Info("Updating config map", "new revision ", newCM.ResourceVersion)
	err := LoadProperties("", newCM)
	if err != nil {
		log.Error(err, "failed to update config map")
	}
}
