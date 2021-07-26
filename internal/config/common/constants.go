package common

// Global constants

const (

	// AlertManagerNamespaceName is the namespace name where alert-manager controllers are running
	AlertManagerNamespaceName = "alert-manager-system"

	// AlertManagerConfigMapName is the config map name for alert-manager namespace
	AlertManagerConfigMapName = "alertmanager-v1alpha1-configmap"

	//WavefrontAPITokenK8sSecretName is the secret name where API token is stored in k8s namespace
	WavefrontAPITokenK8sSecretName = "wavefront.api.token.secret.name"

	//WavefrontAPIUrl is the address of wavefront api
	WavefrontAPIUrl = "wavefront.api.url"
)
