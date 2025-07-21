/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1753185133351 implements MigrationInterface {
  name = 'Migration1753185133351'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD "version" character varying NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "proxyUrl" character varying NOT NULL DEFAULT ''`)
    // Copy apiUrl to proxyUrl for all existing records
    await queryRunner.query(`UPDATE "runner" SET "proxyUrl" = "apiUrl"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "version"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "proxyUrl"`)
  }
}
