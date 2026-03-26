/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { AdminRunnerController } from './controllers/runner.controller'
import { AdminSandboxController } from './controllers/sandbox.controller'
import { SandboxModule } from '../sandbox/sandbox.module'
import { RegionModule } from '../region/region.module'
import { OrganizationModule } from '../organization/organization.module'

@Module({
  imports: [SandboxModule, RegionModule, OrganizationModule],
  controllers: [AdminRunnerController, AdminSandboxController],
})
export class AdminModule {}
