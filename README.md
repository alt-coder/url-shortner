# url-shortner

## Description

This project is a simple URL shortener service.

## Prerequisites

*   None (Docker and Kind are handled by the setup script and Skaffold)

## Setup

### Running the Setup Script

The setup script automates the process of building and deploying the application. It is recommended to run the script as an administrator. There are two setup scripts, one for Linux/macOS (`setup.sh`) and one for Windows (`setup.bat`).

#### Linux/macOS

```bash
./setup.sh
```

#### Windows

```batch
setup.bat
```

The script will install necessary components for the project like kubectl, postgres, redis, zookeeper, kind cluster, etc.

## Running the Application

The application is automatically built and deployed to a Kind cluster using Skaffold.

### Using Skaffold

To run the application:
go to url-shortner where skaffold.yaml file resides and run the following
```bash
skaffold run
```

This command will:

1. Build the Docker image.
2. Deploy the application to the Kind cluster.
3. Automatically update the cluster when changes are made to the code.

### Port Forwarding

To access the application from your local machine, you can use `kubectl` to forward the port:

```bash
kubectl port-forward service/url-shortener 8081:8080
```

## API Details

The URL shortener service exposes the following API endpoints:

### Shorten URL

*   Endpoint: `POST /shorten`
*   Request Body:

    ```json
    {
      "long_url": "https://www.example.com",
      "api_key": "YOUR_API_KEY"
    }
    ```

*   Curl Command:

    ```bash
    curl -X POST -d '{"long_url": "https://www.example.com", "api_key": "YOUR_API_KEY"}' http://localhost:8081/shorten
    ```

*   Response:

    ```json
    {
      "short_url": "shortened_url_code"
    }
    ```

### Redirect to Long URL

*   Endpoint: `GET /d/{short_url}`
*   Example: `GET /d/shortened_url`
*   Curl Command:

    ```bash
    curl http://localhost:8081/d/shortened_url
    ```

### Create User

*   Endpoint: `POST /users`
*   Request Body:

    ```json
    {
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@example.com"
    }
    ```

*   Curl Command:

    ```bash
    curl -X POST -d '{"first_name": "John", "last_name": "Doe", "email": "john.doe@example.com"}' http://localhost:8081/users
    ```

*   Response:

    ```json
    {
      "user_id": "user123",
      "api_key": "YOUR_API_KEY"
    }
    ```

### Fetch API Key

*   Endpoint: `GET /api_key/{email}`
*   Example: `GET /api_key/john.doe@example.com`
*   Curl Command:

    ```bash
    curl http://localhost:8081/api_key/john.doe@example.com
    ```

*   Response:

    ```json
    {
      "api_key": "YOUR_API_KEY"
    }
    ```
    