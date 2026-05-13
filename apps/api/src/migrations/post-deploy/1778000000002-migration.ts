/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1778000000002 implements MigrationInterface {
  name = 'Migration1778000000002'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "warm_pool" DROP COLUMN IF EXISTS "class"`)
    await queryRunner.query(`DROP TYPE IF EXISTS "public"."warm_pool_class_enum"`)

    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN IF EXISTS "class"`)
    await queryRunner.query(`DROP TYPE IF EXISTS "public"."runner_class_enum"`)

    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN IF EXISTS "class"`)
    await queryRunner.query(`DROP TYPE IF EXISTS "public"."sandbox_class_enum"`)

    await queryRunner.query(
      `ALTER TABLE "region_quota" DROP CONSTRAINT IF EXISTS "region_quota_organizationId_regionId_pk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "region_quota" ADD CONSTRAINT "region_quota_organizationId_regionId_sandboxClass_pk" PRIMARY KEY ("organizationId", "regionId", "sandboxClass")`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "region_quota" DROP CONSTRAINT "region_quota_organizationId_regionId_sandboxClass_pk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "region_quota" ADD CONSTRAINT "region_quota_organizationId_regionId_pk" PRIMARY KEY ("organizationId", "regionId")`,
    )

    await queryRunner.query(`CREATE TYPE "public"."sandbox_class_enum" AS ENUM('small', 'medium', 'large')`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "class" "public"."sandbox_class_enum" NOT NULL DEFAULT 'small'`)

    await queryRunner.query(`CREATE TYPE "public"."runner_class_enum" AS ENUM('small', 'medium', 'large')`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "class" "public"."runner_class_enum" NOT NULL DEFAULT 'small'`)

    await queryRunner.query(`CREATE TYPE "public"."warm_pool_class_enum" AS ENUM('small', 'medium', 'large')`)
    await queryRunner.query(
      `ALTER TABLE "warm_pool" ADD "class" "public"."warm_pool_class_enum" NOT NULL DEFAULT 'small'`,
    )
  }
}
