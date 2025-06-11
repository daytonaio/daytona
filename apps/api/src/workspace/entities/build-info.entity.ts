/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, OneToMany, PrimaryColumn, UpdateDateColumn, BeforeInsert } from 'typeorm'
import { Snapshot } from './snapshot.entity'
import { Workspace } from './workspace.entity'
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

  @OneToMany(() => Workspace, (workspace) => workspace.buildInfo)
  workspaces: Workspace[]

  @Column({ type: 'timestamp', default: () => 'CURRENT_TIMESTAMP' })
  lastUsedAt: Date

  @CreateDateColumn()
  createdAt: Date

  @UpdateDateColumn()
  updatedAt: Date

  @BeforeInsert()
  generateHash() {
    this.snapshotRef = generateBuildInfoHash(this.dockerfileContent, this.contextHashes)
  }
}
