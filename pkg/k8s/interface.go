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

package k8s

import (
	"context"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

// Interface defines required functions to be implemented by receivers
type Interface interface {
	SetUpEventHandler(ctx context.Context) record.EventRecorder
	GetConfigMap(ctx context.Context, ns string, name string) *v1.ConfigMap
	GetK8sSecret(ctx context.Context, name string, ns string) (*v1.Secret, error)
}
