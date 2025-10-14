/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1760436353372 implements MigrationInterface {
  name = 'Migration1760436353372'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD "currentCpuLoadAverage" double precision NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "runnerInfoError" character varying`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "runnerInfoError"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentCpuLoadAverage"`)
  }
}
