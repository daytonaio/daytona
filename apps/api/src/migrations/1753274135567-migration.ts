/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1753274135567 implements MigrationInterface {
  name = 'Migration1753274135567'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "project" SET DEFAULT ''`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "project" DROP DEFAULT`)
  }
}
