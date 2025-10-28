# WebSocket Close Code 4008 Test

This test suite demonstrates a Go WebSocket server sending close code 4008 and clients in different languages receiving it.

## Files

- `server.go` - Go WebSocket server that sends close code 4008
- `client.py` - Python WebSocket client
- `client.ts` - TypeScript WebSocket client
- `go.mod` - Go module dependencies
- `package.json` - Node.js/TypeScript dependencies
- `tsconfig.json` - TypeScript configuration
- `run_test.sh` - Automated test runner (Python)
- `run_test_ts.sh` - Automated test runner (TypeScript)

## Setup

### Go Server

```bash
# Install dependencies
go mod download

# Run the server
go run server.go
```

The server will start on `ws://localhost:8088/ws`

### Python Client

```bash
# Install dependencies
pip install websockets

# Run the client
python3 client.py
```

### TypeScript Client

```bash
# Install dependencies
npm install

# Build TypeScript
npm run build

# Run the client
npm run start

# Or run directly with ts-node
npm run dev
```

## Expected Behavior

1. Server starts and listens on port 8088
2. Client connects to the server
3. After 2 seconds, server sends WebSocket close frame with code 4008
4. Client receives the close event and displays:
   - Close code: 4008
   - Close reason: "Custom close code 4008"
5. Connection closes gracefully

## Testing

### Manual Testing

**Option 1: Python Client**

Terminal 1:
```bash
cd /workspaces/daytona/websocket-test
go run server.go
```

Terminal 2:
```bash
cd /workspaces/daytona/websocket-test
python3 client.py
```

**Option 2: TypeScript Client**

Terminal 1:
```bash
cd /workspaces/daytona/websocket-test
go run server.go
```

Terminal 2:
```bash
cd /workspaces/daytona/websocket-test
npm run start
```

### Automated Testing

**Python:**
```bash
./run_test.sh
```

**TypeScript:**
```bash
./run_test_ts.sh
```

Both should show the client successfully receiving close code 4008.

