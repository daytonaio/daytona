/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, OneToMany, PrimaryColumn, UpdateDateColumn, BeforeInsert } from 'typeorm'
import { Snapshot } from './snapshot.entity'
import { Sandbox } from './sandbox.entity'
import { createHash } from 'crypto'

export function generateBuildInfoHash(dockerfileContent: string, contextHashes: string[] = []): string {
  const sortedContextHashes = [...contextHashes].sort() || []
  const combined = dockerfileContent + sortedContextHashes.join('')
  const hash = createHash('sha256').update(combined).digest('hex')
  return 'daytona-' + hash + ':daytona'
}

@Entity()
export class BuildInfo {
  @PrimaryColumn()
  snapshotRef: string

  @Column({ type: 'text', nullable: true })
  dockerfileContent?: string

  @Column('simple-array', { nullable: true })
  contextHashes?: string[]

  @OneToMany(() => Snapshot, (snapshot) => snapshot.buildInfo)
  snapshots: Snapshot[]

  @OneToMany(() => Sandbox, (sandbox) => sandbox.buildInfo)
  sandboxes: Sandbox[]

  @Column({ type: 'timestamp with time zone', default: () => 'CURRENT_TIMESTAMP' })
  lastUsedAt: Date

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  @BeforeInsert()
  generateHash() {
    this.snapshotRef = generateBuildInfoHash(this.dockerfileContent, this.contextHashes)
  }
}
