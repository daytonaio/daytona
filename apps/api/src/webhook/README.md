# Webhook Service

This service provides webhook functionality using [Svix](https://svix.com) as the webhook delivery provider. It automatically creates Svix applications for new organizations and sends webhooks for various events.

## Features

- **Automatic Organization Integration**: Creates a Svix application when a new organization is created
- **Preconfigured Endpoints**: Automatically creates categorized webhook endpoints for all event types
- **Event-Driven Webhooks**: Sends webhooks for sandbox, snapshot, volume, and audit log events
- **REST API**: Provides endpoints for managing webhook endpoints and sending custom webhooks
- **Real-time Updates**: WebSocket gateway for real-time webhook delivery status
- **Idempotency Support**: Optional event IDs for preventing duplicate webhook deliveries

## Configuration

Set the following environment variables:

```bash
# Required: Your Svix authentication token
SVIX_AUTH_TOKEN=your_svix_auth_token_here

# Optional: Custom Svix server URL (for self-hosted instances)
SVIX_SERVER_URL=https://your-svix-instance.com
```

## API Endpoints

### Create Webhook Endpoint

```http
POST /api/webhooks/organizations/{organizationId}/endpoints
```

**Request Body:**

```json
{
  "url": "https://your-webhook-endpoint.com/webhooks",
  "description": "Production webhook endpoint",
  "eventTypes": ["sandbox.created", "sandbox.updated"]
}
```

### List Webhook Endpoints

```http
GET /api/webhooks/organizations/{organizationId}/endpoints
```

### Delete Webhook Endpoint

```http
DELETE /api/webhooks/organizations/{organizationId}/endpoints/{endpointId}
```

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

### Create Preconfigured Endpoints

```http
POST /api/webhooks/organizations/{organizationId}/preconfigured-endpoints
```

Creates automatically categorized webhook endpoints for all event types.

### Get Preconfigured Endpoint Configurations

```http
GET /api/webhooks/preconfigured-endpoints
```

Returns the list of available preconfigured endpoint configurations.

### Initialize Webhooks for an Organization

```http
POST /api/webhooks/organizations/{organizationId}/initialize
```

Initializes webhooks for an existing organization (creates Svix app and endpoints).

### Get Initialization Status

```http
GET /api/webhooks/organizations/{organizationId}/initialization-status
```

Returns the webhook initialization status for an organization.

### Get Initialization Statistics

```http
GET /api/webhooks/initialization-stats
```

Returns overall webhook initialization statistics across all organizations.

### Retry Failed Initializations

```http
POST /api/webhooks/retry-failed-initializations
```

Retries webhook initialization for organizations that previously failed.

### Update All Endpoints

```http
POST /api/webhooks/update-all-endpoints
```

Updates webhook endpoints for all organizations (useful for future webhook updates).

### Get Webhook Status

```http
GET /api/webhooks/status
```

## WebSocket Events

Connect to `/api/webhook-socket.io/` to receive real-time updates:

- `webhook.delivered` - Webhook successfully delivered
- `webhook.failed` - Webhook delivery failed
- `endpoint.created` - New webhook endpoint created
- `endpoint.updated` - Webhook endpoint updated
- `endpoint.deleted` - Webhook endpoint deleted

## Automatic Events

The service automatically sends webhooks for the following events:

### Sandbox Events

- `sandbox.created` - When a sandbox is created
- `sandbox.state.updated` - When sandbox state changes
- `sandbox.desired-state.updated` - When sandbox desired state changes

### Snapshot Events

- `snapshot.created` - When a snapshot is created
- `snapshot.state.updated` - When snapshot state changes
- `snapshot.removed` - When a snapshot is removed

### Volume Events

- `volume.created` - When a volume is created
- `volume.state.updated` - When volume state changes
- `volume.lastUsedAt.updated` - When volume last used timestamp updates

### Audit Events

- `audit-log.created` - When an audit log entry is created
- `audit-log.updated` - When an audit log entry is updated

## Preconfigured Endpoints

When a new organization is created, the service automatically creates the following webhook endpoints:

### 1. Sandbox Events Endpoint

- **URL**: `https://webhook.daytona.io/{organizationId}/sandbox-events`
- **Events**: All sandbox-related events
- **Description**: Dedicated endpoint for sandbox lifecycle events

### 2. Snapshot Events Endpoint

- **URL**: `https://webhook.daytona.io/{organizationId}/snapshot-events`
- **Events**: All snapshot-related events
- **Description**: Dedicated endpoint for snapshot management events

### 3. Volume Events Endpoint

- **URL**: `https://webhook.daytona.io/{organizationId}/volume-events`
- **Events**: All volume-related events
- **Description**: Dedicated endpoint for volume lifecycle events

### 4. Audit Events Endpoint

- **URL**: `https://webhook.daytona.io/{organizationId}/audit-events`
- **Events**: All audit log events
- **Description**: Dedicated endpoint for audit and compliance events

### 5. All Events Endpoint

- **URL**: `https://webhook.daytona.io/{organizationId}/all-events`
- **Events**: All platform events
- **Description**: Comprehensive endpoint for all event types

## Initialization Management

The webhook service tracks initialization status for all organizations to handle existing organizations and future updates.

### Initialization Tracking

Each organization's webhook initialization status is tracked in the database with:

- **Organization ID**: Links to the organization
- **Svix Application Status**: Whether the Svix application was created
- **Endpoints Status**: Whether preconfigured endpoints were created
- **Endpoint IDs**: Array of created endpoint IDs for future updates
- **Error Tracking**: Last error message and retry count
- **Timestamps**: Creation and last update times

### Handling Existing Organizations

For organizations created before webhook service was enabled:

1. **Manual Initialization**: Use `POST /api/webhooks/organizations/{orgId}/initialize`
2. **Status Check**: Use `GET /api/webhooks/organizations/{orgId}/initialization-status`
3. **Bulk Operations**: Use statistics and retry endpoints for management

### Future Webhook Updates

When webhook configurations need to be updated:

1. **Update All**: Use `POST /api/webhooks/update-all-endpoints`
2. **Selective Update**: Check status and update specific organizations
3. **Rollback**: Track changes for potential rollback scenarios

## Webhook Payload Format

All webhooks include:

- `id` - The resource ID
- `organizationId` - The organization ID
- `createdAt` - When the event occurred
- Event-specific data

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

## Security

- All endpoints require authentication via JWT or API key
- Webhooks are scoped to organizations
- Users can only manage webhooks for organizations they have access to

## Error Handling

- Failed webhook deliveries are logged
- Webhook service gracefully handles configuration errors
- Service continues to function even if Svix is unavailable

## Development

### Adding New Event Types

1. Add the event type to `webhook-events.constant.ts`
2. Create an event handler in `webhook-event-handler.service.ts`
3. Use the `@OnEvent()` decorator to listen for the event

### Testing

Use the Svix Play webhook debugger during development:

1. Set up a webhook endpoint pointing to your Svix Play URL
2. Send test webhooks to verify delivery
3. Check the Svix dashboard for delivery status

## Dependencies

- `svix` - Official Svix JavaScript SDK
- `@nestjs/websockets` - WebSocket support
- `@nestjs/event-emitter` - Event handling
