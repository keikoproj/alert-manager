module github.com/keikoproj/alert-manager

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/google/uuid v1.1.1
	github.com/keikoproj/manager v0.0.0-20200414065656-d0d64621fb96
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.7.2
)