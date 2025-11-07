/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { configuration } from '../config/configuration'

export class Migration1761912147638 implements MigrationInterface {
  name = 'Migration1761912147638'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "defaultRegion" character varying NULL`)
    await queryRunner.query(`UPDATE "organization" SET "defaultRegion" = '${configuration.defaultRegion}'`)
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "defaultRegion" SET NOT NULL`)

    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "region" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "target" DROP DEFAULT`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "defaultRegion"`)

    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "region" SET DEFAULT 'us'`)
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "target" SET DEFAULT 'us'`)
  }
}
