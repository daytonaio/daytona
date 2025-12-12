/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1765290567312 implements MigrationInterface {
  name = 'Migration1765290567312'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "public"."job_type_enum" AS ENUM('CREATE_SANDBOX', 'START_SANDBOX', 'STOP_SANDBOX', 'DESTROY_SANDBOX', 'CREATE_BACKUP', 'BUILD_SNAPSHOT', 'PULL_SNAPSHOT', 'REMOVE_SNAPSHOT')`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."job_status_enum" AS ENUM('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED')`,
    )
    await queryRunner.query(`CREATE TYPE "public"."job_resourcetype_enum" AS ENUM('SANDBOX', 'SNAPSHOT', 'BACKUP')`)
    await queryRunner.query(
      `CREATE TABLE "job" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "version" integer NOT NULL, "type" "public"."job_type_enum" NOT NULL, "status" "public"."job_status_enum" NOT NULL DEFAULT 'PENDING', "runnerId" character varying NOT NULL, "resourceType" "public"."job_resourcetype_enum", "resourceId" character varying, "payload" jsonb, "traceContext" jsonb, "errorMessage" text, "startedAt" TIMESTAMP WITH TIME ZONE, "completedAt" TIMESTAMP WITH TIME ZONE, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "job_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(`CREATE INDEX "job_status_index" ON "job" ("status") `)
    await queryRunner.query(`CREATE INDEX "job_runnerId_index" ON "job" ("runnerId") `)
    await queryRunner.query(`CREATE INDEX "job_resourceType_resourceId_index" ON "job" ("resourceType", "resourceId") `)
    await queryRunner.query(`CREATE INDEX "job_status_createdAt_index" ON "job" ("status", "createdAt") `)
    await queryRunner.query(`CREATE INDEX "job_runnerId_status_index" ON "job" ("runnerId", "status") `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "public"."job_runnerId_status_index"`)
    await queryRunner.query(`DROP INDEX "public"."job_status_createdAt_index"`)
    await queryRunner.query(`DROP INDEX "public"."job_resourceType_resourceId_index"`)
    await queryRunner.query(`DROP INDEX "public"."job_runnerId_index"`)
    await queryRunner.query(`DROP INDEX "public"."job_status_index"`)
    await queryRunner.query(`DROP TABLE "job"`)
    await queryRunner.query(`DROP TYPE "public"."job_resourcetype_enum"`)
    await queryRunner.query(`DROP TYPE "public"."job_status_enum"`)
    await queryRunner.query(`DROP TYPE "public"."job_type_enum"`)
  }
}
