/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1757513754037 implements MigrationInterface {
  name = 'Migration1757513754037'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "docker_registry" ADD "region" character varying`)

    await queryRunner.query(
      `ALTER TYPE "public"."docker_registry_registrytype_enum" RENAME TO "docker_registry_registrytype_enum_old"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."docker_registry_registrytype_enum" AS ENUM('internal', 'organization', 'transient', 'backup')`,
    )
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "registryType" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "docker_registry" ALTER COLUMN "registryType" TYPE "public"."docker_registry_registrytype_enum" USING "registryType"::"text"::"public"."docker_registry_registrytype_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "registryType" SET DEFAULT 'internal'`)
    await queryRunner.query(`DROP TYPE "public"."docker_registry_registrytype_enum_old"`)

    // Create the base default registry by copying from the default internal one
    await queryRunner.query(`
          INSERT INTO public.docker_registry (
            name, url, username, password, "isDefault", project, "createdAt", "updatedAt", "registryType", "organizationId", "region"
          )
          SELECT
            'Backup Registry' AS name,
            url,
            username,
            password,
            "isDefault",
            project,
            now() AS "createdAt",
            now() AS "updatedAt",
            'backup' AS "registryType",
            "organizationId",
            "region"
          FROM public.docker_registry
          WHERE "registryType" = 'internal'
            AND "isDefault" = true
          ORDER BY "createdAt" ASC
          LIMIT 1
        `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      DELETE FROM public.docker_registry
      WHERE "registryType" = 'backup'
    `)

    await queryRunner.query(
      `CREATE TYPE "public"."docker_registry_registrytype_enum_old" AS ENUM('internal', 'organization', 'transient', 'backup')`,
    )
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "registryType" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "docker_registry" ALTER COLUMN "registryType" TYPE "public"."docker_registry_registrytype_enum_old" USING "registryType"::"text"::"public"."docker_registry_registrytype_enum_old"`,
    )
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "registryType" SET DEFAULT 'internal'`)
    await queryRunner.query(`DROP TYPE "public"."docker_registry_registrytype_enum"`)
    await queryRunner.query(
      `ALTER TYPE "public"."docker_registry_registrytype_enum_old" RENAME TO "docker_registry_registrytype_enum"`,
    )

    await queryRunner.query(`ALTER TABLE "docker_registry" DROP COLUMN "region"`)
  }
}
