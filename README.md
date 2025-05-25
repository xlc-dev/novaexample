# Nova Example Project

This is a example project built with [Nova](https://github.com/xlc-dev/nova).

## Prerequisites

- Go 1.18 or higher
- (Optional) [curl](https://curl.se/) for testing

## Installation

clone the repo and run `make` or `go build -o novaexample` to build the binary.

## Running the API

```bash
# Default: host=localhost, port=8080
./novaexample

# Custom host/port
./novaexample --host=0.0.0.0 --port=3000
```

## OpenAPI

The API is documented with OpenAPI.

You can view the JSON at [http://localhost:8080/openapi.json](http://localhost:8080/openapi.json).

Or you can view the Swagger UI at [http://localhost:8080/docs](http://localhost:8080/docs).

## Example testing with curl

1. **Health check**

   ```bash
   curl -i http://localhost:8080/
   ```

   ```sh
   HTTP/1.1 200 OK
   Hello, Nova!
   ```

2. **Create a valid item**

   ```bash
   curl -i -X POST http://localhost:8080/api/v1/items \
     -H "Content-Type: application/json" \
     -d '{"name":"Foo","isActive":true}'
   ```

   ```json
   HTTP/1.1 201 Created
   Content-Type: application/json

   {"id":1,"name":"Foo","createdAt":"2025-05-11T13:42:00Z","isActive":true}
   ```

3. **Create with invalid name (too long)**

   ```bash
   curl -i -X POST http://localhost:8080/api/v1/items \
     -H "Content-Type: application/json" \
     -d '{"name":"TooLongName","isActive":false}'
   ```

   ```sh
   HTTP/1.1 400 Bad Request
   validation error: field "name" must be at most 5 characters
   ```

4. **Fetch an existing item**

   ```bash
   curl -i http://localhost:8080/api/v1/items/1
   ```

   ```json
   HTTP/1.1 200 OK
   Content-Type: application/json

   {"id":1,"name":"Foo","createdAt":"2025-05-11T13:42:00Z","isActive":true}
   ```

5. **Fetch a non-existent item**

   ```bash
   curl -i http://localhost:8080/api/v1/items/999
   ```

   ```json
   HTTP/1.1 404 Not Found
   Content-Type: application/json

   {"error":"item 999 not found"}
   ```
