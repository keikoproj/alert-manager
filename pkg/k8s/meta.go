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
	"fmt"
	"time"

	"github.com/keikoproj/alert-manager/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
	clientv1 "k8s.io/client-go/informers/core/v1"
)

// SetUpEventHandler sets up event handler with client-go recorder instead of creating events directly
func (c *Client) SetUpEventHandler(ctx context.Context) record.EventRecorder {
	logk := log.Logger(ctx, "pkg.k8s", "meta", "SetUpEventHandler")
	//This was re-written based on job-controller in kubernetes repo
	//For more info refer: https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/job/job_controller.go
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: c.Cl.CoreV1().Events("")})
	logk.V(1).Info("Successfully added event broadcaster")
	return eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "alert-manager-controller"})
}

func (c *Client) GetConfigMap(ctx context.Context, ns string, name string) *v1.ConfigMap {
	logk := log.Logger(ctx, "k8s", "client", "GetConfigMap")
	logk.WithValues("namespace", ns)
	logk.Info("Retrieving config map")
	res, err := c.Cl.CoreV1().ConfigMaps(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		logk.Error(err, "unable to get config map")
		panic(err)
	}

	return res
}

// GetConfigMapInformer returns shared informer for given config map
func GetConfigMapInformer(ctx context.Context, nsName string, cmName string) cache.SharedIndexInformer {
	logk := log.Logger(context.Background(), "pkg.k8s.client", "GetConfigMapInformer")

	listOptions := func(options *metav1.ListOptions) {
		options.FieldSelector = fmt.Sprintf("metadata.name=%s", cmName)
	}

	// default resync period 24 hours
	cmInformer := clientv1.NewFilteredConfigMapInformer(NewK8sSelfClientDoOrDie().ClientInterface(), nsName, 24*time.Hour, cache.Indexers{}, listOptions)
	logk.V(1).Info("Successfully got config map informer")
	return cmInformer
}

// GetK8sSecret function retrieves the secrets
func (c *Client) GetK8sSecret(ctx context.Context, name string, ns string) (*v1.Secret, error) {
	logk := log.Logger(ctx, "pkg.k8s", "meta", "GetK8sSecret")
	logk = logk.WithValues("secretName", name, "namespace", ns)
	logk.V(1).Info("Retrieving secret")
	secret, err := c.Cl.CoreV1().Secrets(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		logk.Error(err, "unable to retrieve secret", "name", name, "namespace", ns)
		return nil, err
	}
	logk.Info("secret found", "secret_name", secret.Name)

	return secret, nil
}
