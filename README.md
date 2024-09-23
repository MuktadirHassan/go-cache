# HTTP Proxy Server with Caching

This project is an HTTP proxy server that forwards requests to a target server, caches the responses, and serves the cached responses for subsequent requests. It also provides a debug endpoint to retrieve cache information.

## Features

- **Proxy Requests**: Forwards HTTP GET and POST requests to a target server.
- **Caching**: Caches responses to reduce load on the target server and improve response times.
- **Debug Endpoint**: Provides debug information about the cached entries.
- **Health Check Endpoint**: Simple health check endpoint to verify the server is running.

## Installation

1. **Clone the repository**:
    ```sh
    cd go-proxy-cache
    ```

2. **Build the project**:
    ```sh
    go build -o proxy-server ./cmd/main.go
    ```

3. **Run the server**:
    ```sh
    ./proxy-server
    ```

## Usage

### Proxy Endpoint

- **URL**: `/`
- **Method**: `GET` or `POST`
- **Query Parameter**: `target` (The target URL to forward the request to)

Example:

```sh
curl "http://localhost:8080/?target=http://example.com"
```

### Debug Endpoint

- **URL**: `/debug`
- **Method**: [`GET`]()

Example:
```sh
curl "http://localhost:8080/debug"
```

### Health Check Endpoint

- **URL**: `/health`
- **Method**: [`GET`]()

Example:
```sh
curl "http://localhost:8080/health"
```

