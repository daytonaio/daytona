/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1752494676205 implements MigrationInterface {
  name = 'Migration1752494676205'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Convert runner.region from enum to varchar
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "region" TYPE varchar USING "region"::text`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "region" SET DEFAULT 'us'`)

    // Convert sandbox.region from enum to varchar
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "region" TYPE varchar USING "region"::text`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "region" SET DEFAULT 'us'`)

    // Convert warm_pool.target from enum to varchar
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "target" TYPE varchar USING "target"::text`)
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "target" SET DEFAULT 'us'`)

    // Drop the enum type if it exists
    await queryRunner.query(`DROP TYPE IF EXISTS "public"."runner_region_enum"`)
    await queryRunner.query(`DROP TYPE IF EXISTS "public"."sandbox_region_enum"`)
    await queryRunner.query(`DROP TYPE IF EXISTS "public"."warm_pool_target_enum"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Recreate the enum type
    await queryRunner.query(`CREATE TYPE "public"."warm_pool_target_enum" AS ENUM('eu', 'us', 'asia')`)
    await queryRunner.query(`CREATE TYPE "public"."sandbox_region_enum" AS ENUM('eu', 'us', 'asia')`)
    await queryRunner.query(`CREATE TYPE "public"."runner_region_enum" AS ENUM('eu', 'us', 'asia')`)

    // Convert back to enum for runner.region
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "region" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "runner" ALTER COLUMN "region" TYPE "public"."runner_region_enum" USING "region"::"public"."runner_region_enum"`,
    )

    // Convert back to enum for sandbox.region
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "region" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "region" TYPE "public"."sandbox_region_enum" USING "region"::"public"."sandbox_region_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "region" SET DEFAULT 'eu'`)

    // Convert back to enum for warm_pool.target
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "target" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "warm_pool" ALTER COLUMN "target" TYPE "public"."warm_pool_target_enum" USING "target"::"public"."warm_pool_target_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "target" SET DEFAULT 'eu'`)
  }
}
