/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1777170000000 implements MigrationInterface {
  name = 'Migration1777170000000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "volume" ADD "backend" character varying NOT NULL DEFAULT 's3fuse'`)
    await queryRunner.query(`ALTER TABLE "volume" ADD "archilDiskId" character varying`)
    await queryRunner.query(`ALTER TABLE "volume" ADD "archilRegion" character varying`)
    await queryRunner.query(`ALTER TABLE "volume" ADD "archilMountTokenEnc" character varying`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "archilMountTokenEnc"`)
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "archilRegion"`)
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "archilDiskId"`)
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "backend"`)
  }
}
