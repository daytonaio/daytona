/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1779840000000 implements MigrationInterface {
  name = 'Migration1779840000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "region_quota" ADD "total_gpu_quota" integer NOT NULL DEFAULT 0`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "region_quota" DROP COLUMN "total_gpu_quota"`)
  }
}
