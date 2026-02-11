/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1770000000003 implements MigrationInterface {
  name = 'Migration1770000000003'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Drop the old index
    await queryRunner.query(`DROP INDEX IF EXISTS "public"."warm_pool_find_idx"`)

    // Create the new index with organizationId included
    await queryRunner.query(
      `CREATE INDEX "warm_pool_find_idx" ON "warm_pool" ("organizationId", "snapshot", "target", "class", "cpu", "mem", "disk", "gpu", "osUser", "env")`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Drop the new index
    await queryRunner.query(`DROP INDEX IF EXISTS "public"."warm_pool_find_idx"`)

    // Recreate the old index without organizationId
    await queryRunner.query(
      `CREATE INDEX "warm_pool_find_idx" ON "warm_pool" ("snapshot", "target", "class", "cpu", "mem", "disk", "gpu", "osUser", "env")`,
    )
  }
}
