/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1759422602797 implements MigrationInterface {
  name = 'Migration1759422602797'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "blockSharedInfrastructure" boolean NOT NULL DEFAULT false`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "blockSharedInfrastructure"`)
  }
}
