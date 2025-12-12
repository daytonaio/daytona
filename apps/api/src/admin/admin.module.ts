/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { AdminRunnerController } from './controllers/runner.controller'
import { SandboxModule } from '../sandbox/sandbox.module'
import { RegionModule } from '../region/region.module'

@Module({
  imports: [SandboxModule, RegionModule],
  controllers: [AdminRunnerController],
})
export class AdminModule {}
