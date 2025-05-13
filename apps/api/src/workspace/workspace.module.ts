/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module, MiddlewareConsumer, RequestMethod } from '@nestjs/common'
import { RequestLoggerMiddleware } from './middleware/request-logger.middleware'
import { WorkspaceController } from './controllers/workspace.controller'
import { WorkspaceService } from './services/workspace.service'
import { TypeOrmModule } from '@nestjs/typeorm'
import { Workspace } from './entities/workspace.entity'
import { UserModule } from '../user/user.module'
import { NodeService } from './services/node.service'
import { Node } from './entities/node.entity'
import { NodeController } from './controllers/node.controller'
import { NodeApiFactory } from './runner-api/runnerApi'
import { AuthModule } from '../auth/auth.module'
import { ToolboxService } from './services/toolbox.service'
import { DockerRegistryModule } from '../docker-registry/docker-registry.module'
import { WorkspaceManager } from './managers/workspace.manager'
import { ToolboxController } from './controllers/toolbox.controller'
import { Image } from './entities/image.entity'
import { ImageController } from './controllers/image.controller'
import { ImageService } from './services/image.service'
import { ImageManager } from './managers/image.manager'
import { DockerProvider } from './docker/docker-provider'
import { ImageNode } from './entities/image-node.entity'
import { DockerRegistry } from '../docker-registry/entities/docker-registry.entity'
import { WorkspaceSubscriber } from './subscribers/workspace.subscriber'
import { RedisLockProvider } from './common/redis-lock.provider'
import { OrganizationModule } from '../organization/organization.module'
import { WorkspaceWarmPoolService } from './services/workspace-warm-pool.service'
import { WarmPool } from './entities/warm-pool.entity'
import { PreviewController } from './controllers/preview.controller'
import { ImageSubscriber } from './subscribers/image.subscriber'
import { VolumeController } from './controllers/volume.controller'
import { VolumeService } from './services/volume.service'
import { VolumeManager } from './managers/volume.manager'
import { Volume } from './entities/volume.entity'
import { BuildInfo } from './entities/build-info.entity'
import { SnapshotManager } from './managers/snapshot.manager'

@Module({
  imports: [
    UserModule,
    AuthModule,
    DockerRegistryModule,
    OrganizationModule,
    TypeOrmModule.forFeature([Workspace, Node, Image, BuildInfo, ImageNode, DockerRegistry, WarmPool, Volume]),
  ],
  controllers: [
    WorkspaceController,
    NodeController,
    ToolboxController,
    ImageController,
    PreviewController,
    VolumeController,
  ],
  providers: [
    WorkspaceService,
    WorkspaceManager,
    SnapshotManager,
    WorkspaceWarmPoolService,
    NodeService,
    NodeApiFactory,
    ToolboxService,
    ImageService,
    ImageManager,
    DockerProvider,
    WorkspaceSubscriber,
    RedisLockProvider,
    ImageSubscriber,
    VolumeService,
    VolumeManager,
  ],
  exports: [WorkspaceService, NodeService, RedisLockProvider, ImageService, VolumeService, VolumeManager],
})
export class WorkspaceModule {
  configure(consumer: MiddlewareConsumer) {
    consumer.apply(RequestLoggerMiddleware).forRoutes({ path: 'workspace', method: RequestMethod.POST })
  }
}
