/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, Entity, PrimaryGeneratedColumn } from 'typeorm'

@Entity()
export class Team {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  name: string
}
