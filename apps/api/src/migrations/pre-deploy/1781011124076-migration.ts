/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1781011124076 implements MigrationInterface {
  name = 'Migration1781011124076'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Additive, non-breaking: existing snapshots are warm and backfill to the default.
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "propagationFactor" double precision NOT NULL DEFAULT '0.333'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "propagationFactor"`)
  }
}
