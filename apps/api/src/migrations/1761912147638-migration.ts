/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { configuration } from '../config/configuration'

export class Migration1761912147638 implements MigrationInterface {
  name = 'Migration1761912147638'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "organization" ADD "defaultRegion" character varying NOT NULL DEFAULT '${configuration.defaultRegion}'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "defaultRegion"`)
  }
}
