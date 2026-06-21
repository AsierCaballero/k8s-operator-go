#!/bin/bash
set -euo pipefail

TOOLS_DIR="${1:-bin}"

echo "Installing controller-gen..."
GOBIN=$(pwd)/$TOOLS_DIR go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.15.0

echo "Installing setup-envtest..."
GOBIN=$(pwd)/$TOOLS_DIR go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

echo "Installing golangci-lint..."
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$TOOLS_DIR" v1.57.2

echo "Tools installed in $TOOLS_DIR/"
echo "  - controller-gen"
echo "  - setup-envtest"
echo "  - golangci-lint"
