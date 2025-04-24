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

package common_test

import (
	"context"
	"fmt"
	"testing"

	"path/filepath"
	"runtime"

	"github.com/golang/mock/gomock"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/internal/controllers"
	"github.com/keikoproj/alert-manager/internal/controllers/common"
	mock_wavefront "github.com/keikoproj/alert-manager/internal/controllers/mocks"
	"github.com/keikoproj/alert-manager/pkg/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var cfg *rest.Config
var k8sClient client.Client
var k8sCl k8s.Client
var testEnv *envtest.Environment
var mockWavefront *mock_wavefront.MockInterface
var mgrCtx context.Context

// https://github.com/kubernetes-sigs/controller-runtime/issues/1571
var cancelFunc context.CancelFunc

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("../../../", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		BinaryAssetsDirectory: filepath.Join("../../../", "bin", "k8s",
			fmt.Sprintf("%s-%s-%s", "1.28.0", runtime.GOOS, runtime.GOARCH)),
	}

	var err error
	// Only proceed with controller tests if we can successfully start the test environment
	By("starting the test environment")
	cfg, err = testEnv.Start()
	if err != nil {
		// Skip the test environment setup but don't fail the tests
		// This allows the non-controller tests to run
		Skip(fmt.Sprintf("Error starting test environment: %v", err))
		return
	}

	Expect(cfg).NotTo(BeNil())

	err = alertmanagerv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	cl, err := kubernetes.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

	//For wavefront mock
	mockCtrl := gomock.NewController(GinkgoT())
	defer mockCtrl.Finish()
	mockWavefront = mock_wavefront.NewMockInterface(mockCtrl)

	// Setup default mock behaviors for all Wavefront API calls
	mockWavefront.EXPECT().CreateAlert(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, alert interface{}) error {
			// Just return success - we're not testing the actual API call
			return nil
		}).AnyTimes()

	mockWavefront.EXPECT().ReadAlert(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockWavefront.EXPECT().UpdateAlert(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockWavefront.EXPECT().DeleteAlert(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Create k8s client
	k8sCl = k8s.Client{
		Cl: cl,
	}
	commonClient := common.Client{
		Client:   k8sClient,
		Recorder: k8sCl.SetUpEventHandler(context.Background()),
	}

	err = (&controllers.WavefrontAlertReconciler{
		Client:          k8sManager.GetClient(),
		Scheme:          k8sManager.GetScheme(),
		CommonClient:    &commonClient,
		WavefrontClient: mockWavefront,
		Recorder:        k8sCl.SetUpEventHandler(context.Background()),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		mgrCtx, cancelFunc = context.WithCancel(context.Background())
		err = k8sManager.Start(mgrCtx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	if testEnv != nil && cfg != nil {
		cancelFunc()
		err := testEnv.Stop()
		Expect(err).NotTo(HaveOccurred())
	}
})
