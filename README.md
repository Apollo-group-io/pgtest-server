# PG Test Server

PG Test Server is a PostgreSQL-based service designed to provide isolated database environments for unit testing and API testing. It creates temporary database clones from a base database for each connection, ensuring that tests run in a clean, consistent environment.

## Features

- Maintains a base database that can be initialized from a dump file.
- Changes made to the base database are re-synced into the dump file, after the connection to base db closes.
- Serves temporary databases per connection initialized from the current state of the dump file
- Supports concurrent connections, for temporary databases. Each is isolated from the others

## How It Works

1. The server runs on two ports:
   - 5432: For temporary database connections
   - 5433: For base database connections
   - In the docker-compose file these ports are re-mapped to 6543 and 6544 respectively.

2. Base Database:
   - Initialized from a dump file (if available) on server start.
   - All incoming connections on 5433 goto the same basedb instance.
   - After every connection ends, base db is re-dumped to the dump file

3. Temporary Databases:
   - Created for each new connection to port 5432
   - Cloned from the latest dump file
   - Destroyed after the connection is closed

## Usage

1. Start the server:
   ```
   docker compose up -d
   ```

2. Connect to the temporary database port (5432) for isolated test environments.

3. Connect to the base database port (5433) to make persistent changes - that you desire in every test env.

## License
To be added
