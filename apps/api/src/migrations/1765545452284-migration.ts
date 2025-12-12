/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1765545452284 implements MigrationInterface {
  name = 'Migration1765545452284'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "cpu" TYPE double precision`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "memoryGiB" TYPE double precision`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "diskGiB" TYPE double precision`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "diskGiB" TYPE integer`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "memoryGiB" TYPE integer`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "cpu" TYPE integer`)
  }
}
