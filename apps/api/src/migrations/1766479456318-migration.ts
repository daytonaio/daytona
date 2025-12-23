/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1766479456318 implements MigrationInterface {
  name = 'Migration1766479456318'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP CONSTRAINT "UQ_330d74ac3d0e349b4c73c62ad6d"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD CONSTRAINT "UQ_330d74ac3d0e349b4c73c62ad6d" UNIQUE ("domain")`)
  }
}
