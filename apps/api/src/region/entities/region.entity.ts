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

  @Column({ nullable: true })
  proxyUrl: string | null

  @Column({ nullable: true })
  toolboxProxyUrl: string | null

  @Column({ nullable: true })
  proxyApiKeyHash: string | null

  @Column({ nullable: true })
  sshGatewayUrl: string | null

  @Column({ nullable: true })
  sshGatewayApiKeyHash: string | null

  @Column({ nullable: true })
  snapshotManagerUrl: string | null

  @Column({ nullable: true })
  snapshotManagerApiKeyHash: string | null

  constructor(params: {
    name: string
    enforceQuotas: boolean
    regionType: RegionType
    id?: string
    organizationId?: string | null
    proxyUrl?: string | null
    toolboxProxyUrl?: string | null
    sshGatewayUrl?: string | null
    proxyApiKeyHash?: string | null
    sshGatewayApiKeyHash?: string | null
    snapshotManagerUrl?: string | null
    snapshotManagerApiKeyHash?: string | null
  }) {
    this.name = params.name
    this.enforceQuotas = params.enforceQuotas
    this.regionType = params.regionType

    if (params.id) {
      this.id = params.id
    } else {
      this.id = this.name.toLowerCase() + '_' + nanoid(4)
    }
    if (params.organizationId) {
      this.organizationId = params.organizationId
    }

    this.proxyUrl = params.proxyUrl ?? null
    this.toolboxProxyUrl = params.toolboxProxyUrl ?? params.proxyUrl ?? null
    this.sshGatewayUrl = params.sshGatewayUrl ?? null
    this.proxyApiKeyHash = params.proxyApiKeyHash ?? null
    this.sshGatewayApiKeyHash = params.sshGatewayApiKeyHash ?? null
    this.snapshotManagerUrl = params.snapshotManagerUrl ?? null
    this.snapshotManagerApiKeyHash = params.snapshotManagerApiKeyHash ?? null
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
