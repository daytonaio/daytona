/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1778588023059 implements MigrationInterface {
  name = 'Migration1778588023059'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "webhook_initialization" ADD "hasEndpoints" boolean NOT NULL DEFAULT false`)
    await queryRunner.query(`ALTER TABLE "webhook_initialization" ADD "endpointsCheckedAt" TIMESTAMP WITH TIME ZONE`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "webhook_initialization" DROP COLUMN "endpointsCheckedAt"`)
    await queryRunner.query(`ALTER TABLE "webhook_initialization" DROP COLUMN "hasEndpoints"`)
  }
}
