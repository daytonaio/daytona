/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1755521645207 implements MigrationInterface {
  name = 'Migration1755521645207'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "sandbox_usage_periods_archive" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "sandboxId" character varying NOT NULL, "organizationId" character varying NOT NULL, "startAt" TIMESTAMP WITH TIME ZONE NOT NULL, "endAt" TIMESTAMP WITH TIME ZONE NOT NULL, "cpu" double precision NOT NULL, "gpu" double precision NOT NULL, "mem" double precision NOT NULL, "disk" double precision NOT NULL, "region" character varying NOT NULL, CONSTRAINT "sandbox_usage_periods_archive_id_pk" PRIMARY KEY ("id"))`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE "sandbox_usage_periods_archive"`)
  }
}
