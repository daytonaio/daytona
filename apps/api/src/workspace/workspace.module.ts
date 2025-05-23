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
import { RunnerService } from './services/runner.service'
import { Runner } from './entities/runner.entity'
import { RunnerController } from './controllers/runner.controller'
import { RunnerApiFactory } from './runner-api/runnerApi'
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
import { ImageRunner } from './entities/image-runner.entity'
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
import { BackupManager } from './managers/backup.manager'
import { VolumeSubscriber } from './subscribers/volume.subscriber'

@Module({
  imports: [
    UserModule,
    AuthModule,
    DockerRegistryModule,
    OrganizationModule,
    TypeOrmModule.forFeature([Workspace, Runner, Image, BuildInfo, ImageRunner, DockerRegistry, WarmPool, Volume]),
  ],
  controllers: [
    WorkspaceController,
    RunnerController,
    ToolboxController,
    ImageController,
    PreviewController,
    VolumeController,
  ],
  providers: [
    WorkspaceService,
    WorkspaceManager,
    BackupManager,
    WorkspaceWarmPoolService,
    RunnerService,
    RunnerApiFactory,
    ToolboxService,
    ImageService,
    ImageManager,
    DockerProvider,
    WorkspaceSubscriber,
    RedisLockProvider,
    ImageSubscriber,
    VolumeService,
    VolumeManager,
    VolumeSubscriber,
  ],
  exports: [WorkspaceService, RunnerService, RedisLockProvider, ImageService, VolumeService, VolumeManager],
})
export class WorkspaceModule {
  configure(consumer: MiddlewareConsumer) {
    consumer.apply(RequestLoggerMiddleware).forRoutes({ path: 'workspace', method: RequestMethod.POST })
  }
}
