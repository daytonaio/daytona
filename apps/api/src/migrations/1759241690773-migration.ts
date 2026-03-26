/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1759241690773 implements MigrationInterface {
  name = 'Migration1759241690773'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "used"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "capacity"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD "capacity" integer NOT NULL DEFAULT '1000'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "used" integer NOT NULL DEFAULT '0'`)
  }
}
