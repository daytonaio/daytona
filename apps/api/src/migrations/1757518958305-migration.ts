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
    // add region table
    await queryRunner.query(
      `CREATE TABLE "region" ("id" character varying NOT NULL, "name" character varying NOT NULL, "organizationId" uuid NOT NULL, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "region_organizationId_name_unique" UNIQUE ("organizationId", "name"), CONSTRAINT "region_id_pk" PRIMARY KEY ("id"))`,
    )

    // insert global regions records
    await queryRunner.query(
      `INSERT INTO "region" ("id", "name", "organizationId") VALUES ('${GlobalRegionsIds.US}', 'us', '${GLOBAL_REGION_ORGANIZATION_ID}')`,
    )
    await queryRunner.query(
      `INSERT INTO "region" ("id", "name", "organizationId") VALUES ('${GlobalRegionsIds.EU}', 'eu', '${GLOBAL_REGION_ORGANIZATION_ID}')`,
    )

    // sandbox regionId reference
    await queryRunner.renameColumn('sandbox', 'region', 'regionId')
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "regionId" DROP DEFAULT`)

    // warm pool regionId reference
    await queryRunner.renameColumn('warm_pool', 'target', 'regionId')
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "regionId" DROP DEFAULT`)

    // runner regionId reference
    await queryRunner.renameColumn('runner', 'region', 'regionId')
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "regionId" DROP DEFAULT`)

    // usage period regionId reference
    await queryRunner.renameColumn('sandbox_usage_periods', 'region', 'regionId')

    // usage period archive regionId reference
    await queryRunner.renameColumn('sandbox_usage_periods_archive', 'region', 'regionId')

    // registry regionId reference
    await queryRunner.renameColumn('docker_registry', 'region', 'regionId')

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

    // revert sandbox region reference
    await queryRunner.renameColumn('sandbox', 'regionId', 'region')
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "region" SET DEFAULT 'us'`)

    // revert warm pool region reference
    await queryRunner.renameColumn('warm_pool', 'regionId', 'target')
    await queryRunner.query(`ALTER TABLE "warm_pool" ALTER COLUMN "target" SET DEFAULT 'us'`)

    // revert runner region reference
    await queryRunner.renameColumn('runner', 'regionId', 'region')
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "region" SET DEFAULT 'us'`)

    // revert usage period region reference
    await queryRunner.renameColumn('sandbox_usage_periods', 'regionId', 'region')

    // revert usage period archive region reference
    await queryRunner.renameColumn('sandbox_usage_periods_archive', 'regionId', 'region')

    // revert registry region reference
    await queryRunner.renameColumn('docker_registry', 'regionId', 'region')

    // drop region table
    await queryRunner.query(`DROP TABLE "region"`)
  }
}
