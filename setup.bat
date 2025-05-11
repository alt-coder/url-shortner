@echo off
REM Check if Chocolatey is installed
if not exist "%ALLUSERSPROFILE%\chocolatey\choco.exe" (
    REM Install Chocolatey
    echo Installing Chocolatey...
    @"%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -InputFormat None -ExecutionPolicy Bypass -Command "iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))" && SET "PATH=%PATH%;%ALLUSERSPROFILE%\chocolatey\bin"
    if %ERRORLEVEL% equ 0 (
        echo Chocolatey installed successfully.
    ) else (
        echo Chocolatey installation failed.
        exit /b 1
    )
    REM Set environment variable
    setx PATH "%PATH%;%ALLUSERSPROFILE%\chocolatey\bin" /M
    echo Please restart your terminal after the script finishes.
)

REM Install Docker Desktop
echo Installing Docker Desktop...
choco install docker-desktop -y
if %ERRORLEVEL% neq 0 (
    echo Docker Desktop installation failed.
    exit /b 1
)

REM Check if Docker Desktop is running
echo Checking if Docker Desktop is running...
powershell -Command "if ((Get-Service 'Docker Desktop Service').Status -eq 'Running') {$true} else {$false}" > temp.txt
for /f "tokens=*" %%a in (temp.txt) do (
  set "docker_running=%%a"
)
del temp.txt
if "%docker_running%"=="False" (
    echo Docker Desktop is not running. Please start Docker Desktop and run this script again.
    exit /b 1
)

REM Install Go
echo Installing Go...
choco install golang -y
if %ERRORLEVEL% neq 0 (
    echo Go installation failed.
    exit /b 1
)

REM Install Kubectl
echo Installing Kubectl...
choco install kubernetes-cli -y
if %ERRORLEVEL% neq 0 (
    echo Kubectl installation failed.
    exit /b 1
)

REM Install Kind
echo Installing Kind...
choco install kind -y
if %ERRORLEVEL% neq 0 (
    echo Kind installation failed.
    exit /b 1
)

REM Install Make
echo Installing Make...
choco install make -y
if %ERRORLEVEL% neq 0 (
    echo Make installation failed.
    exit /b 1
)

REM Install Protoc
echo Installing Protoc...
choco install protoc -y
if %ERRORLEVEL% neq 0 (
    echo Protoc installation failed.
    exit /b 1
)

REM Install Skaffold
echo Installing Skaffold...
choco install skaffold -y
if %ERRORLEVEL% neq 0 (
    echo Skaffold installation failed.
    exit /b 1
)

REM Install Helm
echo Installing Helm...
choco install kubernetes-helm -y
if %ERRORLEVEL% neq 0 (
    echo Helm installation failed.
    exit /b 1
)

REM Add Go bin to PATH
echo Adding Go bin to PATH...
set "goBinPath=%USERPROFILE%\go\bin"
setx PATH "%PATH%;%goBinPath%" /M

REM Create Kind cluster
echo Creating Kind cluster...
kind create cluster
if %ERRORLEVEL% neq 0 (
    echo Kind cluster creation failed.
    exit /b 1
)

REM Set Docker context to Kind
echo Setting Docker context to Kind...
echo Setting Docker context to kind-kind
docker context use kind-kind

REM Add Helm repository
echo Adding Helm repository...
helm repo add bitnami https://charts.bitnami.com/bitnami
if %ERRORLEVEL% neq 0 (
    echo Adding Helm repository failed.
    exit /b 1
)

REM Update Helm repository
echo Updating Helm repository...
helm repo update
if %ERRORLEVEL% neq 0 (
    echo Updating Helm repository failed.
    exit /b 1
)

REM Install PostgreSQL
echo Installing PostgreSQL...
helm install postgresql bitnami/postgresql
if %ERRORLEVEL% neq 0 (
    echo Installing PostgreSQL failed.
    exit /b 1
)

REM Install ZooKeeper
echo Installing ZooKeeper...
helm install zookeeper bitnami/zookeeper
if %ERRORLEVEL% neq 0 (
    echo Installing ZooKeeper failed.
    exit /b 1
)

REM Install Redis
echo Installing Redis...
helm install redis bitnami/redis
if %ERRORLEVEL% neq 0 (
    echo Installing Redis failed.
    exit /b 1
)

echo Environment setup complete. Please restart your terminal.
pause