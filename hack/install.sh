#!/bin/bash
set -e

# Script to install alert-manager in a Kubernetes cluster
# 
# Usage:
#   ./install.sh <namespace> <monitoring_backend_url> <api_token> [--no-wait]
#
# Arguments:
#   namespace - Kubernetes namespace to install alert-manager (default: alert-manager-system)
#   monitoring_backend_url - URL of the monitoring backend (e.g. wavefront.example.com)
#   api_token - API token for monitoring backend (optional)
#   --no-wait - Skip waiting for deployment to be ready (optional)

# Display usage information
function usage() {
  echo "Usage: $0 <namespace> <monitoring_backend_url> <api_token> [--no-wait]"
  echo ""
  echo "Arguments:"
  echo "  namespace              - Kubernetes namespace where alert-manager will be installed (default: alert-manager-system)"
  echo "  monitoring_backend_url - Monitoring backend URL (e.g., wavefront.example.com)"
  echo "  api_token              - API token for the monitoring backend"
  echo "  --no-wait              - Skip waiting for deployment to be ready"
  echo ""
  echo "Example:"
  echo "  $0 alert-manager-system wavefront.example.com abcde-12345-fghij-67890"
  exit 1
}

# Check for minimum required parameters
if [ $# -lt 2 ]; then
  echo "Error: Missing required parameters."
  usage
fi

# Set variables with defaults
NAMESPACE=${1:-alert-manager-system}
MONITORING_URL=${2:-""}
API_TOKEN=${3:-""}
NO_WAIT="false"

# Check for --no-wait flag (can be in any position)
for arg in "$@"; do
  if [ "$arg" == "--no-wait" ]; then
    NO_WAIT="true"
  fi
done

# Validate parameters
if [ -z "$MONITORING_URL" ]; then
  echo "Error: Monitoring backend URL cannot be empty."
  usage
fi

if [ -z "$API_TOKEN" ]; then
  echo "Warning: API token not provided. You will need to create the secret manually."
  read -p "Continue anyway? (y/n): " CONTINUE
  if [[ ! $CONTINUE =~ ^[Yy]$ ]]; then
    echo "Installation aborted."
    exit 1
  fi
fi

# Verify kubectl is installed and configured
if ! command -v kubectl &> /dev/null; then
  echo "Error: kubectl is not installed or not in PATH."
  echo "Please install kubectl and try again."
  exit 1
fi

# Check if the current kubectl context is valid
if ! kubectl config current-context &> /dev/null; then
  echo "Error: No active Kubernetes context found."
  echo "Please set up your kubeconfig and try again."
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

echo "Installing alert-manager in namespace: $NAMESPACE with monitoring backend: $MONITORING_URL"
echo "---"

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# Create a temporary directory for modified manifests
TEMP_DIR=$(mktemp -d)
echo "Created temporary directory for manifests: $TEMP_DIR"

# Copy and modify the alert-manager.yaml manifest
echo "Preparing manifests with your configuration..."
cp $SCRIPT_DIR/alert-manager.yaml $TEMP_DIR/alert-manager.yaml

# Replace the namespace in the manifest if different from default
if [ "$NAMESPACE" != "alert-manager-system" ]; then
  echo "Setting custom namespace: $NAMESPACE"
  sed -i.bak "s/namespace: alert-manager-system/namespace: $NAMESPACE/g" $TEMP_DIR/alert-manager.yaml
  rm -f $TEMP_DIR/alert-manager.yaml.bak
fi

# Replace the monitoring URL in the ConfigMap
echo "Setting monitoring backend URL: $MONITORING_URL"
sed -i.bak "s|REPLACE_MONITORING_URL|$MONITORING_URL|g" $TEMP_DIR/alert-manager.yaml
rm -f $TEMP_DIR/alert-manager.yaml.bak

# Copy and modify the additional ConfigMap needed by the controller
echo "Preparing additional ConfigMap..."
cp $SCRIPT_DIR/alert-manager-configmap.yaml $TEMP_DIR/alert-manager-configmap.yaml
sed -i.bak "s|REPLACE_MONITORING_URL|$MONITORING_URL|g" $TEMP_DIR/alert-manager-configmap.yaml
rm -f $TEMP_DIR/alert-manager-configmap.yaml.bak

# If namespace is different from default, update it in the ConfigMap
if [ "$NAMESPACE" != "alert-manager-system" ]; then
  sed -i.bak "s/namespace: alert-manager-system/namespace: $NAMESPACE/g" $TEMP_DIR/alert-manager-configmap.yaml
  rm -f $TEMP_DIR/alert-manager-configmap.yaml.bak
fi

# Apply the main manifest
echo "Applying alert-manager manifests..."
kubectl apply -f $TEMP_DIR/alert-manager.yaml

# Apply the additional ConfigMap
echo "Applying additional ConfigMap..."
kubectl apply -f $TEMP_DIR/alert-manager-configmap.yaml

# Create API token secret if provided
if [ ! -z "$API_TOKEN" ]; then
  echo "Creating API token secrets..."
  
  # Copy the original credentials template
  cp $SCRIPT_DIR/alert-manager-credentials.yaml $TEMP_DIR/alert-manager-credentials.yaml
  
  # Update the namespace if needed
  if [ "$NAMESPACE" != "alert-manager-system" ]; then
    sed -i.bak "s/namespace: alert-manager-system/namespace: $NAMESPACE/g" $TEMP_DIR/alert-manager-credentials.yaml
    rm -f $TEMP_DIR/alert-manager-credentials.yaml.bak
  fi
  
  # Set the API token
  sed -i.bak "s|YOUR_API_TOKEN_HERE|$API_TOKEN|g" $TEMP_DIR/alert-manager-credentials.yaml
  rm -f $TEMP_DIR/alert-manager-credentials.yaml.bak
  
  # Apply the credentials
  kubectl apply -f $TEMP_DIR/alert-manager-credentials.yaml
  
  # Copy the wavefront-api-token template
  cp $SCRIPT_DIR/wavefront-api-token.yaml $TEMP_DIR/wavefront-api-token.yaml
  
  # Update the namespace if needed
  if [ "$NAMESPACE" != "alert-manager-system" ]; then
    sed -i.bak "s/namespace: alert-manager-system/namespace: $NAMESPACE/g" $TEMP_DIR/wavefront-api-token.yaml
    rm -f $TEMP_DIR/wavefront-api-token.yaml.bak
  fi
  
  # Set the API token
  sed -i.bak "s|YOUR_API_TOKEN_HERE|$API_TOKEN|g" $TEMP_DIR/wavefront-api-token.yaml
  rm -f $TEMP_DIR/wavefront-api-token.yaml.bak
  
  # Apply the wavefront token secret
  kubectl apply -f $TEMP_DIR/wavefront-api-token.yaml
else
  echo "API token not provided. You'll need to create the secret manually with:"
  echo ""
  cat $SCRIPT_DIR/alert-manager-credentials.yaml | sed "s/YOUR_API_TOKEN_HERE/<your-actual-token>/g"
  echo ""
  echo "Then apply with: kubectl apply -f your-secret.yaml"
fi

# Apply RBAC patch for secrets access
echo "Applying RBAC patch for secrets access..."
cp $SCRIPT_DIR/alert-manager-rbac-patch.yaml $TEMP_DIR/alert-manager-rbac-patch.yaml

# Update the namespace if needed
if [ "$NAMESPACE" != "alert-manager-system" ]; then
  sed -i.bak "s/namespace: alert-manager-system/namespace: $NAMESPACE/g" $TEMP_DIR/alert-manager-rbac-patch.yaml
  rm -f $TEMP_DIR/alert-manager-rbac-patch.yaml.bak
fi

kubectl apply -f $TEMP_DIR/alert-manager-rbac-patch.yaml

echo "---"
echo "Cleaning up temporary files..."
rm -rf $TEMP_DIR

echo "---"
if [ "$NO_WAIT" != "true" ]; then
  echo "Waiting for alert-manager controller to start (timeout: 300s)..."
  echo "Note: This may take longer in environments where image pulling is slow"
  kubectl rollout status deployment/alert-manager-controller-manager -n $NAMESPACE --timeout=300s || {
    echo "Note: Controller deployment is taking longer than expected, but installation completed."
    echo "You can check its status with: kubectl get pods -n $NAMESPACE"
  }
else
  echo "Skipping wait for deployment to be ready (--no-wait specified)"
fi

echo "---"
echo "Installation complete!"
echo "You can verify the installation by running:"
echo "  kubectl get pods -n $NAMESPACE"
echo ""
echo "To check the controller logs:"
echo "  kubectl logs -n $NAMESPACE deployment/alert-manager-controller-manager"
echo ""
echo "To create your first alert, apply the sample WavefrontAlert CR:"
echo "  kubectl apply -f $SCRIPT_DIR/sample-wavefront-alert.yaml"

echo "Done!"
