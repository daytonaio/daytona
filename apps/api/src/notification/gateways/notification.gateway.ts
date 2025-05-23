/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logger, OnModuleInit, UnauthorizedException } from '@nestjs/common'
import { WebSocketGateway, WebSocketServer, OnGatewayInit } from '@nestjs/websockets'
import { Server, Socket } from 'socket.io'
import { createAdapter } from '@socket.io/redis-adapter'
import { OrganizationService } from '../../organization/services/organization.service'
import { WorkspaceEvents } from '../../workspace/constants/workspace-events.constants'
import { WorkspaceState } from '../../workspace/enums/workspace-state.enum'
import { WorkspaceDto } from '../../workspace/dto/workspace.dto'
import { SnapshotDto } from '../../workspace/dto/snapshot.dto'
import { SnapshotEvents } from '../../workspace/constants/snapshot-events'
import { SnapshotState } from '../../workspace/enums/snapshot-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { JwtStrategy } from '../../auth/jwt.strategy'
import { VolumeEvents } from '../../workspace/constants/volume-events'
import { VolumeDto } from '../../workspace/dto/volume.dto'
import { VolumeState } from '../../workspace/enums/volume-state.enum'

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
        const organizations = await this.organizationService.findByUser(payload.sub)
        const organizationIds = organizations.map((organization) => organization.id)
        await socket.join(organizationIds)
        next()
      } catch (error) {
        next(new UnauthorizedException())
      }
    })
  }

  emitWorkspaceCreated(workspace: WorkspaceDto) {
    this.server.to(workspace.organizationId).emit(WorkspaceEvents.CREATED, workspace)
  }

  emitWorkspaceStateUpdated(workspace: WorkspaceDto, oldState: WorkspaceState, newState: WorkspaceState) {
    this.server.to(workspace.organizationId).emit(WorkspaceEvents.STATE_UPDATED, { workspace, oldState, newState })
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
}
