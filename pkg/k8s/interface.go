package k8s

import (
	"context"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

//Interface defines required functions to be implemented by receivers
type Interface interface {
	SetUpEventHandler(ctx context.Context) record.EventRecorder
	GetConfigMap(ctx context.Context, ns string, name string) *v1.ConfigMap
}
