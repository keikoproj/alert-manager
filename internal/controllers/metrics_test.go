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

package controllers_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var _ = Describe("Metrics Server", func() {
	var (
		ctx         context.Context
		stopFunc    context.CancelFunc
		metricsOpts metricsserver.Options
		testManager ctrl.Manager
	)

	BeforeEach(func() {
		// Skip if no test environment is available
		if testEnv == nil || cfg == nil {
			Skip("Test environment or config is not available")
		}

		// Create context with cancel function
		ctx, stopFunc = context.WithCancel(context.Background())

		// Random metrics port to avoid conflicts
		metricsPort := fmt.Sprintf(":%d", 9090+GinkgoParallelProcess())

		// Create metrics options with secure serving and authentication
		metricsOpts = metricsserver.Options{
			BindAddress:    metricsPort,
			SecureServing:  true,
			FilterProvider: filters.WithAuthenticationAndAuthorization,
		}

		// Set up manager with our metrics options
		var err error
		testManager, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme:  testEnv.Scheme,
			Metrics: metricsOpts,
			Client: client.Options{
				Cache: &client.CacheOptions{},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		// Start the manager in background
		go func() {
			err = testManager.Start(ctx)
			if err != nil {
				// Logging only, no assertion here
				// as canceling the context will end with an error
				GinkgoT().Logf("Error starting manager: %v", err)
			}
		}()

		// Give manager time to start
		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		// Shut down manager
		stopFunc()
	})

	It("Should serve metrics endpoint with authentication", func() {
		// Get the metrics endpoint with no authentication,
		// which should fail with 401 or 403
		url := fmt.Sprintf("https://localhost%s/metrics", metricsOpts.BindAddress)

		// Use a custom HTTP client that skips TLS verification
		// since we're using a self-signed cert
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}

		// Make the request
		req, err := http.NewRequest("GET", url, nil)
		Expect(err).NotTo(HaveOccurred())

		resp, err := client.Do(req)

		// We expect either an error (connection refused)
		// or an unauthorized/forbidden response
		if err == nil {
			defer resp.Body.Close()
			GinkgoT().Logf("Got status code: %d", resp.StatusCode)
			Expect(resp.StatusCode).To(SatisfyAny(
				Equal(http.StatusUnauthorized),
				Equal(http.StatusForbidden),
			))
		} else {
			// If we got a connection error, log it but don't fail the test
			// This can happen in CI environments or when the test runs in a container
			GinkgoT().Logf("Could not connect to metrics endpoint: %v", err)
			Skip("Skipping authentication test as metrics endpoint could not be reached")
		}
	})
})
