# Webhook Service

This service provides webhook functionality using [Svix](https://svix.com) as the webhook delivery provider. It automatically creates Svix applications for new organizations and sends webhooks for various events.

## Configuration

Set the following environment variables:

```bash
# Required: Your Svix authentication token
SVIX_AUTH_TOKEN=your_svix_auth_token_here

# Optional: Custom Svix server URL (for self-hosted instances)
SVIX_SERVER_URL=https://your-svix-instance.com
```

## API Endpoints

### Get App Portal Access

```http
POST /api/webhooks/organizations/{organizationId}/app-portal-access
```

**Response:**

```json
{
  "url": "https://app.svix.com/consumer/..."
}
```

Returns a URL that provides access to the Svix Consumer App Portal for managing webhook endpoints, viewing delivery attempts, and monitoring webhook performance.

### Send Custom Webhook

```http
POST /api/webhooks/organizations/{organizationId}/send
```

**Request Body:**

```json
{
  "eventType": "custom.event",
  "payload": {
    "message": "Hello from Daytona!",
    "timestamp": "2025-01-01T00:00:00.000Z"
  },
  "eventId": "optional-unique-id"
}
```

Sends a custom webhook message to all configured endpoints for the specified organization.

### Get Message Delivery Attempts

```http
GET /api/webhooks/organizations/{organizationId}/messages/{messageId}/attempts
```

**Response:**

```json
[
  {
    "id": "msg_attempt_123",
    "status": 200,
    "response": "OK",
    "timestamp": "2025-01-01T00:00:00.000Z"
  }
]
```

Returns the delivery attempts for a specific webhook message, including delivery status and response details.

### Get Service Status

```http
GET /api/webhooks/status
```

**Response:**

```json
{
  "enabled": true
}
```

Returns the current status of the webhook service, indicating whether it's properly configured and enabled.

## Automatic Events

The service automatically sends webhooks for the following events:

### Sandbox Events

- `sandbox.created` - When a sandbox is created
- `sandbox.state.updated` - When sandbox state changes

### Snapshot Events

- `snapshot.created` - When a snapshot is created
- `snapshot.state.updated` - When snapshot state changes
- `snapshot.removed` - When a snapshot is removed

### Volume Events

- `volume.created` - When a volume is created
- `volume.state.updated` - When volume state changes

## Webhook Payload Format

All webhooks include event-specific data relevant to the resource being updated.

### Example Sandbox Created Payload

```json
{
  "id": "sandbox-uuid",
  "organizationId": "org-uuid",
  "state": "STARTED",
  "class": "SMALL",
  "createdAt": "2025-01-01T00:00:00.000Z"
}
```

### Example Sandbox State Updated Payload

```json
{
  "id": "sandbox-uuid",
  "organizationId": "org-uuid",
  "oldState": "STOPPED",
  "newState": "STARTED",
  "updatedAt": "2025-01-01T00:00:00.000Z"
}
```

### Example Snapshot Created Payload

```json
{
  "id": "snapshot-uuid",
  "name": "my-snapshot",
  "organizationId": "org-uuid",
  "state": "ACTIVE",
  "createdAt": "2025-01-01T00:00:00.000Z"
}
```

### Example Volume State Updated Payload

```json
{
  "id": "volume-uuid",
  "name": "my-volume",
  "organizationId": "org-uuid",
  "oldState": "CREATING",
  "newState": "READY",
  "updatedAt": "2025-01-01T00:00:00.000Z"
}
```

## Development

### Adding New Event Types

1. Add the event type to `webhook-events.constants.ts`
2. Create an event handler in `webhook-event-handler.service.ts`
3. Use the `@OnEvent()` decorator to listen for the event
4. Define the payload structure for the new event type
5. Add the events to the `openapi-webhooks.ts`
6. Generate the openapi spec
7. Upload the new schema to the Svix dashboard

### Testing

Use the Svix Play webhook debugger during development:

1. Set up a webhook endpoint pointing to your Svix Play URL
2. Send test webhooks using the `/send` endpoint
3. Check the Svix dashboard for delivery status
4. Monitor delivery attempts through the API

### Local Development

For local development without Svix:

1. Set `SVIX_AUTH_TOKEN` to an empty string or invalid value
2. The service will log warnings but continue to function
3. Event handlers will skip webhook delivery when disabled
4. Use the status endpoint to verify configuration

## Dependencies

- `svix` - Official Svix JavaScript SDK
- `@nestjs/event-emitter` - Event handling
- `@nestjs/common` - Core NestJS functionality

### Event Flow

1. System event occurs (e.g., sandbox created)
2. Event emitter publishes the event
3. Webhook event handler catches the event
4. Handler calls webhook service to send webhook
5. Service delivers webhook through Svix
6. Delivery status is tracked and available via API
