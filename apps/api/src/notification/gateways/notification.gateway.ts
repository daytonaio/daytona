/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logger, OnModuleInit, UnauthorizedException } from '@nestjs/common'
import { WebSocketGateway, WebSocketServer, OnGatewayInit } from '@nestjs/websockets'
import { Server, Socket } from 'socket.io'
import { createAdapter } from '@socket.io/redis-adapter'
import { OrganizationService } from '../../organization/services/organization.service'
import { SandboxEvents } from '../../sandbox/constants/sandbox-events.constants'
import { SandboxState } from '../../sandbox/enums/sandbox-state.enum'
import { SandboxDto } from '../../sandbox/dto/sandbox.dto'
import { SnapshotDto } from '../../sandbox/dto/snapshot.dto'
import { SnapshotEvents } from '../../sandbox/constants/snapshot-events'
import { SnapshotState } from '../../sandbox/enums/snapshot-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { JwtStrategy } from '../../auth/jwt.strategy'
import { VolumeEvents } from '../../sandbox/constants/volume-events'
import { VolumeDto } from '../../sandbox/dto/volume.dto'
import { VolumeState } from '../../sandbox/enums/volume-state.enum'
import { SandboxDesiredState } from '../../sandbox/enums/sandbox-desired-state.enum'
import { AuditLogEvents } from '../../audit/constants/audit-log-events.constant'
import { AuditLogDto } from '../../audit/dto/audit-log.dto'

@WebSocketGateway({
  path: '/api/socket.io/',
  transports: ['websocket'],
})
export class NotificationGateway implements OnGatewayInit, OnModuleInit {
  private readonly logger = new Logger(NotificationGateway.name)

  @WebSocketServer()
  server: Server

  constructor(
    private readonly jwtStrategy: JwtStrategy,
    private readonly organizationService: OrganizationService,
    @InjectRedis() private readonly redis: Redis,
  ) {}

  onModuleInit() {
    const pubClient = this.redis.duplicate()
    const subClient = pubClient.duplicate()
    this.server.adapter(createAdapter(pubClient, subClient))
    this.logger.debug('Socket.io initialized with Redis adapter')
  }

  afterInit(server: Server) {
    this.logger.debug('WebSocket Gateway initialized')

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
      } catch (error) {
        next(new UnauthorizedException())
      }
    })
  }

  emitSandboxCreated(sandbox: SandboxDto) {
    this.server.to(sandbox.organizationId).emit(SandboxEvents.CREATED, sandbox)
  }

  emitSandboxStateUpdated(sandbox: SandboxDto, oldState: SandboxState, newState: SandboxState) {
    this.server.to(sandbox.organizationId).emit(SandboxEvents.STATE_UPDATED, { sandbox, oldState, newState })
  }

  emitSandboxDesiredStateUpdated(
    sandbox: SandboxDto,
    oldDesiredState: SandboxDesiredState,
    newDesiredState: SandboxDesiredState,
  ) {
    this.server
      .to(sandbox.organizationId)
      .emit(SandboxEvents.DESIRED_STATE_UPDATED, { sandbox, oldDesiredState, newDesiredState })
  }

  emitSnapshotCreated(snapshot: SnapshotDto) {
    this.server.to(snapshot.organizationId).emit(SnapshotEvents.CREATED, snapshot)
  }

  emitSnapshotStateUpdated(snapshot: SnapshotDto, oldState: SnapshotState, newState: SnapshotState) {
    this.server
      .to(snapshot.organizationId)
      .emit(SnapshotEvents.STATE_UPDATED, { snapshot: snapshot, oldState, newState })
  }

  emitSnapshotEnabledToggled(snapshot: SnapshotDto) {
    this.server.to(snapshot.organizationId).emit(SnapshotEvents.ENABLED_TOGGLED, snapshot)
  }

  emitSnapshotRemoved(snapshot: SnapshotDto) {
    this.server.to(snapshot.organizationId).emit(SnapshotEvents.REMOVED, snapshot.id)
  }

  emitVolumeCreated(volume: VolumeDto) {
    this.server.to(volume.organizationId).emit(VolumeEvents.CREATED, volume)
  }

  emitVolumeStateUpdated(volume: VolumeDto, oldState: VolumeState, newState: VolumeState) {
    this.server.to(volume.organizationId).emit(VolumeEvents.STATE_UPDATED, { volume, oldState, newState })
  }

  emitVolumeLastUsedAtUpdated(volume: VolumeDto) {
    this.server.to(volume.organizationId).emit(VolumeEvents.LAST_USED_AT_UPDATED, volume)
  }

  emitAuditLogCreated(auditLog: AuditLogDto) {
    this.server.to(auditLog.organizationId).emit(AuditLogEvents.CREATED, auditLog)
  }

  emitAuditLogUpdated(auditLog: AuditLogDto) {
    this.server.to(auditLog.organizationId).emit(AuditLogEvents.UPDATED, auditLog)
  }
}
