#Requires -RunAsAdministrator

# Check if Chocolatey is installed
if (-not (Get-Command choco -ErrorAction SilentlyContinue)) {
    # Install Chocolatey
    Write-Host "Installing Chocolatey..."
    iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))
    if ($?) {
        Write-Host "Chocolatey installed successfully."
    } else {
        Write-Error "Chocolatey installation failed."
        exit 1
    }
    $env:PATH = "$env:PATH;%ALLUSERSPROFILE%\chocolatey\bin"
    [Environment]::SetEnvironmentVariable("Path", $env:PATH, "User")
    Write-Host "Please restart your terminal after the script finishes."
}

# Install Docker Desktop
Write-Host "Installing Docker Desktop..."
choco install docker-desktop -y
if (-not $?) {
    Write-Error "Docker Desktop installation failed."
    exit 1
}

# Check if Docker Desktop is running
Write-Host "Checking if Docker Desktop is running..."
if (-not (Get-Service "Docker Desktop Service" | Where-Object {$_.Status -eq "Running"})) {
    Write-Warning "Docker Desktop is not running. Please start Docker Desktop and run this script again."
    exit 1
}

# Install Go
Write-Host "Installing Go..."
choco install golang -y
if (-not $?) {
    Write-Error "Go installation failed."
    exit 1
}

# Install Kubectl
Write-Host "Installing Kubectl..."
choco install kubernetes-cli -y
if (-not $?) {
    Write-Error "Kubectl installation failed."
    exit 1
}

# Install Kind
Write-Host "Installing Kind..."
choco install kind -y
if (-not $?) {
    Write-Error "Kind installation failed."
    exit 1
}

# Install Make
Write-Host "Installing Make..."
choco install make -y
if (-not $?) {
    Write-Error "Make installation failed."
    exit 1
}

# Install Protoc
Write-Host "Installing Protoc..."
choco install protoc -y
if (-not $?) {
    Write-Error "Protoc installation failed."
    exit 1
}

# Install Skaffold
Write-Host "Installing Skaffold..."
choco install skaffold -y
if (-not $?) {
    Write-Error "Skaffold installation failed."
    exit 1
}

# Install Helm
Write-Host "Installing Helm..."
choco install kubernetes-helm -y
if (-not $?) {
    Write-Error "Helm installation failed."
    exit 1
}

# Add Go bin to PATH
$goBinPath = "$env:USERPROFILE\go\bin"
$env:PATH = "$env:PATH;$goBinPath"
[Environment]::SetEnvironmentVariable("Path", $env:PATH, "User")

# Create Kind cluster
Write-Host "Creating Kind cluster..."
kind create cluster
if (-not $?) {
    Write-Error "Kind cluster creation failed."
    exit 1
}

# Add Helm repository
Write-Host "Adding Helm repository..."
helm repo add bitnami https://charts.bitnami.com/bitnami
if (-not $?) {
    Write-Error "Adding Helm repository failed."
    exit 1
}

# Update Helm repository
Write-Host "Updating Helm repository..."
helm repo update
if (-not $?) {
    Write-Error "Updating Helm repository failed."
    exit 1
}

# Install PostgreSQL
Write-Host "Installing PostgreSQL..."
helm install postgresql bitnami/postgresql
if (-not $?) {
    Write-Error "Installing PostgreSQL failed."
    exit 1
}

# Install ZooKeeper
Write-Host "Installing ZooKeeper..."
helm install zookeeper bitnami/zookeeper
if (-not $?) {
    Write-Error "Installing ZooKeeper failed."
    exit 1
}

# Install Redis
Write-Host "Installing Redis..."
helm install redis bitnami/redis
if (-not $?) {
    Write-Error "Installing Redis failed."
    exit 1
}

Write-Host "Environment setup complete. Please restart your terminal."
