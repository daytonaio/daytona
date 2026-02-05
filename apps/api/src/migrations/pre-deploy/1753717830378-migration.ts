/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1753717830378 implements MigrationInterface {
  name = 'Migration1753717830378'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "networkBlockAll" boolean NOT NULL DEFAULT false`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "networkAllowList" character varying`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "networkAllowList"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "networkBlockAll"`)
  }
}
