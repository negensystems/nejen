#!/bin/bash
# Install NEJEN pacman hooks to the system

set -euo pipefail

HOOKS_DIR="/etc/pacman.d/hooks"
REPO_HOOKS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../system/hooks" && pwd)"

echo "=== Deploying NEJEN Pacman Hooks ==="

# Ensure hooks directory exists
sudo mkdir -p "$HOOKS_DIR"



# Copy all NEJEN hooks and adjust binary paths for dev mode if needed
for hook in "$REPO_HOOKS_DIR"/*.hook; do
  hook_name=$(basename "$hook")
  echo "Installing $hook_name..."
  sudo cp "$hook" "$HOOKS_DIR/"
  if [[ "$hook_name" == "nejen-walker-restart.hook" && ! -f /usr/bin/nejen ]]; then
    dev_nejen_path="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)/bin/nejen"
    sudo sed -i "s|/usr/bin/nejen|$dev_nejen_path|g" "$HOOKS_DIR/$hook_name"
  fi
done

# Run them once to ensure current system is branded/ready
echo "Applying initial branding fixes..."
sudo sed -i 's/PRETTY_NAME="Arch Linux"/PRETTY_NAME="NEJEN Arch Linux"/g; s/NAME="Arch Linux"/NAME="NEJEN Arch Linux"/g' /usr/lib/os-release

echo "NEJEN Pacman hooks successfully deployed."
