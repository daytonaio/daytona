/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TeamService } from './team.service'
import { TypeOrmModule } from '@nestjs/typeorm'
import { Team } from './team.entity'

@Module({
  imports: [TypeOrmModule.forFeature([Team])],
  providers: [TeamService],
})
export class TeamModule {}
