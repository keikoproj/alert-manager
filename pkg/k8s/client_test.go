package k8s

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// SetupTestEnv initializes and returns a new test environment and config for testing
func SetupTestEnv() (*envtest.Environment, *rest.Config) {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
	}

	cfg, err := testEnv.Start()
	if err != nil {
		panic(fmt.Sprintf("Failed to start testEnv: %v", err))
	}

	return testEnv, cfg
}

// TeardownTestEnv stops the test environment
func TeardownTestEnv(testEnv *envtest.Environment) {
	if err := testEnv.Stop(); err != nil {
		fmt.Printf("Failed to stop testEnv: %v\n", err)
	}
}
