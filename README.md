# ShortURL Service

ShortURL Service is a Go-based URL shortener service that provides HTTP and gRPC APIs for shortening, retrieving, and managing URLs.

## Features
- Shorten URLs via HTTP and gRPC
- Retrieve original URLs
- Batch URL shortening support
- User authentication and authorization
- URL deletion support
- Middleware for authentication, logging, and compression
- Graceful shutdown handling

## Technologies Used
- Go (Golang)
- Chi Router
- gRPC
- PostgreSQL (or any GORM-compatible database)
- Middleware for logging, authentication, and compression

## Installation

### Prerequisites
- Go 1.17 or later
- PostgreSQL (or any compatible database)

### Steps
1. Clone the repository:
   ```sh
   git clone https://github.com/golangTroshin/shorturl.git
   cd shorturl
   ```
2. Install dependencies:
   ```sh
   go mod tidy
   ```
3. Run the application:
   ```sh
   go run main.go
   ```

## Configuration
The application uses both environment variables and command-line flags for configuration.

### Available Configuration Options:
| Environment Variable       | Flag | Default Value | Description |
|----------------------------|------|--------------|-------------|
| `SERVER_ADDRESS`           | `-a` | `:8080`      | HTTP server address |
| `BASE_URL`                 | `-b` | `http://localhost:8080` | Base URL for short URLs |
| `FILE_STORAGE_PATH`        | `-f` | `""`         | File storage path for data persistence |
| `DATABASE_DSN`             | `-d` | `""`         | Database connection string |
| `ENABLE_HTTPS`             | `-s` | `false`      | Enable HTTPS mode |
| `CONFIG`                   | `-c` | `""`         | Path to JSON configuration file |
| `TRUSTED_SUBNET`           | `-t` | `192.168.1.0/24` | Trusted subnet for internal operations |

These configurations can be provided through environment variables or modified using command-line flags at runtime. Additionally, if a configuration file is specified, it will override command-line flags and environment variables.

## API Endpoints
### Authentication
- `POST /api/user/register` - Register a new user
- `POST /api/user/login` - Login an existing user

### URL Shortening
- `POST /` - Shorten a URL
- `POST /api/shorten` - Shorten a URL via API
- `POST /api/shorten/batch` - Shorten multiple URLs in batch
- `GET /{id}` - Retrieve the original URL

### User Operations (Requires Authentication)
- `GET /api/user/urls` - Retrieve URLs created by the user
- `DELETE /api/user/urls` - Delete multiple URLs created by the user
- `GET /ping` - Database health check

## gRPC API
The gRPC server is available at `:50051` and provides the following services:
- `ShortenURL` - Shorten a URL
- `GetOriginalURL` - Retrieve the original URL
- `GetUserURLs` - Retrieve URLs created by a user
- `DeleteUserURLs` - Delete multiple URLs created by a user
- `GetStats` - Retrieve service statistics (total URLs and users count)
- `Ping` - Check service health status

## Graceful Shutdown
The application handles OS signals (`SIGTERM`, `SIGINT`, `SIGQUIT`) to allow a graceful shutdown, ensuring all ongoing processes are completed before termination.

