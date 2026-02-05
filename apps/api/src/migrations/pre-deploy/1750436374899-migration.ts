/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1750436374899 implements MigrationInterface {
  name = 'Migration1750436374899'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "audit_log" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "actorId" character varying NOT NULL, "actorEmail" character varying NOT NULL DEFAULT '', "organizationId" character varying, "action" character varying NOT NULL, "targetType" character varying, "targetId" character varying, "statusCode" integer, "errorMessage" character varying, "ipAddress" character varying, "userAgent" text, "source" character varying, "metadata" jsonb, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "audit_log_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `CREATE INDEX "audit_log_organizationId_createdAt_index" ON "audit_log" ("organizationId", "createdAt") `,
    )
    await queryRunner.query(`CREATE INDEX "audit_log_createdAt_index" ON "audit_log" ("createdAt") `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP INDEX "public"."audit_log_createdAt_index"`)
    await queryRunner.query(`DROP INDEX "public"."audit_log_organizationId_createdAt_index"`)
    await queryRunner.query(`DROP TABLE "audit_log"`)
  }
}
