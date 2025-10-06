/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { v4 } from 'uuid'

@Entity()
export class WarmPool {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  pool: number

  @Column()
  snapshot: string

  @Column({
    default: 'us',
  })
  target: string

  @Column()
  cpu: number

  @Column()
  mem: number

  @Column()
  disk: number

  @Column()
  gpu: number

  @Column()
  gpuType: string

  @Column({
    type: 'enum',
    enum: SandboxClass,
    default: SandboxClass.SMALL,
  })
  class: SandboxClass

  @Column()
  osUser: string

  @Column({ nullable: true, type: String })
  errorReason: string | null

  @Column({
    type: 'simple-json',
    default: {},
  })
  env: { [key: string]: string }

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  constructor(createParams: {
    pool: number
    snapshot: string
    target: string
    cpu: number
    mem: number
    disk: number
    gpu: number
    gpuType: string
    class: SandboxClass
    osUser: string
    env?: { [key: string]: string }
  }) {
    this.id = v4()
    this.pool = createParams.pool
    this.snapshot = createParams.snapshot
    this.target = createParams.target
    this.cpu = createParams.cpu
    this.mem = createParams.mem
    this.disk = createParams.disk
    this.gpu = createParams.gpu
    this.gpuType = createParams.gpuType
    this.class = createParams.class
    this.osUser = createParams.osUser
    this.env = createParams.env || {}
    this.errorReason = null
    this.createdAt = new Date()
    this.updatedAt = new Date()
  }
}
