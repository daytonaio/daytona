/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logger, OnModuleInit, UnauthorizedException } from '@nestjs/common'
import { WebSocketGateway, WebSocketServer, OnGatewayInit } from '@nestjs/websockets'
import { Server, Socket } from 'socket.io'
import { createAdapter } from '@socket.io/redis-adapter'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { JwtStrategy } from '../../auth/jwt.strategy'
import { WebhookService } from '../services/webhook.service'

@WebSocketGateway({
  path: '/api/webhook-socket.io/',
  transports: ['websocket'],
})
export class WebhookGateway implements OnGatewayInit, OnModuleInit {
  private readonly logger = new Logger(WebhookGateway.name)

  @WebSocketServer()
  server: Server

  constructor(
    private readonly jwtStrategy: JwtStrategy,
    private readonly webhookService: WebhookService,
    @InjectRedis() private readonly redis: Redis,
  ) {}

  onModuleInit() {
    const pubClient = this.redis.duplicate()
    const subClient = pubClient.duplicate()
    this.server.adapter(createAdapter(pubClient, subClient))
    this.logger.debug('Webhook WebSocket initialized with Redis adapter')
  }

  afterInit(server: Server) {
    this.logger.debug('Webhook WebSocket Gateway initialized')

    server.use(async (socket: Socket, next) => {
      const token = socket.handshake.auth.token
      if (!token) {
        return next(new UnauthorizedException())
      }

      try {
        const payload = await this.jwtStrategy.verifyToken(token)

        // Join the user room for user scoped notifications
        await socket.join(payload.sub)

        // Join the organization room for organization scoped notifications
        const organizationId = socket.handshake.query.organizationId as string | undefined
        if (organizationId) {
          await socket.join(organizationId)
        }

        next()
      } catch {
        next(new UnauthorizedException())
      }
    })
  }

  /**
   * Emit webhook delivery status to organization
   */
  emitWebhookDelivered(organizationId: string, messageId: string, endpointId: string, status: string) {
    this.server.to(organizationId).emit('webhook.delivered', {
      messageId,
      endpointId,
      status,
      timestamp: new Date().toISOString(),
    })
  }

  /**
   * Emit webhook delivery failed to organization
   */
  emitWebhookFailed(organizationId: string, messageId: string, endpointId: string, error: string) {
    this.server.to(organizationId).emit('webhook.failed', {
      messageId,
      endpointId,
      error,
      timestamp: new Date().toISOString(),
    })
  }

  /**
   * Emit webhook endpoint created to organization
   */
  emitEndpointCreated(organizationId: string, endpoint: any) {
    this.server.to(organizationId).emit('endpoint.created', {
      endpoint,
      timestamp: new Date().toISOString(),
    })
  }

  /**
   * Emit webhook endpoint deleted to organization
   */
  emitEndpointDeleted(organizationId: string, endpointId: string) {
    this.server.to(organizationId).emit('endpoint.deleted', {
      endpointId,
      timestamp: new Date().toISOString(),
    })
  }

  /**
   * Emit webhook endpoint updated to organization
   */
  emitEndpointUpdated(organizationId: string, endpoint: any) {
    this.server.to(organizationId).emit('endpoint.updated', {
      endpoint,
      timestamp: new Date().toISOString(),
    })
  }
}
