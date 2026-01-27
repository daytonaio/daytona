/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1769516172576 implements MigrationInterface {
  name = 'Migration1769516172576'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Add RESIZING to sandbox_state_enum
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" ADD VALUE IF NOT EXISTS 'resizing'`)

    // Drop the resizing boolean column (no longer needed since we use state)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "resizing"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Re-add the resizing column
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "resizing" boolean NOT NULL DEFAULT false`)

    // Note: PostgreSQL does not support removing enum values directly
    // The 'resizing' value in sandbox_state_enum will remain but be unused
  }
}
