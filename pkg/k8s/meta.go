package k8s

import (
	"context"
	"github.com/keikoproj/manager/pkg/log"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

//SetUpEventHandler sets up event handler with client-go recorder instead of creating events directly
func (c *Client) SetUpEventHandler(ctx context.Context) record.EventRecorder {
	log := log.Logger(ctx, "pkg.k8s", "meta", "SetUpEventHandler")
	//This was re-written based on job-controller in kubernetes repo
	//For more info refer: https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/job/job_controller.go
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: c.cl.CoreV1().Events("")})
	log.V(1).Info("Successfully added event broadcaster")
	return eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "manager"})
}

func (c *Client) GetConfigMap(ctx context.Context, ns string, name string) *v1.ConfigMap {
	log := log.Logger(ctx, "k8s", "client", "GetConfigMap")
	log.WithValues("namespace", ns)
	log.Info("Retrieving config map")
	res, err := c.cl.CoreV1().ConfigMaps(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "unable to get config map")
		panic(err)
	}

	return res
}

// GetConfigMapInformer returns shared informer for given config map
//func GetConfigMapInformer(ctx context.Context, nsName string, cmName string) cache.SharedIndexInformer {
//	log := log.Logger(context.Background(), "pkg.k8s.client", "GetConfigMapInformer")
//
//	listOptions := func(options *metav1.ListOptions) {
//		options.FieldSelector = fmt.Sprintf("metadata.name=%s", cmName)
//	}
//
//	// default resync period 24 hours
//	cmInformer := clientv1.NewFilteredConfigMapInformer(NewK8sSelfClientDoOrDie().ClientInterface(), nsName, 24*time.Hour, cache.Indexers{}, listOptions)
//	log.V(1).Info("Successfully got config map informer")
//	return cmInformer
//}
