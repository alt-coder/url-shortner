# url-shortner

## Description

This project is a simple URL shortener service.

## Design

The URL shortener service is designed as a standalone application that can be deployed in a containerized environment (e.g., Kubernetes).

**Core Components:**

* **gRPC Service:** The core logic for URL shortening, user management, and API key handling is implemented as a gRPC service. This allows for efficient inter-service communication if the project were to expand. The protobuf definitions can be found in [`url-shortener/proto/`](url-shortener/proto/).
* **gRPC Gateway:** To provide a user-friendly RESTful API, a gRPC gateway is used. It translates HTTP/JSON requests from clients into gRPC requests for the backend service.
* **PostgreSQL Database:** User data, URL mappings (long URL to short ID), and API keys are stored in a PostgreSQL database. The schema is managed using GORM auto-migration.
* **Zookeeper:** A distributed counter is implemented using Zookeeper. This counter is used to generate unique, sequential IDs that are then base62 encoded to create the short URL slugs. This approach ensures uniqueness even if multiple instances of the service are running.
* **Redis Client:** A Redis client is initialized and connected as part of the service setup ([`base/go/redis.go`](base/go/redis.go:1)). However, for demonstration purposes, caching of frequently accessed short URLs (to reduce database load) is **intentionally not implemented**.

**Workflow (URL Shortening):**

1. A user sends an HTTP POST request to the `/shorten` endpoint with their API key and the long URL.
2. The gRPC Gateway receives the request and forwards it to the gRPC `ShortenURL` method.
3. The service validates the API key against the PostgreSQL database.
4. If valid, the service requests a new unique ID from the Zookeeper-based distributed counter.
5. The ID is base62 encoded to create a short URL slug.
6. The mapping between the short URL slug and the original long URL is stored in PostgreSQL.
7. The short URL is returned to the user.

**Workflow (URL Redirection):**

1. A user accesses a short URL like `http://localhost:8081/d/{short_url}`.
2. The gRPC Gateway routes this to a custom handler which calls the gRPC `GetURL` method.
3. The service queries PostgreSQL for the long URL associated with the given short URL slug.
4. If found, the service returns the long URL, and the user is redirected.

**Intentionally Omitted Features (for simplicity/demonstration):**

* **Separate API Server Implementation:** The current API is exposed via the gRPC gateway. A more complex system might have a dedicated API server.
* **Rate Limiting:** No rate limiting is implemented on API requests. In a production system, this would be crucial to prevent abuse.
* **Multiple API Keys per User:** Each user currently has a single API key. A more advanced system might allow users to generate and manage multiple API keys.
* **Redis Caching:** While the Redis client ([`base/go/redis.go`](base/go/redis.go:1)) is set up and connected, it is not actively used for caching URL lookups. This was an intentional decision for this demonstration to keep the core logic simpler. In a high-traffic scenario, caching would significantly improve performance and reduce database load.

## Prerequisites

* None (Docker and Kind are handled by the setup script and Skaffold)

## Setup

### Running the Setup Script

The setup script automates the process of building and deploying the application. It is recommended to run the script as an administrator. There are two setup scripts, one for Linux/macOS (`setup.sh`) and one for Windows (`setup.bat`).

#### Linux/macOS

```bash
sudo ./setup.sh
```

#### Windows

Run the following in Elevated mode.

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
cd url-shortener
skaffold run 
# sudo skaffold run
```

This command will:

1. Build the Docker image.
2. Deploy the application to the Kind cluster.
3. Automatically update the cluster when changes are made to the code.

### Port Forwarding

To access the application from your local machine, you can use `kubectl` to forward the port:

```bash
kubectl port-forward service/url-shortener 8081:8081
```

## API Details

The URL shortener service exposes the following API endpoints:

### Shorten URL

* Endpoint: `POST /shorten`

* Request Body:
  
  ```json
  {
    "long_url": "https://www.example.com",
    "api_key": "YOUR_API_KEY"
  }
  ```

* Curl Command:
  
  ```bash
  curl -X POST -d '{"long_url": "https://www.example.com", "api_key": "YOUR_API_KEY"}' http://localhost:8081/shorten
  ```

* Response:
  
  ```json
  {
    "short_url": "shortened_url_code"
  }
  ```

### Redirect to Long URL

* Endpoint: `GET /d/{short_url}`

* Example: `GET /d/shortened_url`

* Curl Command:
  
  ```bash
  curl http://localhost:8081/d/shortened_url
  ```

### Create User

* Endpoint: `POST /users`

* Request Body:
  
  ```json
  {
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com"
  }
  ```

* Curl Command:
  
  ```bash
  curl -X POST -d '{"first_name": "John", "last_name": "Doe", "email": "john.doe@example.com"}' http://localhost:8081/users
  ```

* Response:
  
  ```json
  {
    "user_id": "user123",
    "api_key": "YOUR_API_KEY"
  }
  ```

### Get Top Shortened Domains (Metrics)

* Endpoint: `GET /metrics/top_domains`

* Description: Returns the top 3 domain names that have been shortened the most number of times.

* Curl Command:
  
  ```bash
  curl http://localhost:8081/metrics/top_domains
  ```

* Response:
  
  ```json
  {
    "top_domains": [
      {
        "domain": "udemy.com",
        "count": "6"
      },
      {
        "domain": "youtube.com",
        "count": "4"
      },
      {
        "domain": "wikipedia.com",
        "count": "2"
      }
    ]
  }
  ```

### Fetch API Key

* Endpoint: `GET /api_key/{email}`

* Example: `GET /api_key/john.doe@example.com`

* Curl Command:
  
  ```bash
  curl http://localhost:8081/api_key/john.doe@example.com
  ```

* Response:
  
  ```json
  {
    "api_key": "YOUR_API_KEY"
  }
  ```