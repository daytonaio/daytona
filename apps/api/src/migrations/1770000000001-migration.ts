/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1770000000001 implements MigrationInterface {
  name = 'Migration1770000000001'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Create index for efficient querying of unassigned sandboxes
    await queryRunner.query(`CREATE INDEX "sandbox_unassigned_idx" ON "sandbox" ("unassigned")`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "public"."sandbox_unassigned_idx"`)
  }
}
