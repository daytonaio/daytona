/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1764844895056 implements MigrationInterface {
  name = 'Migration1764844895056'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD "name" character varying`)
    await queryRunner.query(`UPDATE "runner" SET "name" = "id"`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "name" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ADD CONSTRAINT "runner_region_name_unique" UNIQUE ("region", "name")`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "name"`)
  }
}
