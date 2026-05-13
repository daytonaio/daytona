/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1777700000004 implements MigrationInterface {
  name = 'Migration1777700000004'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Note: not using CONCURRENTLY + skipping transactions because of reverting issue: https://github.com/typeorm/typeorm/issues/9981
    await queryRunner.query(
      `CREATE INDEX "warm_pool_find_idx" ON "warm_pool" ("snapshot", "target", "cpu", "mem", "disk", "gpu", "osUser", "env")`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX IF EXISTS "public"."warm_pool_find_idx"`)
  }
}
