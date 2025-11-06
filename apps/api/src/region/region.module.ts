/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { RegionController } from '../region/controllers/region.controller'
import { Region } from '../region/entities/region.entity'
import { RegionService } from '../region/services/region.service'
import { OrganizationModule } from '../organization/organization.module'
import { Runner } from '../sandbox/entities/runner.entity'

@Module({
  imports: [OrganizationModule, TypeOrmModule.forFeature([Region, Runner])],
  controllers: [RegionController],
  providers: [RegionService],
  exports: [RegionService],
})
export class RegionModule {}
