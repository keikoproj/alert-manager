package common_test

import (
	"context"

	"github.com/golang/mock/gomock"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/internal/controllers"
	"github.com/keikoproj/alert-manager/internal/controllers/common"
	mock_wavefront "github.com/keikoproj/alert-manager/internal/controllers/mocks"
	"github.com/keikoproj/alert-manager/pkg/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"

	"path/filepath"
	"testing"

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
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = alertmanagerv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

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
		Expect(err).ToNot(HaveOccurred())
	}()

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancelFunc()

	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
