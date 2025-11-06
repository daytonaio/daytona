/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import * as crypto from 'crypto'

export class Migration1761912147645 implements MigrationInterface {
  name = 'Migration1761912147645'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // add region table
    await queryRunner.query(
      `CREATE TABLE "region" ("id" character varying NOT NULL, "name" character varying NOT NULL, "organizationId" uuid, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "region_organizationId_name_unique" UNIQUE ("organizationId", "name"), CONSTRAINT "region_id_pk" PRIMARY KEY ("id"))`,
    )

    // organization defaultRegionId reference
    await queryRunner.renameColumn('organization', 'defaultRegion', 'defaultRegionId')

    // sandbox regionId reference
    await queryRunner.renameColumn('sandbox', 'region', 'regionId')

    // warm pool regionId reference
    await queryRunner.renameColumn('warm_pool', 'target', 'regionId')

    // runner regionId reference
    await queryRunner.renameColumn('runner', 'region', 'regionId')
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "regionId" DROP DEFAULT`)

    // runner token
    await queryRunner.query(`ALTER TABLE "runner" ADD "tokenHash" character varying NOT NULL DEFAULT ''`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "tokenPrefix" character varying NOT NULL DEFAULT ''`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "tokenSuffix" character varying NOT NULL DEFAULT ''`)

    const existingRunners = await queryRunner.query(`SELECT "apiKey" FROM "runner"`)
    for (const runner of existingRunners) {
      const token = runner.apiKey
      const tokenHash = crypto.createHash('sha256').update(token).digest('hex')
      const tokenPrefix = token.substring(0, 3)
      const tokenSuffix = token.slice(-3)
      await queryRunner.query(
        `UPDATE "runner" SET "tokenHash" = $1, "tokenPrefix" = $2, "tokenSuffix" = $3 WHERE "apiKey" = $4`,
        [tokenHash, tokenPrefix, tokenSuffix, token],
      )
    }
    await queryRunner.query(`ALTER TABLE "runner" ADD CONSTRAINT "runner_tokenHash_unique" UNIQUE ("tokenHash")`)

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

    // revert tokenHash
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "tokenSuffix"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "tokenPrefix"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "tokenHash"`)

    // revert organization defaultRegionId reference
    await queryRunner.renameColumn('organization', 'defaultRegionId', 'defaultRegion')

    // revert sandbox region reference
    await queryRunner.renameColumn('sandbox', 'regionId', 'region')

    // revert warm pool region reference
    await queryRunner.renameColumn('warm_pool', 'regionId', 'target')

    // revert runner region reference
    await queryRunner.renameColumn('runner', 'regionId', 'region')
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "region" SET DEFAULT 'us'`)

    // drop region table
    await queryRunner.query(`DROP TABLE "region"`)
  }
}
