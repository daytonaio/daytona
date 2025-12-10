/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  BeforeInsert,
  BeforeUpdate,
  Check,
  Column,
  CreateDateColumn,
  Entity,
  Index,
  PrimaryColumn,
  UpdateDateColumn,
} from 'typeorm'
import { nanoid } from 'nanoid'
import { RegionType } from '../enums/region-type.enum'

@Entity()
@Index('region_organizationId_name_unique', ['organizationId', 'name'], {
  unique: true,
  where: '"organizationId" IS NOT NULL',
})
@Index('region_organizationId_null_name_unique', ['name'], {
  unique: true,
  where: '"organizationId" IS NULL',
})
@Check('region_not_shared', '"organizationId" IS NULL OR "regionType" != \'shared\'')
@Check('region_not_custom', '"organizationId" IS NOT NULL OR "regionType" != \'custom\'')
export class Region {
  @PrimaryColumn()
  id: string

  @Column()
  name: string

  @Column({
    type: 'uuid',
    nullable: true,
  })
  organizationId: string | null

  @Column({
    type: 'enum',
    enum: RegionType,
  })
  regionType: RegionType

  @Column({
    type: 'boolean',
    default: true,
  })
  enforceQuotas: boolean

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  constructor(name: string, enforceQuotas: boolean, regionType: RegionType, id?: string, organizationId?: string) {
    this.name = name
    this.enforceQuotas = enforceQuotas
    this.regionType = regionType

    if (id) {
      this.id = id
    } else {
      this.id = name.toLowerCase() + '_' + nanoid(4)
    }

    if (organizationId) {
      this.organizationId = organizationId
    }
  }

  @BeforeInsert()
  @BeforeUpdate()
  validateRegionType() {
    if (this.regionType === RegionType.SHARED) {
      if (this.organizationId) {
        throw new Error('Shared regions cannot be associated with an organization.')
      }
    }
    if (this.regionType === RegionType.CUSTOM) {
      if (!this.organizationId) {
        throw new Error('Custom regions must be associated with an organization.')
      }
    }
  }
}
