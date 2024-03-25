# Gochirp API

Gochirp is a Golang HTTP server that provides various endpoints for managing chirps and users.

The API is built using the following technologies:

- [Golang](https://golang.org/)
- [Golang jwt](https://golang-jwt.github.io/jwt/)
- [Gorilla Mux](https://github.com/gorilla/mux)


The API provides the following features:
## User Management
    - User registration
    - User retrieval
    - User deletion
    - Polka webhook handling
    - User login
## Chirp Management
    - Chirp creation
    - Chirp validation
    - Chirp retrieval
    - Chirp deletion
    - Metrics

## Endpoints

### `/api/healthz`

A health check endpoint.

### `/admin/metrics`

An endpoint that provides metrics about the API usage.

### `/api/reset`

An endpoint to reset the API.

### `POST /api/validate_chirp`

An endpoint to validate a chirp.

### `/api/chirp`

An endpoint to handle chirp-related requests. Requires authentication.

### `/api/users`

An endpoint to handle user-related requests. Requires authentication.

### `POST /api/login`

An endpoint for user login.

### `POST /api/register`

An endpoint for user registration.

### `/api/polka/webhooks`

An endpoint to handle Polka webhooks.

## Running the Server

To run the server, execute the following command:

```sh
go run main.go
```
#  Environment Variables
- JWT_SECRET: The secret used for JWT authentication.
- POLKA_SECRET: The secret used for Polka webhooks.
