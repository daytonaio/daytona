/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

/**
 * Creates the durable `session_template` table — the only Postgres-backed part of the Sessions
 * feature. `Session` and `SessionInstance` live entirely in Redis (SessionRepository /
 * SessionInstanceStore), so no Postgres tables are created for them. `session_template` is general
 * config (seeded by the post-deploy python-default migration) and references `snapshot` via FK.
 */
export class Migration1778367241000 implements MigrationInterface {
  name = 'Migration1778367241000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      CREATE TABLE "session_template" (
        "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
        "name" varchar NOT NULL,
        "organizationId" uuid,
        "general" boolean NOT NULL DEFAULT false,
        "description" text,
        "languages" text[] NOT NULL DEFAULT ARRAY[]::text[],
        "packages" text[],
        "snapshotId" uuid NOT NULL,
        "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
        "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
        CONSTRAINT "PK_session_template" PRIMARY KEY ("id"),
        CONSTRAINT "UQ_session_template_snapshot" UNIQUE ("snapshotId"),
        CONSTRAINT "FK_session_template_snapshot"
          FOREIGN KEY ("snapshotId") REFERENCES "snapshot"("id") ON DELETE RESTRICT
      )
    `)
    await queryRunner.query(`
      CREATE UNIQUE INDEX "session_template_org_name_uidx"
      ON "session_template" (COALESCE("organizationId", '00000000-0000-0000-0000-000000000000'), "name")
    `)
    await queryRunner.query(`CREATE INDEX "session_template_org_id_idx" ON "session_template" ("organizationId")`)
    await queryRunner.query(`CREATE INDEX "session_template_general_idx" ON "session_template" ("general")`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE IF EXISTS "session_template"`)
  }
}
