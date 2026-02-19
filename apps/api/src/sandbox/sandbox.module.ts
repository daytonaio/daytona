/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { DataSource } from 'typeorm'
import { SandboxController } from './controllers/sandbox.controller'
import { SandboxService } from './services/sandbox.service'
import { TypeOrmModule } from '@nestjs/typeorm'
import { Sandbox } from './entities/sandbox.entity'
import { UserModule } from '../user/user.module'
import { RunnerService } from './services/runner.service'
import { Runner } from './entities/runner.entity'
import { RunnerController } from './controllers/runner.controller'
import { ToolboxService } from './services/toolbox.deprecated.service'
import { DockerRegistryModule } from '../docker-registry/docker-registry.module'
import { SandboxManager } from './managers/sandbox.manager'
import { ToolboxController } from './controllers/toolbox.deprecated.controller'
import { Snapshot } from './entities/snapshot.entity'
import { SnapshotController } from './controllers/snapshot.controller'
import { SnapshotService } from './services/snapshot.service'
import { SnapshotManager } from './managers/snapshot.manager'
import { SnapshotRunner } from './entities/snapshot-runner.entity'
import { DockerRegistry } from '../docker-registry/entities/docker-registry.entity'
import { RedisLockProvider } from './common/redis-lock.provider'
import { OrganizationModule } from '../organization/organization.module'
import { SandboxWarmPoolService } from './services/sandbox-warm-pool.service'
import { WarmPool } from './entities/warm-pool.entity'
import { PreviewController } from './controllers/preview.controller'
import { SnapshotSubscriber } from './subscribers/snapshot.subscriber'
import { VolumeController } from './controllers/volume.controller'
import { VolumeService } from './services/volume.service'
import { VolumeManager } from './managers/volume.manager'
import { Volume } from './entities/volume.entity'
import { BuildInfo } from './entities/build-info.entity'
import { BackupManager } from './managers/backup.manager'
import { VolumeSubscriber } from './subscribers/volume.subscriber'
import { RunnerSubscriber } from './subscribers/runner.subscriber'
import { WorkspaceController } from './controllers/workspace.deprecated.controller'
import { RunnerAdapterFactory } from './runner-adapter/runnerAdapter'
import { SandboxStartAction } from './managers/sandbox-actions/sandbox-start.action'
import { SandboxStopAction } from './managers/sandbox-actions/sandbox-stop.action'
import { SandboxDestroyAction } from './managers/sandbox-actions/sandbox-destroy.action'
import { SandboxArchiveAction } from './managers/sandbox-actions/sandbox-archive.action'
import { SshAccess } from './entities/ssh-access.entity'
import { SandboxRepository } from './repositories/sandbox.repository'
import { ProxyCacheInvalidationService } from './services/proxy-cache-invalidation.service'
import { RegionModule } from '../region/region.module'
import { Region } from '../region/entities/region.entity'
import { SnapshotRegion } from './entities/snapshot-region.entity'
import { JobController } from './controllers/job.controller'
import { JobService } from './services/job.service'
import { JobStateHandlerService } from './services/job-state-handler.service'
import { Job } from './entities/job.entity'
import { SandboxLookupCacheInvalidationService } from './services/sandbox-lookup-cache-invalidation.service'
import { EventEmitter2 } from '@nestjs/event-emitter'

@Module({
  imports: [
    UserModule,
    DockerRegistryModule,
    OrganizationModule,
    RegionModule,
    TypeOrmModule.forFeature([
      Sandbox,
      Runner,
      Snapshot,
      BuildInfo,
      SnapshotRunner,
      SnapshotRegion,
      DockerRegistry,
      WarmPool,
      Volume,
      SshAccess,
      Region,
      Job,
    ]),
  ],
  controllers: [
    SandboxController,
    RunnerController,
    ToolboxController,
    SnapshotController,
    WorkspaceController,
    PreviewController,
    VolumeController,
    JobController,
  ],
  providers: [
    SandboxService,
    SandboxManager,
    BackupManager,
    SandboxWarmPoolService,
    RunnerService,
    ToolboxService,
    SnapshotService,
    ProxyCacheInvalidationService,
    SandboxLookupCacheInvalidationService,
    SnapshotManager,
    RedisLockProvider,
    SnapshotSubscriber,
    VolumeService,
    VolumeManager,
    VolumeSubscriber,
    RunnerSubscriber,
    RunnerAdapterFactory,
    SandboxStartAction,
    SandboxStopAction,
    SandboxDestroyAction,
    SandboxArchiveAction,
    JobService,
    JobStateHandlerService,
    {
      provide: SandboxRepository,
      inject: [DataSource, EventEmitter2, SandboxLookupCacheInvalidationService],
      useFactory: (
        dataSource: DataSource,
        eventEmitter: EventEmitter2,
        sandboxLookupCacheInvalidationService: SandboxLookupCacheInvalidationService,
      ) => new SandboxRepository(dataSource, eventEmitter, sandboxLookupCacheInvalidationService),
    },
  ],
  exports: [
    SandboxService,
    RunnerService,
    RedisLockProvider,
    SnapshotService,
    VolumeService,
    VolumeManager,
    SandboxRepository,
    RunnerAdapterFactory,
  ],
})
export class SandboxModule {}
