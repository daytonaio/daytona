/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1778494331904 implements MigrationInterface {
  name = 'Migration1778494331904'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD "tags" text array NOT NULL DEFAULT '{}'`)
    await queryRunner.query(`CREATE INDEX "runner_tags_gin_idx" ON "runner" USING GIN ("tags")`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX IF EXISTS "runner_tags_gin_idx"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "tags"`)
  }
}
