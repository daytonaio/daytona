/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OpenAPIObject, getSchemaPath } from '@nestjs/swagger'
import { WebhookEvents } from './webhook/constants/webhook-events.constants'
import {
  SandboxCreatedWebhookDto,
  SandboxStateUpdatedWebhookDto,
  SnapshotCreatedWebhookDto,
  SnapshotStateUpdatedWebhookDto,
  SnapshotRemovedWebhookDto,
  VolumeCreatedWebhookDto,
  VolumeStateUpdatedWebhookDto,
} from './webhook/dto/webhook-event-payloads.dto'

export interface OpenAPIObjectWithWebhooks extends OpenAPIObject {
  webhooks?: {
    [key: string]: {
      post: {
        requestBody: {
          description: string
          content: {
            'application/json': {
              schema: any
            }
          }
        }
        responses: {
          [statusCode: string]: {
            description: string
          }
        }
      }
    }
  }
}

export function addWebhookDocumentation(document: OpenAPIObject): OpenAPIObjectWithWebhooks {
  return {
    ...document,
    webhooks: {
      [WebhookEvents.SANDBOX_CREATED]: {
        post: {
          requestBody: {
            description: 'Sandbox created event',
            content: {
              'application/json': {
                schema: { $ref: getSchemaPath(SandboxCreatedWebhookDto) },
              },
            },
          },
          responses: {
            '200': {
              description: 'Webhook received successfully',
            },
          },
        },
      },
      [WebhookEvents.SANDBOX_STATE_UPDATED]: {
        post: {
          requestBody: {
            description: 'Sandbox state updated event',
            content: {
              'application/json': {
                schema: { $ref: getSchemaPath(SandboxStateUpdatedWebhookDto) },
              },
            },
          },
          responses: {
            '200': {
              description: 'Webhook received successfully',
            },
          },
        },
      },
      [WebhookEvents.SNAPSHOT_CREATED]: {
        post: {
          requestBody: {
            description: 'Snapshot created event',
            content: {
              'application/json': {
                schema: { $ref: getSchemaPath(SnapshotCreatedWebhookDto) },
              },
            },
          },
          responses: {
            '200': {
              description: 'Webhook received successfully',
            },
          },
        },
      },
      [WebhookEvents.SNAPSHOT_STATE_UPDATED]: {
        post: {
          requestBody: {
            description: 'Snapshot state updated event',
            content: {
              'application/json': {
                schema: { $ref: getSchemaPath(SnapshotStateUpdatedWebhookDto) },
              },
            },
          },
          responses: {
            '200': {
              description: 'Webhook received successfully',
            },
          },
        },
      },
      [WebhookEvents.SNAPSHOT_REMOVED]: {
        post: {
          requestBody: {
            description: 'Snapshot removed event',
            content: {
              'application/json': {
                schema: { $ref: getSchemaPath(SnapshotRemovedWebhookDto) },
              },
            },
          },
          responses: {
            '200': {
              description: 'Webhook received successfully',
            },
          },
        },
      },
      [WebhookEvents.VOLUME_CREATED]: {
        post: {
          requestBody: {
            description: 'Volume created event',
            content: {
              'application/json': {
                schema: { $ref: getSchemaPath(VolumeCreatedWebhookDto) },
              },
            },
          },
          responses: {
            '200': {
              description: 'Webhook received successfully',
            },
          },
        },
      },
      [WebhookEvents.VOLUME_STATE_UPDATED]: {
        post: {
          requestBody: {
            description: 'Volume state updated event',
            content: {
              'application/json': {
                schema: { $ref: getSchemaPath(VolumeStateUpdatedWebhookDto) },
              },
            },
          },
          responses: {
            '200': {
              description: 'Webhook received successfully',
            },
          },
        },
      },
    },
  }
}
