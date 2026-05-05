/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1777973463276 implements MigrationInterface {
  name = 'Migration1777973463276'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "snapshot_quota" SET DEFAULT '30'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "snapshot_quota" SET DEFAULT '100'`)
  }
}
