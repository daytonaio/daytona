/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { configuration } from '../config/configuration'

export class Migration1764073472180 implements MigrationInterface {
  name = 'Migration1764073472180'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Remove defaultRegion column from organization table
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "defaultRegion"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Restore defaultRegion column to organization table
    await queryRunner.query(`ALTER TABLE "organization" ADD "defaultRegion" character varying NULL`)
    await queryRunner.query(`UPDATE "organization" SET "defaultRegion" = "defaultRegionId"`)
    await queryRunner.query(
      `ALTER TABLE "organization" ALTER COLUMN "defaultRegion" SET DEFAULT '${configuration.defaultRegion.id}'`,
    )
    await queryRunner.query(`ALTER TABLE "organization" ALTER COLUMN "defaultRegion" SET NOT NULL`)
  }
}
