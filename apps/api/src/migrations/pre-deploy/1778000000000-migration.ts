/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1778000000000 implements MigrationInterface {
  name = 'Migration1778000000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "sandboxClass" character varying NOT NULL DEFAULT 'container'`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "sandboxClass" character varying NOT NULL DEFAULT 'container'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "sandboxClass" character varying NOT NULL DEFAULT 'container'`)
    await queryRunner.query(
      `ALTER TABLE "sandbox_usage_periods" ADD "sandboxClass" character varying NOT NULL DEFAULT 'container'`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox_usage_periods_archive" ADD "sandboxClass" character varying NOT NULL DEFAULT 'container'`,
    )
    await queryRunner.query(
      `ALTER TABLE "region_quota" ADD "sandboxClass" character varying NOT NULL DEFAULT 'container'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "region_quota" DROP COLUMN "sandboxClass"`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods_archive" DROP COLUMN "sandboxClass"`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" DROP COLUMN "sandboxClass"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "sandboxClass"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "sandboxClass"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "sandboxClass"`)
  }
}
