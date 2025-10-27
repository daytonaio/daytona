/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { AdminDockerRegistryController } from './controllers/docker-registry.controller'
import { AdminRunnerController } from './controllers/runner.controller'
import { DockerRegistryModule } from '../docker-registry/docker-registry.module'
import { SandboxModule } from '../sandbox/sandbox.module'

@Module({
  imports: [SandboxModule, DockerRegistryModule],
  controllers: [AdminDockerRegistryController, AdminRunnerController],
})
export class AdminModule {}
