/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1744378115901 implements MigrationInterface {
  name = 'Migration1744378115901'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "suspended" boolean NOT NULL DEFAULT false`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "suspensionReason" character varying`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "suspendedUntil" TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "suspendedAt" TIMESTAMP`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "suspendedAt"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "suspendedUntil"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "suspensionReason"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "suspended"`)
  }
}
