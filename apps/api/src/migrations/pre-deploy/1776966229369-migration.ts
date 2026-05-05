/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1776966229369 implements MigrationInterface {
  name = 'Migration1776966229369'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "job" ADD "payloadType" character varying NULL`)
    await queryRunner.query(`ALTER TABLE "job" ADD "resultType" character varying NULL`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "job" DROP COLUMN "resultType"`)
    await queryRunner.query(`ALTER TABLE "job" DROP COLUMN "payloadType"`)
  }
}
