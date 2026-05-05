/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1778500000000 implements MigrationInterface {
  name = 'Migration1778500000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TYPE "public"."sandbox_class_enum_new" ADD VALUE IF NOT EXISTS 'container'`)
  }

  // PostgreSQL does not support removing a value from an enum type, so this
  // migration is one-way. Recreating the enum without 'container' would
  // require rewriting every column that references it, which isn't worth it
  // for a forward-only schema fix.
  public async down(): Promise<void> {
    // intentional no-op
  }
}
