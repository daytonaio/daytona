/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Column,
  CreateDateColumn,
  Entity,
  Index,
  JoinColumn,
  ManyToOne,
  PrimaryGeneratedColumn,
  UpdateDateColumn,
} from 'typeorm'
import { Snapshot } from '../../sandbox/entities/snapshot.entity'

/**
 * SessionTemplate is the user-facing template entity (e.g. `python-default`).
 *
 * It links to an existing `Snapshot` via FK. The `Snapshot` entity itself is **not modified**
 * — a snapshot becomes "an session template" iff a row in this table points at it.
 *
 * NOTE on uniqueness: there is intentionally no `@Unique(['organizationId', 'name'])` decorator.
 * That decorator would emit a plain composite unique that treats every `null` organization_id
 * as distinct, which would silently allow duplicate general templates with the same name.
 * Uniqueness is enforced exclusively by the SQL `CREATE UNIQUE INDEX … (COALESCE(organization_id,
 * '00000000-...'), name)` declared in the migration. Don't "tidy this up" by adding `@Unique` —
 * doing so would change the semantics.
 */
@Entity('session_template')
@Index('session_template_org_id_idx', ['organizationId'])
@Index('session_template_general_idx', ['general'])
export class SessionTemplate {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  name: string

  @Column({ type: 'uuid', nullable: true })
  organizationId?: string

  @Column({ type: 'boolean', default: false })
  general = false

  @Column({ type: 'text', nullable: true })
  description?: string

  @Column({ array: true, type: 'text', default: () => 'ARRAY[]::text[]' })
  languages: string[] = []

  @Column({ array: true, type: 'text', nullable: true })
  packages?: string[]

  @Column({ type: 'uuid', unique: true })
  snapshotId: string

  @ManyToOne(() => Snapshot, { eager: false, onDelete: 'RESTRICT' })
  @JoinColumn({ name: 'snapshotId' })
  snapshot?: Snapshot

  @CreateDateColumn({ type: 'timestamp with time zone' })
  createdAt: Date

  @UpdateDateColumn({ type: 'timestamp with time zone' })
  updatedAt: Date
}
