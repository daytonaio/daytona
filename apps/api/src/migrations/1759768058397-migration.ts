/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1759768058397 implements MigrationInterface {
  name = 'Migration1759768058397'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "name" character varying`)
    await queryRunner.query(`UPDATE "sandbox" SET "name" = "id"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "name" SET NOT NULL`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "name" SET DEFAULT 'sandbox-' || substring(gen_random_uuid()::text, 1, 10)`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox" ADD CONSTRAINT "sandbox_organizationId_name_unique" UNIQUE ("organizationId", "name")`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" DROP CONSTRAINT "sandbox_organizationId_name_unique"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "name"`)
  }
}
