/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { GlobalRegionsIds } from '../sandbox/constants/global-regions.constant'
import { GLOBAL_REGION_ORGANIZATION_ID } from '../sandbox/constants/region.constants'

export class Migration1757518958305 implements MigrationInterface {
  name = 'Migration1757518958305'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // regions table
    await queryRunner.query(
      `CREATE TABLE "region" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "name" character varying NOT NULL, "organizationId" uuid NOT NULL, "dockerRegistryId" uuid, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "region_organizationId_name_unique" UNIQUE ("organizationId", "name"), CONSTRAINT "region_id_pk" PRIMARY KEY ("id"))`,
    )

    // global regions
    await queryRunner.query(
      `INSERT INTO "region" ("id", "name", "organizationId", "dockerRegistryId") VALUES ('${GlobalRegionsIds.US}', 'us', '${GLOBAL_REGION_ORGANIZATION_ID}', NULL)`,
    )
    await queryRunner.query(
      `INSERT INTO "region" ("id", "name", "organizationId", "dockerRegistryId") VALUES ('${GlobalRegionsIds.EU}', 'eu', '${GLOBAL_REGION_ORGANIZATION_ID}', NULL)`,
    )

    // switch runner region reference to id
    await queryRunner.renameColumn('runner', 'region', 'regionId')
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "regionId" DROP DEFAULT`)

    // update api key permission enum
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum" RENAME TO "api_key_permissions_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'read:regions', 'write:regions', 'delete:regions', 'read:runners', 'write:runners', 'delete:runners', 'read:audit_logs')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum"[] USING "permissions"::"text"::"public"."api_key_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum_old"`)

    // update organization role permission enum
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum" RENAME TO "organization_role_permissions_enum_old"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'read:regions', 'write:regions', 'delete:regions', 'read:runners', 'write:runners', 'delete:runners', 'read:audit_logs')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum_old"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // revert organization role permission enum
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum_old" AS ENUM('delete:registries', 'delete:sandboxes', 'delete:snapshots', 'delete:volumes', 'read:audit_logs', 'read:volumes', 'write:registries', 'write:sandboxes', 'write:snapshots', 'write:volumes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum_old"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum"`)
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum_old" RENAME TO "organization_role_permissions_enum"`,
    )

    // revert api key permission enum
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum_old" AS ENUM('delete:registries', 'delete:sandboxes', 'delete:snapshots', 'delete:volumes', 'read:audit_logs', 'read:volumes', 'write:registries', 'write:sandboxes', 'write:snapshots', 'write:volumes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum_old"[] USING "permissions"::"text"::"public"."api_key_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum_old" RENAME TO "api_key_permissions_enum"`)

    // revert to region varchar
    await queryRunner.renameColumn('runner', 'regionId', 'region')
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "region" SET DEFAULT 'us'`)

    // drop region table
    await queryRunner.query(`DROP TABLE "region"`)
  }
}
