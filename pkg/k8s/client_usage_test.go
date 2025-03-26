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
