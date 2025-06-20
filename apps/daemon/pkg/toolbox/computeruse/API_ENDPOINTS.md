# ComputerUse API Endpoints

This document describes the new API endpoints for managing VNC desktop processes through the toolbox API.

## Overview

The ComputerUse API provides endpoints to start, stop, monitor, and manage the VNC desktop environment processes (Xvfb, xfce4, x11vnc, novnc) that were previously managed by supervisor.

## Base URL

All endpoints are available under the `/api/toolbox/{sandboxId}/toolbox/computeruse` path.

## Authentication

All endpoints require authentication and proper sandbox access permissions.

## Endpoints

### 1. Start Computer Use Processes

**POST** `/api/toolbox/{sandboxId}/toolbox/computeruse/start`

Starts all VNC desktop processes in the correct order:

1. Xvfb (X Virtual Framebuffer)
2. xfce4 (Desktop Environment)
3. x11vnc (VNC Server)
4. novnc (Web-based VNC client)

**Parameters:**

- `sandboxId` (path parameter): The ID of the sandbox

**Response:**

```json
{
  "message": "Computer use processes started successfully",
  "status": {
    "xvfb": {
      "running": true,
      "priority": 100,
      "autoRestart": true,
      "pid": 12345
    },
    "xfce4": {
      "running": true,
      "priority": 200,
      "autoRestart": true,
      "pid": 12346
    },
    "x11vnc": {
      "running": true,
      "priority": 300,
      "autoRestart": true,
      "pid": 12347
    },
    "novnc": {
      "running": true,
      "priority": 400,
      "autoRestart": true,
      "pid": 12348
    }
  }
}
```

### 2. Stop Computer Use Processes

**POST** `/api/toolbox/{sandboxId}/toolbox/computeruse/stop`

Stops all VNC desktop processes in reverse order.

**Parameters:**

- `sandboxId` (path parameter): The ID of the sandbox

**Response:**

```json
{
  "message": "Computer use processes stopped successfully",
  "status": {
    "xvfb": {
      "running": false,
      "priority": 100,
      "autoRestart": true
    },
    "xfce4": {
      "running": false,
      "priority": 200,
      "autoRestart": true
    },
    "x11vnc": {
      "running": false,
      "priority": 300,
      "autoRestart": true
    },
    "novnc": {
      "running": false,
      "priority": 400,
      "autoRestart": true
    }
  }
}
```

### 3. Get Computer Use Status

**GET** `/api/toolbox/{sandboxId}/toolbox/computeruse/status`

Returns the status of all VNC desktop processes.

**Parameters:**

- `sandboxId` (path parameter): The ID of the sandbox

**Response:**

```json
{
  "status": {
    "xvfb": {
      "running": true,
      "priority": 100,
      "autoRestart": true,
      "pid": 12345
    },
    "xfce4": {
      "running": true,
      "priority": 200,
      "autoRestart": true,
      "pid": 12346
    },
    "x11vnc": {
      "running": true,
      "priority": 300,
      "autoRestart": true,
      "pid": 12347
    },
    "novnc": {
      "running": true,
      "priority": 400,
      "autoRestart": true,
      "pid": 12348
    }
  }
}
```

### 4. Get Process Status

**GET** `/api/toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/status`

Returns the status of a specific VNC process.

**Parameters:**

- `sandboxId` (path parameter): The ID of the sandbox
- `processName` (path parameter): The name of the process (xvfb, xfce4, x11vnc, novnc)

**Response:**

```json
{
  "processName": "xfce4",
  "running": true
}
```

### 5. Restart Process

**POST** `/api/toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/restart`

Restarts a specific VNC process.

**Parameters:**

- `sandboxId` (path parameter): The ID of the sandbox
- `processName` (path parameter): The name of the process (xvfb, xfce4, x11vnc, novnc)

**Response:**

```json
{
  "message": "Process xfce4 restarted successfully",
  "processName": "xfce4"
}
```

**Error Response:**

```json
{
  "error": "process xfce4 not found"
}
```

### 6. Get Process Logs

**GET** `/api/toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/logs`

Returns the logs for a specific VNC process.

**Parameters:**

- `sandboxId` (path parameter): The ID of the sandbox
- `processName` (path parameter): The name of the process (xvfb, xfce4, x11vnc, novnc)

**Response:**

```json
{
  "processName": "xfce4",
  "logs": "Starting XFCE desktop environment...\nDisplay :1 configured...\n"
}
```

**Error Response:**

```json
{
  "error": "no log file configured for process xvfb"
}
```

### 7. Get Process Errors

**GET** `/api/toolbox/{sandboxId}/toolbox/computeruse/process/{processName}/errors`

Returns the error logs for a specific VNC process.

**Parameters:**

- `sandboxId` (path parameter): The ID of the sandbox
- `processName` (path parameter): The name of the process (xvfb, xfce4, x11vnc, novnc)

**Response:**

```json
{
  "processName": "xfce4",
  "errors": "Error: Display :1 not available\n"
}
```

**Error Response:**

```json
{
  "error": "no error file configured for process xvfb"
}
```

## Process Names

The following process names are supported:

- `xvfb` - X Virtual Framebuffer
- `xfce4` - XFCE Desktop Environment
- `x11vnc` - VNC Server
- `novnc` - Web-based VNC Client

## Error Handling

All endpoints return appropriate HTTP status codes:

- `200` - Success
- `400` - Bad Request (e.g., process not found, no log file configured)
- `401` - Unauthorized
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (sandbox not found)

## Usage Examples

### Start VNC Desktop Environment

```bash
curl -X POST \
  "https://api.daytona.io/api/toolbox/my-sandbox/toolbox/computeruse/start" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Organization-Id: YOUR_ORG_ID"
```

### Check Process Status

```bash
curl -X GET \
  "https://api.daytona.io/api/toolbox/my-sandbox/toolbox/computeruse/process/xfce4/status" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Organization-Id: YOUR_ORG_ID"
```

### Get Process Logs

```bash
curl -X GET \
  "https://api.daytona.io/api/toolbox/my-sandbox/toolbox/computeruse/process/xfce4/logs" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Organization-Id: YOUR_ORG_ID"
```

### Restart a Process

```bash
curl -X POST \
  "https://api.daytona.io/api/toolbox/my-sandbox/toolbox/computeruse/process/x11vnc/restart" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Organization-Id: YOUR_ORG_ID"
```

## Integration with Existing Endpoints

These new endpoints complement the existing computer use endpoints:

- **Screenshot endpoints**: `/computeruse/screenshot/*`
- **Mouse control endpoints**: `/computeruse/mouse/*`
- **Keyboard control endpoints**: `/computeruse/keyboard/*`
- **Display info endpoints**: `/computeruse/display/*`

The process management endpoints allow you to control the underlying VNC infrastructure, while the existing endpoints provide interaction with the desktop environment.

## Graceful Shutdown

The daemon automatically handles graceful shutdown of all computer use processes when it receives a termination signal (SIGINT, SIGTERM). This ensures that all VNC processes are properly stopped and cleaned up.
