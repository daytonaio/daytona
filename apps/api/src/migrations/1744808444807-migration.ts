/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1744808444807 implements MigrationInterface {
  name = 'Migration1744808444807'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "volume" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "organizationId" uuid, "name" character varying NOT NULL, "state" character varying NOT NULL, "errorReason" character varying, "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), "lastUsedAt" TIMESTAMP, CONSTRAINT "volume_organizationId_name_unique" UNIQUE ("organizationId", "name"), CONSTRAINT "volume_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "authToken" SET DEFAULT MD5(random()::text)`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "authToken" SET DEFAULT md5((random()))`)
    await queryRunner.query(`DROP TABLE "volume"`)
  }
}
