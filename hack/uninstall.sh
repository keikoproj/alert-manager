#!/bin/bash
set -e

# Script to uninstall alert-manager from a Kubernetes cluster
# 
# Usage:
#   ./uninstall.sh [namespace]
#
# Arguments:
#   namespace - Kubernetes namespace where alert-manager is installed (default: alert-manager-system)

# Display usage information
function usage() {
  echo "Usage: $0 [namespace]"
  echo ""
  echo "Arguments:"
  echo "  namespace - Kubernetes namespace where alert-manager is installed (default: alert-manager-system)"
  echo ""
  echo "Example:"
  echo "  $0 alert-manager-system"
  exit 1
}

# Set variables with defaults
NAMESPACE=${1:-alert-manager-system}

# Verify kubectl is installed and configured
if ! command -v kubectl &> /dev/null; then
  echo "Error: kubectl is not installed or not in PATH."
  echo "Please install kubectl and try again."
  exit 1
fi

# Check if current kubectl context is pointing to the correct cluster
CURRENT_CONTEXT=$(kubectl config current-context 2>/dev/null || echo "none")
echo "Current kubectl context: $CURRENT_CONTEXT"
read -p "Is this the correct Kubernetes cluster? (y/n): " CORRECT_CLUSTER
if [[ ! $CORRECT_CLUSTER =~ ^[Yy]$ ]]; then
  echo "Please set the correct kubectl context and try again."
  exit 1
fi

echo "This will uninstall alert-manager from namespace: $NAMESPACE"
echo "WARNING: This action will also delete all alerts managed by alert-manager."
read -p "Continue with uninstallation? (y/n): " CONTINUE
if [[ ! $CONTINUE =~ ^[Yy]$ ]]; then
  echo "Uninstallation aborted."
  exit 1
fi

echo "---"
echo "Step 1: Deleting all WavefrontAlerts and AlertsConfigs..."
# We'll try to delete CRs from all namespaces but continue if they don't exist
kubectl delete wavefrontalerts --all --all-namespaces 2>/dev/null || true
kubectl delete alertsconfigs --all --all-namespaces 2>/dev/null || true

echo "---"
echo "Step 2: Removing alert-manager controller and resources..."
# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# Delete the deployment
kubectl delete -f $SCRIPT_DIR/alert-manager.yaml 2>/dev/null || true

echo "---"
echo "Step 3: Removing CRDs..."
kubectl delete crd wavefrontalerts.alertmanager.keikoproj.io 2>/dev/null || true
kubectl delete crd alertsconfigs.alertmanager.keikoproj.io 2>/dev/null || true

echo "---"
echo "Step 4: Removing namespace $NAMESPACE..."
kubectl delete namespace $NAMESPACE 2>/dev/null || true

echo "---"
echo "Uninstallation complete!"
echo "If you need to reinstall alert-manager, you can use: ./hack/install.sh"
