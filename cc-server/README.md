# C&C Server

A secure, versatile Command and Control (C&C) server with a CLI interface inspired by Qwen Code.

## Features

- Secure WebSocket-based communication between server and clients
- Supabase-powered database for storing client information and command history
- JWT-based authentication for clients
- CLI interface for server management
- Client registration with tokens
- Persistent client connections
- Command execution with result reporting
- Heartbeat mechanism to track client status

## Architecture

The C&C server consists of three main components:

1. **Server**: Central management component that handles client registration, command distribution, and status tracking.
2. **Client**: Agent that connects to the server and executes commands.
3. **CLI**: Command-line interface for managing the server, sending commands, and monitoring clients.

## Database Schema

The application uses Supabase as its database backend. The schema includes:

- `clients`: Stores registered client information
- `registration_tokens`: Temporary tokens for client registration
- `commands`: Command execution history

## Setup

1. Create the database tables by running the schema.sql file against your Supabase database:
   ```sql
   -- Run the contents of docs/schema.sql in your Supabase SQL editor
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Usage

### Running the Server

```bash
go run cmd/server/main.go
```

### Running a Client

```bash
go run cmd/client/main.go -token YOUR_REGISTRATION_TOKEN
```

### Using the CLI

List all clients:
```bash
go run cmd/cli/main.go list-clients
```

Send a command to a client:
```bash
go run cmd/cli/main.go send CLIENT_ID "ls -la"
```

Get command history for a client:
```bash
go run cmd/cli/main.go get-commands CLIENT_ID
```

## Security

- All client-server communication is authenticated
- Registration tokens are required for initial client registration
- Client connections are tracked and managed
- JWT tokens are used for ongoing client authentication

## License

This project is licensed under the MIT License.