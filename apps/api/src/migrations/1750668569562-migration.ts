/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1750668569562 implements MigrationInterface {
  name = 'Migration1750668569562'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "autoDeleteInterval" integer NOT NULL DEFAULT '-1'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "autoDeleteInterval"`)
  }
}
