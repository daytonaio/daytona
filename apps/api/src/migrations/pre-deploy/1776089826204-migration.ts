/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1776089826204 implements MigrationInterface {
  name = 'Migration1776089826204'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "sandbox_fork" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "parentId" character varying NOT NULL, "childId" character varying NOT NULL, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "sandbox_fork_childId_unique" UNIQUE ("childId"), CONSTRAINT "sandbox_fork_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox_fork" ADD CONSTRAINT "sandbox_fork_parentId_fk" FOREIGN KEY ("parentId") REFERENCES "sandbox"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox_fork" ADD CONSTRAINT "sandbox_fork_childId_fk" FOREIGN KEY ("childId") REFERENCES "sandbox"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )
    await queryRunner.query(`CREATE TYPE "public"."runner_runnerclass_enum" AS ENUM('container', 'vm')`)
    await queryRunner.query(
      `ALTER TABLE "runner" ADD "runnerClass" "public"."runner_runnerclass_enum" NOT NULL DEFAULT 'container'`,
    )
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" ADD VALUE IF NOT EXISTS 'snapshotting'`)
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" ADD VALUE IF NOT EXISTS 'forking'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "sandbox_fork" DROP CONSTRAINT "sandbox_fork_childId_fk"`)
    await queryRunner.query(`ALTER TABLE "sandbox_fork" DROP CONSTRAINT "sandbox_fork_parentId_fk"`)
    await queryRunner.query(`DROP TABLE "sandbox_fork"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "runnerClass"`)
    await queryRunner.query(`DROP TYPE "public"."runner_runnerclass_enum"`)
  }
}
