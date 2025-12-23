/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1766415256696 implements MigrationInterface {
  name = 'Migration1766415256696'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // region snapshot manager fields
    await queryRunner.query(`ALTER TABLE "region" ADD "snapshotManagerUrl" character varying`)
    await queryRunner.query(`ALTER TABLE "region" ADD "snapshotManagerApiKeyHash" character varying`)

    // make docker registry username and password nullable
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "username" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "password" DROP NOT NULL`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // drop region snapshot manager fields
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "snapshotManagerApiKeyHash"`)
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "snapshotManagerUrl"`)

    // revert docker registry username and password to not nullable
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "password" SET DEFAULT ''`)
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "username" SET DEFAULT ''`)

    await queryRunner.query(`UPDATE "docker_registry" SET "password" = '' WHERE "password" IS NULL`)
    await queryRunner.query(`UPDATE "docker_registry" SET "username" = '' WHERE "username" IS NULL`)

    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "password" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "username" SET NOT NULL`)

    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "password" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "username" DROP DEFAULT`)
  }
}
