/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { OpenAPIObject } from '@nestjs/swagger'
import { WebhookEvents } from './webhook/constants/webhook-events.constants'

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
                schema: {
                  type: 'object',
                  properties: {
                    event: { type: 'string', example: 'sandbox.created' },
                    timestamp: { type: 'string', format: 'date-time' },
                    data: {
                      type: 'object',
                      properties: {
                        id: { type: 'string', example: 'sandbox123' },
                        organizationId: { type: 'string', example: 'organization123' },
                        target: { type: 'string', example: 'local' },
                        snapshot: { type: 'string', example: 'daytonaio/sandbox:latest' },
                        user: { type: 'string', example: 'daytona' },
                        env: { type: 'object', additionalProperties: { type: 'string' } },
                        cpu: { type: 'number', example: 2 },
                        gpu: { type: 'number', example: 0 },
                        memory: { type: 'number', example: 4 },
                        disk: { type: 'number', example: 10 },
                        public: { type: 'boolean', example: false },
                        labels: { type: 'object', additionalProperties: { type: 'string' } },
                        state: {
                          type: 'string',
                          enum: [
                            'creating',
                            'restoring',
                            'destroyed',
                            'destroying',
                            'started',
                            'stopped',
                            'starting',
                            'stopping',
                            'error',
                            'build_failed',
                            'pending_build',
                            'building_snapshot',
                            'unknown',
                            'pulling_snapshot',
                            'archived',
                            'archiving',
                          ],
                        },
                        desiredState: {
                          type: 'string',
                          enum: ['destroyed', 'started', 'stopped', 'resized', 'archived'],
                        },
                        createdAt: { type: 'string', format: 'date-time' },
                        updatedAt: { type: 'string', format: 'date-time' },
                      },
                    },
                  },
                },
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
                schema: {
                  type: 'object',
                  properties: {
                    event: { type: 'string', example: 'sandbox.state.updated' },
                    timestamp: { type: 'string', format: 'date-time' },
                    data: {
                      type: 'object',
                      properties: {
                        sandbox: {
                          type: 'object',
                          properties: {
                            id: { type: 'string', example: 'sandbox123' },
                            organizationId: { type: 'string', example: 'organization123' },
                            state: {
                              type: 'string',
                              enum: [
                                'creating',
                                'restoring',
                                'destroyed',
                                'destroying',
                                'started',
                                'stopped',
                                'starting',
                                'stopping',
                                'error',
                                'build_failed',
                                'pending_build',
                                'building_snapshot',
                                'unknown',
                                'pulling_snapshot',
                                'archived',
                                'archiving',
                              ],
                            },
                          },
                        },
                        oldState: {
                          type: 'string',
                          enum: [
                            'creating',
                            'restoring',
                            'destroyed',
                            'destroying',
                            'started',
                            'stopped',
                            'starting',
                            'stopping',
                            'error',
                            'build_failed',
                            'pending_build',
                            'building_snapshot',
                            'unknown',
                            'pulling_snapshot',
                            'archived',
                            'archiving',
                          ],
                        },
                        newState: {
                          type: 'string',
                          enum: [
                            'creating',
                            'restoring',
                            'destroyed',
                            'destroying',
                            'started',
                            'stopped',
                            'starting',
                            'stopping',
                            'error',
                            'build_failed',
                            'pending_build',
                            'building_snapshot',
                            'unknown',
                            'pulling_snapshot',
                            'archived',
                            'archiving',
                          ],
                        },
                      },
                    },
                  },
                },
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
                schema: {
                  type: 'object',
                  properties: {
                    event: { type: 'string', example: 'snapshot.created' },
                    timestamp: { type: 'string', format: 'date-time' },
                    data: {
                      type: 'object',
                      properties: {
                        id: { type: 'string', example: 'snapshot123' },
                        organizationId: { type: 'string', example: 'organization123' },
                        general: { type: 'boolean', example: false },
                        name: { type: 'string', example: 'my-snapshot' },
                        imageName: { type: 'string', example: 'my-image:latest' },
                        state: {
                          type: 'string',
                          enum: [
                            'build_pending',
                            'building',
                            'pending',
                            'pulling',
                            'pending_validation',
                            'validating',
                            'active',
                            'inactive',
                            'error',
                            'build_failed',
                            'removing',
                          ],
                        },
                        size: { type: 'number', example: 1024 },
                        cpu: { type: 'number', example: 2 },
                        gpu: { type: 'number', example: 0 },
                        mem: { type: 'number', example: 4 },
                        disk: { type: 'number', example: 10 },
                        createdAt: { type: 'string', format: 'date-time' },
                        updatedAt: { type: 'string', format: 'date-time' },
                      },
                    },
                  },
                },
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
                schema: {
                  type: 'object',
                  properties: {
                    event: { type: 'string', example: 'snapshot.state.updated' },
                    timestamp: { type: 'string', format: 'date-time' },
                    data: {
                      type: 'object',
                      properties: {
                        snapshot: {
                          type: 'object',
                          properties: {
                            id: { type: 'string', example: 'snapshot123' },
                            organizationId: { type: 'string', example: 'organization123' },
                            state: {
                              type: 'string',
                              enum: [
                                'build_pending',
                                'building',
                                'pending',
                                'pulling',
                                'pending_validation',
                                'validating',
                                'active',
                                'inactive',
                                'error',
                                'build_failed',
                                'removing',
                              ],
                            },
                          },
                        },
                        oldState: {
                          type: 'string',
                          enum: [
                            'build_pending',
                            'building',
                            'pending',
                            'pulling',
                            'pending_validation',
                            'validating',
                            'active',
                            'inactive',
                            'error',
                            'build_failed',
                            'removing',
                          ],
                        },
                        newState: {
                          type: 'string',
                          enum: [
                            'build_pending',
                            'building',
                            'pending',
                            'pulling',
                            'pending_validation',
                            'validating',
                            'active',
                            'inactive',
                            'error',
                            'build_failed',
                            'removing',
                          ],
                        },
                      },
                    },
                  },
                },
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
                schema: {
                  type: 'object',
                  properties: {
                    event: { type: 'string', example: 'snapshot.removed' },
                    timestamp: { type: 'string', format: 'date-time' },
                    data: {
                      type: 'string',
                      description: 'The ID of the removed snapshot',
                      example: 'snapshot123',
                    },
                  },
                },
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
                schema: {
                  type: 'object',
                  properties: {
                    event: { type: 'string', example: 'volume.created' },
                    timestamp: { type: 'string', format: 'date-time' },
                    data: {
                      type: 'object',
                      properties: {
                        id: { type: 'string', example: 'vol-12345678' },
                        name: { type: 'string', example: 'my-volume' },
                        organizationId: { type: 'string', example: '123e4567-e89b-12d3-a456-426614174000' },
                        state: {
                          type: 'string',
                          enum: [
                            'creating',
                            'ready',
                            'pending_create',
                            'pending_delete',
                            'deleting',
                            'deleted',
                            'error',
                          ],
                        },
                        createdAt: { type: 'string', format: 'date-time' },
                        updatedAt: { type: 'string', format: 'date-time' },
                        lastUsedAt: { type: 'string', format: 'date-time' },
                        errorReason: { type: 'string', example: 'Error processing volume' },
                      },
                    },
                  },
                },
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
                schema: {
                  type: 'object',
                  properties: {
                    event: { type: 'string', example: 'volume.state.updated' },
                    timestamp: { type: 'string', format: 'date-time' },
                    data: {
                      type: 'object',
                      properties: {
                        volume: {
                          type: 'object',
                          properties: {
                            id: { type: 'string', example: 'vol-12345678' },
                            organizationId: { type: 'string', example: '123e4567-e89b-12d3-a456-426614174000' },
                            state: {
                              type: 'string',
                              enum: [
                                'creating',
                                'ready',
                                'pending_create',
                                'pending_delete',
                                'deleting',
                                'deleted',
                                'error',
                              ],
                            },
                          },
                        },
                        oldState: {
                          type: 'string',
                          enum: [
                            'creating',
                            'ready',
                            'pending_create',
                            'pending_delete',
                            'deleting',
                            'deleted',
                            'error',
                          ],
                        },
                        newState: {
                          type: 'string',
                          enum: [
                            'creating',
                            'ready',
                            'pending_create',
                            'pending_delete',
                            'deleting',
                            'deleted',
                            'error',
                          ],
                        },
                      },
                    },
                  },
                },
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
