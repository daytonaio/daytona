/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1757513754038 implements MigrationInterface {
  name = 'Migration1757513754038'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "docker_registry" ADD "isFallback" boolean NOT NULL DEFAULT false`)

    // Update existing registries that have isDefault = true and region = null to be fallback registries
    await queryRunner.query(`
      UPDATE "docker_registry"
      SET "isFallback" = true
      WHERE "isDefault" = true AND "region" IS NULL AND "registryType" = 'backup'
    `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "docker_registry" DROP COLUMN "isFallback"`)
  }
}
