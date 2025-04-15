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
