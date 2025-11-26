/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, Index, PrimaryColumn, UpdateDateColumn } from 'typeorm'
import { nanoid } from 'nanoid'

@Entity()
@Index('region_organizationId_name_unique', ['organizationId', 'name'], {
  unique: true,
  where: '"organizationId" IS NOT NULL',
})
@Index('region_organizationId_null_name_unique', ['name'], {
  unique: true,
  where: '"organizationId" IS NULL',
})
export class Region {
  @PrimaryColumn()
  id: string

  @Column()
  name: string

  @Column({
    type: 'uuid',
    nullable: true,
  })
  organizationId?: string

  @Column({ default: false })
  hidden: boolean

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

  constructor(params: {
    name: string
    id?: string
    organizationId?: string | null
    proxyUrl?: string | null
    toolboxProxyUrl?: string | null
    sshGatewayUrl?: string | null
    proxyApiKeyHash?: string | null
    sshGatewayApiKeyHash?: string | null
  }) {
    this.name = params.name
    if (params.id) {
      this.id = params.id
    } else {
      this.id = nanoid(12)
    }
    if (params.organizationId) {
      this.organizationId = params.organizationId
    }

    this.proxyUrl = params.proxyUrl ?? null
    this.toolboxProxyUrl = params.toolboxProxyUrl ?? params.proxyUrl ?? null
    this.sshGatewayUrl = params.sshGatewayUrl ?? null
    this.proxyApiKeyHash = params.proxyApiKeyHash ?? null
    this.sshGatewayApiKeyHash = params.sshGatewayApiKeyHash ?? null
  }
}
