#!/bin/bash
set -e

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

echo "Setting up development environment..."

# Check/Install Go
if ! command_exists go; then
    echo "Installing Go..."
    wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    rm go1.21.5.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    source ~/.bashrc
fi

# Check/Install kubectl
if ! command_exists kubectl; then
    echo "Installing kubectl..."
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
    rm kubectl
fi

# Check/Install restic
if ! command_exists restic; then
    echo "Installing restic..."
    sudo apt-get update
    sudo apt-get install -y restic
fi

# Create namespace if it doesn't exist
if ! kubectl get namespace cnpg-system >/dev/null 2>&1; then
    echo "Creating cnpg-system namespace..."
    kubectl create namespace cnpg-system
fi

# Install CloudNative PostgreSQL operator
echo "Installing CloudNative PostgreSQL operator v1.26.1..."
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.26/releases/cnpg-1.26.1.yaml --server-side --force-conflicts

echo "Development environment setup complete!"
echo "Next steps:"
echo "1. Run: source ~/.bashrc"
echo "2. Build the plugin: make build"
echo "3. Run tests: make test"
