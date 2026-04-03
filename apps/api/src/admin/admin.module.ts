/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { AdminRunnerController } from './controllers/runner.controller'
import { AdminSandboxController } from './controllers/sandbox.controller'
import { AdminUserController } from './controllers/user.controller'
import { AdminWebhookController } from './controllers/webhook.controller'
import { AdminDockerRegistryController } from './controllers/docker-registry.controller'
import { AdminSnapshotController } from './controllers/snapshot.controller'
import { AdminAuditController } from './controllers/audit.controller'
import { SandboxModule } from '../sandbox/sandbox.module'
import { RegionModule } from '../region/region.module'
import { OrganizationModule } from '../organization/organization.module'
import { UserModule } from '../user/user.module'
import { WebhookModule } from '../webhook/webhook.module'
import { DockerRegistryModule } from '../docker-registry/docker-registry.module'
import { AuditModule } from '../audit/audit.module'

@Module({
  imports: [
    SandboxModule,
    RegionModule,
    OrganizationModule,
    UserModule,
    WebhookModule,
    DockerRegistryModule,
    AuditModule,
  ],
  controllers: [
    AdminRunnerController,
    AdminSandboxController,
    AdminUserController,
    AdminWebhookController,
    AdminDockerRegistryController,
    AdminSnapshotController,
    AdminAuditController,
  ],
})
export class AdminModule {}
