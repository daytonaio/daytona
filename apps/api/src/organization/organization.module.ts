/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { OrganizationController } from './controllers/organization.controller'
import { OrganizationRoleController } from './controllers/organization-role.controller'
import { OrganizationUserController } from './controllers/organization-user.controller'
import { OrganizationInvitationController } from './controllers/organization-invitation.controller'
import { Organization } from './entities/organization.entity'
import { OrganizationRole } from './entities/organization-role.entity'
import { OrganizationUser } from './entities/organization-user.entity'
import { OrganizationInvitation } from './entities/organization-invitation.entity'
import { OrganizationService } from './services/organization.service'
import { OrganizationRoleService } from './services/organization-role.service'
import { OrganizationUserService } from './services/organization-user.service'
import { OrganizationInvitationService } from './services/organization-invitation.service'
import { UserModule } from '../user/user.module'
import { Workspace } from '../workspace/entities/workspace.entity'
import { Image } from '../workspace/entities/image.entity'
import { Volume } from '../workspace/entities/volume.entity'

@Module({
  imports: [
    UserModule,
    TypeOrmModule.forFeature([
      Organization,
      OrganizationRole,
      OrganizationUser,
      OrganizationInvitation,
      Workspace,
      Image,
      Volume,
    ]),
  ],
  controllers: [
    OrganizationController,
    OrganizationRoleController,
    OrganizationUserController,
    OrganizationInvitationController,
  ],
  providers: [OrganizationService, OrganizationRoleService, OrganizationUserService, OrganizationInvitationService],
  exports: [OrganizationService, OrganizationRoleService, OrganizationUserService, OrganizationInvitationService],
})
export class OrganizationModule {}
