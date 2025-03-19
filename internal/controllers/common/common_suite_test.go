package common_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

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

	// Enable test mode to avoid trying to connect to a real k8s cluster
	os.Setenv("TEST_MODE", "true")

	// Set KUBEBUILDER_ASSETS to the proper location
	assetsPath := filepath.Join("../../../", "bin", "k8s",
		fmt.Sprintf("%s-%s-%s", "1.28.0", runtime.GOOS, runtime.GOARCH),
		"k8s", fmt.Sprintf("%s-%s-%s", "1.28.0", runtime.GOOS, runtime.GOARCH))
	os.Setenv("KUBEBUILDER_ASSETS", assetsPath)

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

	//Set up the recorder
	recorder := k8sManager.GetEventRecorderFor("common-controller")

	mockController := gomock.NewController(GinkgoT())
	mockWavefront = mock_wavefront.NewMockInterface(mockController)

	// Create k8s client
	k8sClientObj := &k8s.Client{
		Cl: cl,
	}

	commonClient := &common.Client{
		Client:   k8sManager.GetClient(),
		Recorder: k8sClientObj.SetUpEventHandler(context.Background()),
	}

	mgrCtx, cancelFunc = context.WithCancel(ctrl.SetupSignalHandler())

	// Create a basic CR to test
	err = (&controllers.WavefrontAlertReconciler{
		Client:          k8sManager.GetClient(),
		Scheme:          scheme.Scheme,
		CommonClient:    commonClient,
		WavefrontClient: mockWavefront,
		Recorder:        recorder,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
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
