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
  keyPair: UserSSHKeyPair | null

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

  constructor(
    id: string,
    name: string,
    email: string,
    emailVerified: boolean,
    keyPair: UserSSHKeyPair | null,
    publicKeys: UserPublicKey[],
    role: SystemRole = SystemRole.USER,
  ) {
    this.id = id
    this.name = name
    this.email = email
    this.emailVerified = emailVerified
    this.keyPair = keyPair
    this.publicKeys = publicKeys
    this.role = role
    this.createdAt = new Date()
  }
}
