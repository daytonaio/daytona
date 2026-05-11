/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { SessionTemplate } from './entities/session-template.entity'
import { SessionTemplateService } from './services/session-template.service'
import { SessionRepository } from './services/session-repository.service'
import { SessionInstanceStore } from './services/session-instance-store.service'
import { SessionGcService } from './services/session-gc.service'
import { SessionPoolService } from './services/session-pool.service'
import { SessionLoadService } from './services/session-load.service'
import { SessionScheduler } from './services/session-scheduler.service'
import { SessionService } from './services/session.service'
import { SessionController } from './controllers/session.controller'
import { SandboxModule } from '../sandbox/sandbox.module'
import { OrganizationModule } from '../organization/organization.module'
import { Sandbox } from '../sandbox/entities/sandbox.entity'

/**
 * SessionModule wires up the sessions product. Existing modules are imported but never edited:
 *  - SandboxModule provides SandboxService, RedisLockProvider (via export).
 *  - OrganizationModule supplies OrganizationAuthContextGuard.
 *  - The Sandbox entity is registered here via TypeOrmModule.forFeature so the pool service
 *    can read sandbox state for reconcile (without owning the sandbox repository).
 *
 * Session and SessionInstance live entirely in Redis (SessionRepository / SessionInstanceStore);
 * only SessionTemplate remains a Postgres entity.
 */
@Module({
  imports: [SandboxModule, OrganizationModule, TypeOrmModule.forFeature([SessionTemplate, Sandbox])],
  controllers: [SessionController],
  providers: [
    SessionTemplateService,
    SessionRepository,
    SessionInstanceStore,
    SessionGcService,
    SessionLoadService,
    SessionScheduler,
    SessionPoolService,
    SessionService,
  ],
  exports: [SessionService, SessionRepository, SessionPoolService, SessionTemplateService],
})
export class SessionModule {}
