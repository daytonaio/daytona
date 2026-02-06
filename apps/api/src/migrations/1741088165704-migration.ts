/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1741088165704 implements MigrationInterface {
  name = 'Migration1741088165704'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "image" DROP COLUMN "internalRegistryId"`)
    await queryRunner.query(
      `CREATE TYPE "public"."docker_registry_registrytype_enum" AS ENUM('internal', 'user', 'public', 'transient')`,
    )
    await queryRunner.query(
      `ALTER TABLE "docker_registry" ADD "registryType" "public"."docker_registry_registrytype_enum" NOT NULL DEFAULT 'internal'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "docker_registry" DROP COLUMN "registryType"`)
    await queryRunner.query(`DROP TYPE "public"."docker_registry_registrytype_enum"`)
    await queryRunner.query(`ALTER TABLE "image" ADD "internalRegistryId" character varying`)
  }
}
