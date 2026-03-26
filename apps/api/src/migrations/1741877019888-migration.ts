/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1741877019888 implements MigrationInterface {
  name = 'Migration1741877019888'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`CREATE TYPE "public"."warm_pool_target_enum" AS ENUM('eu', 'us', 'asia')`)
    await queryRunner.query(`CREATE TYPE "public"."warm_pool_class_enum" AS ENUM('small', 'medium', 'large')`)
    await queryRunner.query(
      `CREATE TABLE "warm_pool" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "pool" integer NOT NULL, "image" character varying NOT NULL, "target" "public"."warm_pool_target_enum" NOT NULL DEFAULT 'eu', "cpu" integer NOT NULL, "mem" integer NOT NULL, "disk" integer NOT NULL, "gpu" integer NOT NULL, "gpuType" character varying NOT NULL, "class" "public"."warm_pool_class_enum" NOT NULL DEFAULT 'small', "osUser" character varying NOT NULL, "errorReason" character varying, "env" text NOT NULL DEFAULT '{}', "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "PK_fb06a13baeb3ac0ced145345d90" PRIMARY KEY ("id"))`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE "warm_pool"`)
    await queryRunner.query(`DROP TYPE "public"."warm_pool_class_enum"`)
    await queryRunner.query(`DROP TYPE "public"."warm_pool_target_enum"`)
  }
}
