/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1768306129179 implements MigrationInterface {
  name = 'Migration1768306129179'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "public"."job_status_enum" AS ENUM('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED')`,
    )
    await queryRunner.query(`CREATE TYPE "public"."job_resourcetype_enum" AS ENUM('SANDBOX', 'SNAPSHOT', 'BACKUP')`)
    await queryRunner.query(
      `CREATE TABLE "job" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "version" integer NOT NULL, "type" character varying NOT NULL, "status" "public"."job_status_enum" NOT NULL DEFAULT 'PENDING', "runnerId" character varying NOT NULL, "resourceType" "public"."job_resourcetype_enum" NOT NULL, "resourceId" character varying NOT NULL, "payload" character varying, "resultMetadata" character varying, "traceContext" jsonb, "errorMessage" text, "startedAt" TIMESTAMP WITH TIME ZONE, "completedAt" TIMESTAMP WITH TIME ZONE, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "job_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `CREATE UNIQUE INDEX "IDX_UNIQUE_INCOMPLETE_JOB" ON "job" ("resourceType", "resourceId", "runnerId") WHERE "completedAt" IS NULL`,
    )
    await queryRunner.query(`CREATE INDEX "job_resourceType_resourceId_index" ON "job" ("resourceType", "resourceId") `)
    await queryRunner.query(`CREATE INDEX "job_status_createdAt_index" ON "job" ("status", "createdAt") `)
    await queryRunner.query(`CREATE INDEX "job_runnerId_status_index" ON "job" ("runnerId", "status") `)
    await queryRunner.query(`ALTER TABLE "runner" RENAME COLUMN "version" TO "apiVersion"`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "appVersion" character varying DEFAULT 'v0.0.0-dev'`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "domain" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" DROP CONSTRAINT "UQ_330d74ac3d0e349b4c73c62ad6d"`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "apiUrl" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "proxyUrl" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "cpu" TYPE double precision`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "memoryGiB" TYPE double precision`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "diskGiB" TYPE double precision`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "gpu" DROP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "gpuType" DROP NOT NULL`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "gpuType" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "gpu" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "diskGiB" TYPE integer`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "memoryGiB" TYPE integer`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "cpu" TYPE integer`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "proxyUrl" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "apiUrl" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ADD CONSTRAINT "UQ_330d74ac3d0e349b4c73c62ad6d" UNIQUE ("domain")`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "domain" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" RENAME COLUMN "apiVersion" TO "version"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "appVersion"`)
    await queryRunner.query(`DROP INDEX "public"."job_runnerId_status_index"`)
    await queryRunner.query(`DROP INDEX "public"."job_status_createdAt_index"`)
    await queryRunner.query(`DROP INDEX "public"."job_resourceType_resourceId_index"`)
    await queryRunner.query(`DROP INDEX "public"."IDX_UNIQUE_INCOMPLETE_JOB"`)
    await queryRunner.query(`DROP TABLE "job"`)
    await queryRunner.query(`DROP TYPE "public"."job_resourcetype_enum"`)
    await queryRunner.query(`DROP TYPE "public"."job_status_enum"`)
  }
}
