package k8s

import (
	"fmt"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	Cl            kubernetes.Interface
	runtimeClient client.Client
}

//NewK8sSelfClientDoOrDie gets the new k8s go client
func NewK8sSelfClientDoOrDie() *Client {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("THIS IS LOCAL")
		// Do i need to panic here?
		//How do i test this from local?
		//Lets get it from local config file
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
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

func (c *Client) ClientInterface() kubernetes.Interface {
	return c.Cl
}
