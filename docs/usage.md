# Goly Usage Guide

## Introduction

Goly is a URL shortener application written in Go. It provides a simple API for creating, managing, and redirecting shortened URLs. This guide explains how to set up and run the Goly application.

## Prerequisites

To run Goly, you will need the following tools installed on your system:

- **Go:** The Go programming language. You can find installation instructions at [https://golang.org/doc/install](https://golang.org/doc/install).
- **Docker:** A containerization platform used to run the PostgreSQL database. You can find installation instructions at [https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/).

## Setup

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd goly-app
    ```

2.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

## Database Setup

The `README.md` file specifies that the application uses a PostgreSQL database. You can start a PostgreSQL container using the following Docker command:

```bash
docker run --name goly-postgres -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=test -p 5432:5432 -d postgres:14
```

**Note:** There is a discrepancy between the `README.md` and the application code. The code is currently configured to use a SQLite database (`goly.db`), not PostgreSQL. To use PostgreSQL, you would need to modify the database connection code in `database/database.go`.

## Running the Application

To run the application, use the following command:

```bash
go run goly/main.go
```

The server should start on port `3000`.

**Note on the execution environment:** During development, I encountered issues running the application in the provided sandbox environment. The `go run` and `go test` commands would hang indefinitely, even with the database connection code commented out. This suggests an issue with the environment itself. I was unable to get the application to run successfully.

## Testing

This project contains unit tests for the `goly/model` package. The tests are located in `goly/model/goly_test.go`.

### Running the Tests

To run the tests, use the following command from the root of the project:

```bash
go test -v ./goly/model
```

### Test Coverage

The tests cover the following functions:

-   `GetAllGolies`
-   `GetGoly`
-   `CreateGoly`
-   `UpdateGoly`
-   `DeleteGoly`
-   `FindByGolyUrl`

The tests use a manual mock of the `gorm.DB` object to simulate the database. This allows the tests to be run without a real database connection.

**Note:** As I was unable to run the tests in the provided environment, they have been written "blind" and may contain errors. They should be run and verified in a stable environment.

## API Usage

The Goly API provides endpoints for managing and using shortened URLs. For detailed information about the API, please refer to the OpenAPI documentation:

- **[OpenAPI Specification](openapi.yaml)**

You can use a tool like [Swagger UI](https://swagger.io/tools/swagger-ui/) or [Postman](https://www.postman.com/) to interact with the API.
