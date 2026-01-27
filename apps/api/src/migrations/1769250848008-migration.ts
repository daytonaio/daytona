/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1769250848008 implements MigrationInterface {
  name = 'Migration1769250848008'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD "actionLoadPoints" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "actionLoadPenalty" integer NOT NULL DEFAULT '0'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "actionLoadPenalty"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "actionLoadPoints"`)
  }
}
