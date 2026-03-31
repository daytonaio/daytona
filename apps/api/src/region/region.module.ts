/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { DataSource } from 'typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { Region } from './entities/region.entity'
import { RegionService } from './services/region.service'
import { Runner } from '../sandbox/entities/runner.entity'
import { RegionController } from './controllers/region.controller'
import { SnapshotRepository } from '../sandbox/repositories/snapshot.repository'

@Module({
  imports: [TypeOrmModule.forFeature([Region, Runner])],
  controllers: [RegionController],
  providers: [
    RegionService,
    {
      provide: SnapshotRepository,
      inject: [DataSource, EventEmitter2],
      useFactory: (dataSource: DataSource, eventEmitter: EventEmitter2) =>
        new SnapshotRepository(dataSource, eventEmitter),
    },
  ],
  exports: [RegionService],
})
export class RegionModule {}
