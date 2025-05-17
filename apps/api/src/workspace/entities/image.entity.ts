/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Column,
  CreateDateColumn,
  Entity,
  JoinColumn,
  ManyToOne,
  OneToMany,
  OneToOne,
  PrimaryGeneratedColumn,
  Unique,
  UpdateDateColumn,
} from 'typeorm'
import { ImageNode } from './image-node.entity'
import { ImageState } from '../enums/image-state.enum'
import { BuildInfo } from './build-info.entity'

@Entity()
@Unique(['organizationId', 'name'])
export class Image {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    nullable: true,
    type: 'uuid',
  })
  organizationId?: string

  //  general image is available to all organizations
  @Column({ default: false })
  general: boolean

  @Column()
  name: string

  @Column({ nullable: true })
  internalName?: string

  @Column({ default: true })
  enabled: boolean

  @Column({
    type: 'enum',
    enum: ImageState,
    default: ImageState.PENDING,
  })
  state: ImageState

  @Column({ nullable: true })
  errorReason?: string

  @Column({ type: 'float', nullable: true })
  size?: number

  @OneToMany(() => ImageNode, (node) => node.imageRef)
  nodes: ImageNode[]

  @Column({ array: true, type: 'text', nullable: true })
  entrypoint?: string[]

  @CreateDateColumn()
  createdAt: Date

  @UpdateDateColumn()
  updatedAt: Date

  @Column({ nullable: true })
  lastUsedAt: Date

  @ManyToOne(() => BuildInfo, (buildInfo) => buildInfo.images, {
    nullable: true,
    eager: true,
  })
  @JoinColumn()
  buildInfo?: BuildInfo

  @Column({ nullable: true })
  buildNodeId?: string
}
