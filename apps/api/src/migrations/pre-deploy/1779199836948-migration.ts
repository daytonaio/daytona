/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1779199836948 implements MigrationInterface {
  name = 'Migration1779199836948'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Note: not using CONCURRENTLY + skipping transactions because of reverting issue: https://github.com/typeorm/typeorm/issues/9981
    await queryRunner.query(
      `CREATE UNIQUE INDEX "sandbox_usage_periods_one_open_period_per_sandbox_idx" ON "sandbox_usage_periods" ("sandboxId") WHERE "endAt" IS NULL`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "public"."sandbox_usage_periods_one_open_period_per_sandbox_idx"`)
  }
}
