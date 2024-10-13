# PostgreSQL Proxy for Testing

This project provides a PostgreSQL proxy server that creates a new PostgreSQL instance for each incoming connection, ideal for unit and API testing.

## Overview

- The proxy listens on port 5432 internally and is mapped to port 6007 on the host machine via Docker.
- Each connection creates a new PostgreSQL instance in a temporary directory, ensuring a clean state for testing.

## Prerequisites

- Docker
- Docker Compose

## Getting Started

1. **Clone the repository:**

   ```bash
   git clone <repository-url>
   cd <repository-directory>
   ```

2. **Start the server:**

   ```bash
   docker-compose up
   ```

   The server will start and listen on port 6007 of the host machine.

## Usage

Connect to the PostgreSQL proxy using any PostgreSQL client from any programming language:

- Host: `localhost`
- Port: `6007`
- Username: postgres
- Password: you can use anything here
- Database: test

Each connection creates a new, clean PostgreSQL instance.

## Project Structure

- `README.md`: This documentation file
- `server.go`: Main server logic
- `cmd/main.go`: Application entry point
- `Dockerfile`: Docker image configuration
- `docker-compose.yml`: Docker Compose configuration

## Stopping the Server

Use `Ctrl+C` in the terminal where the server is running to stop it.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
