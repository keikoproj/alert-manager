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

package common

// Global constants

const (

	// AlertManagerNamespaceName is the namespace name where alert-manager controllers are running
	AlertManagerNamespaceName = "alert-manager-system"

	// AlertManagerConfigMapName is the config map name for alert-manager namespace
	AlertManagerConfigMapName = "alert-manager-configmap"

	//WavefrontAPITokenK8sSecretName is the secret name where API token is stored in k8s namespace
	WavefrontAPITokenK8sSecretName = "wavefront.api.token.secret.name"

	//WavefrontAPIUrl is the address of wavefront api
	WavefrontAPIUrl = "wavefront.api.url"
)
