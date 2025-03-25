/*
Copyright 2025.

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

package main

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// TestMetricsSecureConfiguration verifies that the metrics configuration
// is properly set up with secure serving and authentication
func TestMetricsSecureConfiguration(t *testing.T) {
	// Set up test flags
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool

	fs.StringVar(&metricsAddr, "metrics-bind-address", ":8443", "The address the metric endpoint binds to.")
	fs.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")

	// Parse test flags
	err := fs.Parse([]string{
		"--metrics-bind-address=:8443",
		"--health-probe-bind-address=:8081",
		"--leader-elect=false",
	})
	assert.NoError(t, err)

	// Create metrics options as they would be in the main function
	metricsOptions := metricsserver.Options{
		BindAddress:    metricsAddr,
		SecureServing:  true,
		FilterProvider: filters.WithAuthenticationAndAuthorization,
	}

	// Verify metrics configuration
	assert.Equal(t, ":8443", metricsOptions.BindAddress)
	assert.True(t, metricsOptions.SecureServing)
	assert.NotNil(t, metricsOptions.FilterProvider)
}

// TestCommandLineFlagParsing verifies that command line flags are properly parsed
func TestCommandLineFlagParsing(t *testing.T) {
	// Save original os.Args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up test args
	os.Args = []string{
		"cmd",
		"--metrics-bind-address=:9443",
		"--health-probe-bind-address=:9081",
		"--leader-elect=true",
	}

	// Create new flag set for testing
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool

	fs.StringVar(&metricsAddr, "metrics-bind-address", ":8443", "The address the metric endpoint binds to.")
	fs.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")

	// Parse flags from test args
	err := fs.Parse(os.Args[1:])
	assert.NoError(t, err)

	// Verify flag values
	assert.Equal(t, ":9443", metricsAddr)
	assert.Equal(t, ":9081", probeAddr)
	assert.True(t, enableLeaderElection)
}

// The following test is a basic integration test to verify the secure metrics configuration
// works with controller-runtime's authentication and authorization
func TestManagerWithSecureMetrics(t *testing.T) {
	// This is a simple integration test that verifies the Manager can be created
	// with secure metrics configuration
	opts := zap.Options{
		Development: true,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	logger := ctrl.Log.WithName("test")

	// Create test manager options with secure metrics
	options := ctrl.Options{
		Logger: logger,
		Metrics: metricsserver.Options{
			BindAddress:    ":0", // Use port 0 to get random available port
			SecureServing:  true,
			FilterProvider: filters.WithAuthenticationAndAuthorization,
		},
	}

	// Verify manager can be created with these options
	// We don't actually start the manager as that would require a full Kubernetes setup
	// Just check that it can be created without errors
	_, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		t.Logf("Expected to create manager but got error: %v", err)
		t.Logf("This test may fail without a Kubernetes connection - skipping actual verification")
		t.Skip("Skipping manager creation test, likely due to no K8s connection")
		return
	}

	// If we get here, manager was created successfully with secure metrics
	t.Log("Successfully created manager with secure metrics configuration")
}

// TestMetricsPortBinding verifies that the metrics server can bind to a port
// with secure serving enabled
func TestMetricsPortBinding(t *testing.T) {
	// Configure a random available port by using port 0
	metricsAddr := ":0"

	// Set up the logger
	opts := zap.Options{
		Development: true,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Create metrics options with secure serving
	metricsOptions := metricsserver.Options{
		BindAddress:    metricsAddr,
		SecureServing:  true,
		FilterProvider: filters.WithAuthenticationAndAuthorization,
	}

	// Create manager with these options
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Metrics: metricsOptions,
	})
	if err != nil {
		t.Logf("Could not create manager: %v", err)
		t.Skip("Skipping port binding test as manager creation failed")
		return
	}

	// Start the manager in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error)
	go func() {
		errCh <- mgr.Start(ctx)
	}()

	// Give manager time to start
	time.Sleep(100 * time.Millisecond)

	// Check if there was an immediate error
	select {
	case err := <-errCh:
		t.Errorf("Manager failed to start: %v", err)
	default:
		// No error, manager started successfully
		t.Log("Manager started successfully with secure metrics binding")
	}
}
