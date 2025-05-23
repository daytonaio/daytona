/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { ImageRunnerState } from '../enums/image-runner-state.enum'

@Entity()
export class ImageRunner {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    type: 'enum',
    enum: ImageRunnerState,
    default: ImageRunnerState.PULLING_IMAGE,
  })
  state: ImageRunnerState

  @Column({ nullable: true })
  errorReason?: string

  @Column({
    //  todo: remove default
    default: '',
  })
  imageRef: string

  @Column()
  runnerId: string

  @CreateDateColumn()
  createdAt: Date

  @UpdateDateColumn()
  updatedAt: Date
}
