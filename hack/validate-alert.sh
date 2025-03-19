#!/bin/bash
set -e

# Script to validate alert YAML files
# 
# Usage:
#   ./validate-alert.sh <file.yaml>
#
# Arguments:
#   file.yaml - Path to the alert YAML file to validate

# Display usage information
function usage() {
  echo "Usage: $0 <file.yaml>"
  echo ""
  echo "Arguments:"
  echo "  file.yaml - Path to the alert YAML file to validate"
  echo ""
  echo "Example:"
  echo "  $0 examples/sample-wavefront-alert.yaml"
  exit 1
}

# Check for required parameters
if [ $# -ne 1 ]; then
  echo "Error: Alert YAML file is required."
  usage
fi

FILE=$1

# Check if the file exists
if [ ! -f "$FILE" ]; then
  echo "Error: File not found: $FILE"
  exit 1
fi

echo "Validating alert file: $FILE"
echo "---"

# Basic YAML syntax validation
echo "Step 1: Validating YAML syntax..."
kubectl apply --dry-run=client -f "$FILE" > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "❌ YAML validation failed. File contains syntax errors:"
  kubectl apply --dry-run=client -f "$FILE"
  exit 1
else
  echo "✅ YAML syntax is valid."
fi

# Validate specific fields depending on the resource type
echo "---"
echo "Step 2: Validating alert configuration..."

# Extract kind and name for better messages
KIND=$(grep -E "^kind:" "$FILE" | awk '{print $2}')
NAME=$(grep -E "^  name:" "$FILE" | head -1 | awk '{print $2}')

if [ "$KIND" == "WavefrontAlert" ]; then
  # Check for required fields in WavefrontAlert
  MISSING=""
  
  # Check alertName
  if ! grep -q "alertName:" "$FILE"; then
    MISSING="$MISSING\n  - alertName is required"
  fi
  
  # Check condition
  if ! grep -q "condition:" "$FILE"; then
    MISSING="$MISSING\n  - condition is required"
  fi
  
  # Check severity format
  if grep -q "severity:" "$FILE"; then
    SEVERITY=$(grep "severity:" "$FILE" | awk '{print $2}' | tr -d '"' | tr -d "'")
    if [[ ! "$SEVERITY" =~ ^(info|smoke|warn|severe)$ ]]; then
      MISSING="$MISSING\n  - severity must be one of: info, smoke, warn, severe (was: $SEVERITY)"
    fi
  else
    MISSING="$MISSING\n  - severity is required"
  fi
  
  if [ -n "$MISSING" ]; then
    echo "❌ WavefrontAlert '$NAME' validation failed:"
    echo -e "$MISSING"
    exit 1
  else
    echo "✅ WavefrontAlert '$NAME' configuration is valid."
  fi
elif [ "$KIND" == "AlertsConfig" ]; then
  # Check for required fields in AlertsConfig
  MISSING=""
  
  # Check globalGVK
  if ! grep -q "globalGVK:" "$FILE"; then
    MISSING="$MISSING\n  - globalGVK is required"
  else
    # Check group
    if ! grep -q "group:" "$FILE"; then
      MISSING="$MISSING\n  - globalGVK.group is required"
    fi
    
    # Check version
    if ! grep -q "version:" "$FILE"; then
      MISSING="$MISSING\n  - globalGVK.version is required"
    fi
    
    # Check kind
    if ! grep -q "kind:" "$FILE" | grep -v "^kind:"; then
      MISSING="$MISSING\n  - globalGVK.kind is required"
    fi
  fi
  
  # Check alerts section
  if ! grep -q "alerts:" "$FILE"; then
    MISSING="$MISSING\n  - alerts section is required"
  fi
  
  if [ -n "$MISSING" ]; then
    echo "❌ AlertsConfig '$NAME' validation failed:"
    echo -e "$MISSING"
    exit 1
  else
    echo "✅ AlertsConfig '$NAME' configuration is valid."
  fi
else
  echo "⚠️ Unknown resource kind: $KIND. Only basic YAML validation performed."
fi

echo "---"
echo "✅ Alert validation passed successfully."
