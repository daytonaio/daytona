# SSH Gateway

A standalone SSH gateway application that authenticates users using tokens and proxies connections to Daytona runners.

## Features

- **Token-based authentication**: Username is used as the authentication token
- **Automatic runner discovery**: Validates tokens and finds the appropriate runner
- **SSH keypair management**: Retrieves SSH credentials from runners automatically
- **Connection proxying**: Seamlessly forwards SSH connections to runners

## Usage

### Connection Format

```bash
ssh -p 2222 <TOKEN>@<GATEWAY_HOST>
```

**Example:**

```bash
ssh -p 2222 Fg8Jx2nPtWAVY5pVN0TlUcbCDNPF-ePB@localhost
```

Where:

- `Fg8Jx2nPtWAVY5pVN0TlUcbCDNPF-ePB` is the SSH access token
- `localhost` is the SSH gateway host
- `2222` is the SSH gateway port

### How It Works

1. **Authentication**: The gateway extracts the token from the username
2. **Token Validation**: Calls the Daytona API to validate the token and get runner information
3. **Credential Retrieval**: Fetches the SSH keypair from the identified runner
4. **Connection Proxying**: Establishes a connection to the runner's SSH gateway (port 2222)
5. **Session Forwarding**: Proxies the SSH session between the client and the runner

## Configuration

### Environment Variables

| Variable           | Description                           | Default                 | Required |
| ------------------ | ------------------------------------- | ----------------------- | -------- |
| `SSH_GATEWAY_PORT` | Port for the SSH gateway to listen on | `2222`                  | No       |
| `API_URL`          | Daytona API base URL                  | `http://localhost:3000` | No       |
| `API_KEY`          | Daytona API authentication key        | -                       | **Yes**  |

### Example Environment

```bash
export SSH_GATEWAY_PORT=2222
export API_URL=https://api.daytona.example.com
export API_KEY=your-api-key-here
```

## Building

### Local Build

```bash
go mod tidy
go build -o ssh-gateway .
```

### Docker Build

```bash
docker build -t ssh-gateway .
```

## Running

### Local Execution

```bash
./ssh-gateway
```

### Docker Execution

```bash
docker run -p 2222:2222 \
  -e API_URL=https://api.daytona.example.com \
  -e API_KEY=your-api-key-here \
  ssh-gateway
```

## Security

- **No password authentication**: Only token-based authentication is supported
- **No public key authentication**: Public keys are not accepted
- **Temporary host keys**: The gateway generates new host keys on each startup
- **Secure token validation**: All tokens are validated against the Daytona API
- **Runner isolation**: Each connection is isolated to its specific runner

## Architecture

```
SSH Client → SSH Gateway → Daytona API → Runner SSH Gateway
     ↓              ↓           ↓              ↓
  Token Auth   Validate    Get Keypair    SSH Session
```

## API Endpoints Used

- `GET /sandbox/ssh-access/validate?token={token}` - Validate SSH access token
- `GET /runners/{id}/ssh-keypair` - Get runner SSH keypair

## Error Handling

The gateway provides clear error messages for common failure scenarios:

- Invalid or expired tokens
- Runner unavailability
- Network connectivity issues
- Authentication failures

## Logging

The application logs all connection attempts and errors for debugging and monitoring purposes.
