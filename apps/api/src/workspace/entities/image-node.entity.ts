/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { ImageNodeState } from '../enums/image-node-state.enum'

@Entity()
export class ImageNode {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    type: 'enum',
    enum: ImageNodeState,
    default: ImageNodeState.PULLING_IMAGE,
  })
  state: ImageNodeState

  @Column({ nullable: true })
  errorReason?: string

  @Column({
    //  todo: remove default
    default: '',
  })
  imageRef: string

  @Column()
  nodeId: string

  @CreateDateColumn()
  createdAt: Date

  @UpdateDateColumn()
  updatedAt: Date
}
