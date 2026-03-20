/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1773936687553 implements MigrationInterface {
  name = 'Migration1773936687553'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Note: not using CONCURRENTLY because of reverting issue: https://github.com/typeorm/typeorm/issues/9981
    await queryRunner.query(`CREATE INDEX "idx_sandbox_recoverable" ON "sandbox" ("id") WHERE "recoverable" = true`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "public"."idx_sandbox_recoverable"`)
  }
}
