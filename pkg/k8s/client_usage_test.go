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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("K8s Client No KUBECONFIG", func() {
	It("should create a client without requiring KUBECONFIG", func() {
		// The BeforeSuite in suite_test.go has already set up the test environment
		// and verified that NewK8sSelfClientDoOrDie works, but we'll confirm again here

		// Explicitly unset KUBECONFIG to verify it works without it
		GinkgoT().Setenv("KUBECONFIG", "")

		client := NewK8sSelfClientDoOrDie()
		Expect(client).NotTo(BeNil(), "Client should not be nil")
		Expect(client.Cl).NotTo(BeNil(), "Client interface should not be nil")

		// Verify we can actually use the client by listing namespaces
		// This will fail if the client is not properly configured
		_, err := client.Cl.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "Should be able to list namespaces")
	})
})
