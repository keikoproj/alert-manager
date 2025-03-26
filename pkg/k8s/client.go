package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

type Client struct {
	Cl            kubernetes.Interface
	runtimeClient client.Client
}

// NewK8sSelfClientDoOrDie gets the new k8s go client
func NewK8sSelfClientDoOrDie() *Client {
	// For testing - return a mock client if TEST env is set
	if os.Getenv("TEST") == "true" {
		// Check if we're running in envtest (KUBEBUILDER_ASSETS will be set)
		if os.Getenv("KUBEBUILDER_ASSETS") != "" {
			// Use the envtest config
			testEnv := &envtest.Environment{
				CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
				ErrorIfCRDPathMissing: false,
			}

			cfg, err := testEnv.Start()
			if err != nil {
				fmt.Printf("Error starting test environment: %v\n", err)
				// Fall back to fake client
				return createFakeClient()
			}

			// Create clients using the envtest config
			cl, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				fmt.Printf("Error creating client with test config: %v\n", err)
				return createFakeClient()
			}

			scheme, err := alertmanagerv1alpha1.SchemeBuilder.Register(
				&alertmanagerv1alpha1.WavefrontAlert{},
				&alertmanagerv1alpha1.WavefrontAlertList{},
				&alertmanagerv1alpha1.AlertsConfig{},
				&alertmanagerv1alpha1.AlertsConfigList{}).Build()
			if err != nil {
				fmt.Printf("Error building scheme: %v\n", err)
				return createFakeClient()
			}

			dClient, err := client.New(cfg, client.Options{Scheme: scheme})
			if err != nil {
				fmt.Printf("Error creating runtime client: %v\n", err)
				return createFakeClient()
			}

			return &Client{
				Cl:            cl,
				runtimeClient: dClient,
			}
		}

		// If not in envtest, use fake client
		return createFakeClient()
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("THIS IS LOCAL")
		// Do i need to panic here?
		//How do i test this from local?
		//Lets get it from local config file
		var configErr error
		config, configErr = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if configErr != nil {
			panic(configErr)
		}
	}
	cl, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	//This is used for custom resources
	//https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client#New
	//Lets make sure we add all our custom types to the scheme
	scheme, err := alertmanagerv1alpha1.SchemeBuilder.Register(&alertmanagerv1alpha1.WavefrontAlert{}, &alertmanagerv1alpha1.WavefrontAlertList{}, &alertmanagerv1alpha1.AlertsConfig{}, &alertmanagerv1alpha1.AlertsConfigList{}).Build()
	if err != nil {
		panic(err)
	}
	dClient, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}

	k8sCl := &Client{
		Cl:            cl,
		runtimeClient: dClient,
	}
	return k8sCl
}

// createFakeClient creates a fake client for testing
func createFakeClient() *Client {
	// Create a fake kubernetes client for testing
	mockClient := fake.NewSimpleClientset()

	// Pre-create the configmap that will be requested by the config package
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alert-manager-configmap",
			Namespace: "alert-manager-system",
		},
		Data: map[string]string{
			"wavefront.api.token.secret": "wavefront-api-token",
			"wavefront.api.url":          "https://wavefront.example.com",
		},
	}

	// Add the configmap to the fake client
	_, err := mockClient.CoreV1().ConfigMaps("alert-manager-system").Create(
		context.Background(), configMap, metav1.CreateOptions{})
	if err != nil {
		// In testing, just print the error but don't panic
		fmt.Printf("Error creating mock configmap: %v\n", err)
	}

	return &Client{
		Cl:            mockClient,
		runtimeClient: nil,
	}
}

func (c *Client) ClientInterface() kubernetes.Interface {
	return c.Cl
}
