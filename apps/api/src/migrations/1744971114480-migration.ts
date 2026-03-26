/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1744971114480 implements MigrationInterface {
  name = 'Migration1744971114480'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "workspace" ADD "volumes" jsonb NOT NULL DEFAULT '[]'`)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "authToken" SET DEFAULT MD5(random()::text)`)
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "state"`)
    await queryRunner.query(
      `CREATE TYPE "public"."volume_state_enum" AS ENUM('creating', 'ready', 'pending_create', 'pending_delete', 'deleting', 'deleted', 'error')`,
    )
    await queryRunner.query(
      `ALTER TABLE "volume" ADD "state" "public"."volume_state_enum" NOT NULL DEFAULT 'pending_create'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "state"`)
    await queryRunner.query(`DROP TYPE "public"."volume_state_enum"`)
    await queryRunner.query(`ALTER TABLE "volume" ADD "state" character varying NOT NULL`)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "authToken" SET DEFAULT md5((random()))`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "volumes"`)
  }
}
