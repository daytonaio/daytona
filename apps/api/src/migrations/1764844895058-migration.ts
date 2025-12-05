/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1764844895058 implements MigrationInterface {
  name = 'Migration1764844895058'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // drop region hidden field
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "hidden"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // revert drop region hidden field
    await queryRunner.query(`ALTER TABLE "region" ADD "hidden" boolean NOT NULL DEFAULT false`)
  }
}
