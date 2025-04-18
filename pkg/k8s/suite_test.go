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
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	managerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	//logf "sigs.k8s.io/controller-runtime/pkg/log"
	//"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var cl Client

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8s Suite")
}

var _ = BeforeSuite(func() {
	//logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	// Set TEST=true environment variable for our client implementation
	os.Setenv("TEST", "true")

	By("bootstrapping test environment")

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	// Set KUBEBUILDER_ASSETS to let our client know we're using envtest
	if testEnv.BinaryAssetsDirectory != "" {
		os.Setenv("KUBEBUILDER_ASSETS", testEnv.BinaryAssetsDirectory)
	}

	err = managerv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	c, err := kubernetes.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())
	cl = Client{
		Cl:            c,
		runtimeClient: k8sClient,
	}
	Expect(cl).ToNot(BeNil())

	// Now also verify that our NewK8sSelfClientDoOrDie works properly
	// with the environment we've set up
	testClient := NewK8sSelfClientDoOrDie()
	Expect(testClient).ToNot(BeNil())
	Expect(testClient.Cl).ToNot(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	// Clean up environment variables
	os.Unsetenv("TEST")
	os.Unsetenv("KUBEBUILDER_ASSETS")

	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
