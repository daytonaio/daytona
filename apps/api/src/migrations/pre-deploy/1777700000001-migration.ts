/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1777700000001 implements MigrationInterface {
  name = 'Migration1777700000001'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Note: not using CONCURRENTLY + skipping transactions because of reverting issue: https://github.com/typeorm/typeorm/issues/9981
    await queryRunner.query(`CREATE INDEX "sandbox_linked_sandbox_id_idx" ON "sandbox" ("linkedSandboxId")`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "sandbox_linked_sandbox_id_idx"`)
  }
}
