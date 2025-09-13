# Goly - The Go URL Shortener

Goly is a simple, fast, and powerful URL shortener built with Go and Svelte. It provides a clean and easy-to-use interface for creating, managing, and tracking shortened URLs.

## Features

- **URL Shortening**: Create short, memorable URLs for your long links.
- **Random URL Generation**: Automatically generate random, unique short URLs.
- **Click Tracking**: Track the number of clicks on each shortened URL.
- **Simple API**: A straightforward REST API for managing your links.
- **Modern Frontend**: A reactive and user-friendly interface built with Svelte.

## Technology Stack

- **Backend**: Go with Fiber
- **Frontend**: Svelte
- **Database**: PostgreSQL
- **ORM**: GORM

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.15 or higher)
- [Node.js](https://nodejs.org/en/download/) (version 14 or higher)
- [Yarn](https://classic.yarnpkg.com/en/docs/install/)
- [Docker](https://docs.docker.com/get-docker/)

### Setup

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/n3d1117/goly.git
    cd goly
    ```

2.  **Set up the database:**

    The application uses a PostgreSQL database. The easiest way to get a database running is with Docker.

    ```bash
    docker run --name goly-db -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=test -p 5432:5432 -d postgres:14
    ```

    This will start a PostgreSQL container named `goly-db` on port `5432`.

3.  **Configure the backend:**

    The backend requires a database connection string. The `app/model/model.go` file is hardcoded to connect to the Docker container started in the previous step. If your database is running on a different host or with different credentials, you will need to update the DSN (Data Source Name) in that file.

    ```go
    // app/model/model.go
    dsn := "host=localhost user=admin password=test dbname=admin port=5432 sslmode=disable"
    ```

    *Note: The original DSN was hardcoded to a specific Docker IP address. It has been updated to use `localhost` for better compatibility.*

4.  **Install backend dependencies and run the server:**

    ```bash
    cd app
    go mod tidy
    go run main.go
    ```

    The backend server will start on `http://localhost:3000`.

5.  **Install frontend dependencies and run the development server:**

    In a new terminal window:

    ```bash
    cd view
    yarn install
    yarn dev
    ```

    The frontend development server will start on `http://localhost:8080`. You can now access the application in your browser at this address.

## API Endpoints

The backend provides the following REST API endpoints:

| Method | Path           | Description                   |
| ------ | -------------- | ----------------------------- |
| `GET`    | `/r/:redirect` | Redirect to the original URL. |
| `GET`    | `/goly`        | Get all Goly links.           |
| `GET`    | `/goly/:id`    | Get a Goly link by ID.        |
| `POST`   | `/goly`        | Create a new Goly link.       |
| `PATCH`  | `/goly`        | Update a Goly link.           |
| `DELETE` | `/goly/:id`    | Delete a Goly link by ID.     |

### Example Payloads

**Create Goly Link (`POST /goly`)**

```json
{
    "redirect": "https://www.google.com",
    "goly": "google",
    "random": false
}
```

**Update Goly Link (`PATCH /goly`)**

```json
{
    "id": 1,
    "redirect": "https://www.google.com",
    "goly": "google-new",
    "random": false
}
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
