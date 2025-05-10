# Check if Chocolatey is installed
if (-not (Get-Command choco -ErrorAction SilentlyContinue)) {
    # Install Chocolatey
    Write-Host "Installing Chocolatey..."
    iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))
    $env:PATH = "$env:PATH;%ALLUSERSPROFILE%\chocolatey\bin"
    Write-Host "Please restart your terminal after the script finishes."
}

# Install Docker Desktop
Write-Host "Installing Docker Desktop..."
choco install docker-desktop -y

# Install Go
Write-Host "Installing Go..."
choco install golang -y

# Install Kubectl
Write-Host "Installing Kubectl..."
choco install kubernetes-cli -y

# Install Kind
Write-Host "Installing Kind..."
choco install kind -y

# Add Go bin to PATH
[Environment]::SetEnvironmentVariable("Path", [Environment]::GetEnvironmentVariable("Path", "User") + ";$env:USERPROFILE\go\bin", "User")

# Create Kind cluster
Write-Host "Creating Kind cluster..."
kind create cluster

# Install Helm
Write-Host "Installing Helm..."
choco install kubernetes-helm -y

# Add Helm repository
Write-Host "Adding Helm repository..."
helm repo add bitnami https://charts.bitnami.com/bitnami

# Update Helm repository
Write-Host "Updating Helm repository..."
helm repo update

# Install PostgreSQL
Write-Host "Installing PostgreSQL..."
helm install postgresql bitnami/postgresql

# Install ZooKeeper
Write-Host "Installing ZooKeeper..."
helm install zookeeper bitnami/zookeeper

# Install Redis
Write-Host "Installing Redis..."
helm install redis bitnami/redis
