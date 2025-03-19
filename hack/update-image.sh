#!/bin/bash
set -e

# Script to update the controller image tag in alert-manager.yaml
# 
# Usage:
#   ./update-image.sh <new_tag>
#
# Arguments:
#   new_tag - New tag for the keikoproj/alert-manager image

# Display usage information
function usage() {
  echo "Usage: $0 <new_tag>"
  echo ""
  echo "Arguments:"
  echo "  new_tag - New tag for the keikoproj/alert-manager image"
  echo ""
  echo "Example:"
  echo "  $0 v0.5.0"
  exit 1
}

# Check for required parameters
if [ $# -ne 1 ]; then
  echo "Error: New image tag is required."
  usage
fi

NEW_TAG=$1

# Validate the tag format (basic validation)
if [[ ! $NEW_TAG =~ ^[a-zA-Z0-9\.\-_]+$ ]]; then
  echo "Error: Invalid tag format. Tags should only contain alphanumeric characters, dots, hyphens, and underscores."
  exit 1
fi

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
YAML_FILE="$SCRIPT_DIR/alert-manager.yaml"

# Check if the file exists
if [ ! -f "$YAML_FILE" ]; then
  echo "Error: $YAML_FILE does not exist."
  exit 1
fi

echo "Updating image tag to: keikoproj/alert-manager:$NEW_TAG"

# On macOS, sed -i requires an extension argument
if [[ "$OSTYPE" == "darwin"* ]]; then
  sed -i.bak "s|keikoproj/alert-manager:latest|keikoproj/alert-manager:$NEW_TAG|g" "$YAML_FILE"
  rm -f "$YAML_FILE.bak"
else
  sed -i "s|keikoproj/alert-manager:latest|keikoproj/alert-manager:$NEW_TAG|g" "$YAML_FILE"
fi

echo "Image tag updated successfully in $YAML_FILE"
echo "New image: keikoproj/alert-manager:$NEW_TAG"
