/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1766402443700 implements MigrationInterface {
  name = 'Migration1766402443700'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_runnerclass_enum" AS ENUM('linux', 'linux-exp', 'windows-exp')`,
    )
    await queryRunner.query(
      `ALTER TABLE "snapshot" ADD "runnerClass" "public"."snapshot_runnerclass_enum" NOT NULL DEFAULT 'linux'`,
    )
    await queryRunner.query(`ALTER TYPE "public"."runner_class_enum" RENAME TO "runner_class_enum_old"`)
    await queryRunner.query(`CREATE TYPE "public"."runner_class_enum" AS ENUM('linux', 'linux-exp', 'windows-exp')`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "class" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "runner" ALTER COLUMN "class" TYPE "public"."runner_class_enum" USING "class"::"text"::"public"."runner_class_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "class" SET DEFAULT 'linux'`)
    await queryRunner.query(`DROP TYPE "public"."runner_class_enum_old"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "runnerClass"`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_runnerclass_enum"`)
    await queryRunner.query(`CREATE TYPE "public"."runner_class_enum_old" AS ENUM('large', 'medium', 'small')`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "class" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "runner" ALTER COLUMN "class" TYPE "public"."runner_class_enum_old" USING "class"::"text"::"public"."runner_class_enum_old"`,
    )
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "class" SET DEFAULT 'small'`)
    await queryRunner.query(`DROP TYPE "public"."runner_class_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."runner_class_enum_old" RENAME TO "runner_class_enum"`)
  }
}
