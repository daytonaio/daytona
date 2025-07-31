/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryColumn } from 'typeorm'
import { SystemRole } from './enums/system-role.enum'

export interface UserSSHKeyPair {
  privateKey: string
  publicKey: string
}

export interface UserPublicKey {
  key: string
  name: string
}

@Entity()
export class User {
  @PrimaryColumn()
  id: string

  @Column()
  name: string

  @Column({
    default: '',
  })
  email: string

  @Column({
    default: false,
  })
  emailVerified: boolean

  @Column({
    type: 'simple-json',
    nullable: true,
  })
  keyPair: UserSSHKeyPair

  @Column('simple-json')
  publicKeys: UserPublicKey[]

  @Column({
    type: 'enum',
    enum: SystemRole,
    default: SystemRole.USER,
  })
  role: SystemRole

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date
}
