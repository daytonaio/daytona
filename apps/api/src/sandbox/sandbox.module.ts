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
import { DockerProvider } from './docker/docker-provider'
import { SnapshotRunner } from './entities/snapshot-runner.entity'
import { DockerRegistry } from '../docker-registry/entities/docker-registry.entity'
import { SandboxSubscriber } from './subscribers/sandbox.subscriber'
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
import { WorkspaceController } from './controllers/workspace.deprecated.controller'
import { RunnerAdapterFactory } from './runner-adapter/runnerAdapter'
import { SandboxStartAction } from './managers/sandbox-actions/sandbox-start.action'
import { SandboxStopAction } from './managers/sandbox-actions/sandbox-stop.action'
import { SandboxDestroyAction } from './managers/sandbox-actions/sandbox-destroy.action'
import { SandboxArchiveAction } from './managers/sandbox-actions/sandbox-archive.action'
import { SshAccess } from './entities/ssh-access.entity'
import { SandboxRepository } from './repositories/sandbox.repository'

@Module({
  imports: [
    UserModule,
    DockerRegistryModule,
    OrganizationModule,
    TypeOrmModule.forFeature([
      Sandbox,
      Runner,
      Snapshot,
      BuildInfo,
      SnapshotRunner,
      DockerRegistry,
      WarmPool,
      Volume,
      SshAccess,
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
  ],
  providers: [
    SandboxService,
    SandboxManager,
    BackupManager,
    SandboxWarmPoolService,
    RunnerService,
    ToolboxService,
    SnapshotService,
    SnapshotManager,
    DockerProvider,
    SandboxSubscriber,
    RedisLockProvider,
    SnapshotSubscriber,
    VolumeService,
    VolumeManager,
    VolumeSubscriber,
    RunnerAdapterFactory,
    SandboxStartAction,
    SandboxStopAction,
    SandboxDestroyAction,
    SandboxArchiveAction,
    {
      provide: SandboxRepository,
      inject: [DataSource],
      useFactory: (dataSource: DataSource) => new SandboxRepository(dataSource),
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
  ],
})
export class SandboxModule {}
