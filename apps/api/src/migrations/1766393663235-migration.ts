/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1766393663235 implements MigrationInterface {
  name = 'Migration1766393663235'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "public"."job_status_index"`)
    await queryRunner.query(`DROP INDEX "public"."job_runnerId_index"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`CREATE INDEX "job_runnerId_index" ON "job" ("runnerId") `)
    await queryRunner.query(`CREATE INDEX "job_status_index" ON "job" ("status") `)
  }
}
