#!/bin/bash

set -e

# Update package list
apt-get update

# Install prerequisites
apt-get install -y apt-transport-https ca-certificates curl

# Add Kubernetes apt repository
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /' | tee /etc/apt/sources.list.d/kubernetes.list

# Install kubectl
apt-get update
apt-get install -y kubectl

# Install Docker
apt-get update
apt-get install -y docker.io

# Install Kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.23.0/kind-linux-amd64
chmod +x ./kind
mv ./kind /usr/local/bin/kind

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install Skaffold
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
chmod +x skaffold
mv skaffold /usr/local/bin

# Create Kind cluster
kind create cluster

# Install Helm
helm repo add bitnami https://charts.bitnami.com/bitnami

# Update Helm repository
helm repo update

# Install PostgreSQL
helm install postgresql bitnami/postgresql

# Install ZooKeeper
helm install zookeeper bitnami/zookeeper

# Install Redis
helm install redis bitnami/redis

echo "Environment setup complete."