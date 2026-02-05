/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { GlobalOrganizationRolesIds } from '../../organization/constants/global-organization-roles.constant'
import { OrganizationResourcePermission } from '../../organization/enums/organization-resource-permission.enum'

export class Migration1764844895057 implements MigrationInterface {
  name = 'Migration1764844895057'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // add region type field with its type and constraints
    await queryRunner.query(`CREATE TYPE "public"."region_regiontype_enum" AS ENUM('shared', 'dedicated', 'custom')`)
    await queryRunner.query(`ALTER TABLE "region" ADD "regionType" "public"."region_regiontype_enum"`)
    await queryRunner.query(
      `ALTER TABLE "region" ADD CONSTRAINT "region_not_custom" CHECK ("organizationId" IS NOT NULL OR "regionType" != 'custom')`,
    )
    await queryRunner.query(
      `ALTER TABLE "region" ADD CONSTRAINT "region_not_shared" CHECK ("organizationId" IS NULL OR "regionType" != 'shared')`,
    )
    await queryRunner.query(`UPDATE "region" SET "regionType" = 'custom' WHERE "organizationId" IS NOT NULL`)
    await queryRunner.query(`UPDATE "region" SET "regionType" = 'shared' WHERE "organizationId" IS NULL`)
    await queryRunner.query(`ALTER TABLE "region" ALTER COLUMN "regionType" SET NOT NULL`)

    // update api key permission enum
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum" RENAME TO "api_key_permissions_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'write:regions', 'delete:regions', 'read:runners', 'write:runners', 'delete:runners', 'read:audit_logs')`,
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
      `CREATE TYPE "public"."organization_role_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'write:regions', 'delete:regions', 'read:runners', 'write:runners', 'delete:runners', 'read:audit_logs')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum_old"`)

    // add infrastructure admin role
    await queryRunner.query(`
      INSERT INTO "organization_role" 
        ("id", "name", "description", "permissions", "isGlobal")
      VALUES 
        (
          '${GlobalOrganizationRolesIds.INFRASTRUCTURE_ADMIN}',    
          'Infrastructure Admin', 
          'Grants admin access to infrastructure in the organization', 
          ARRAY[
            '${OrganizationResourcePermission.WRITE_REGIONS}',
            '${OrganizationResourcePermission.DELETE_REGIONS}',
            '${OrganizationResourcePermission.READ_RUNNERS}',
            '${OrganizationResourcePermission.WRITE_RUNNERS}',
            '${OrganizationResourcePermission.DELETE_RUNNERS}'
          ]::organization_role_permissions_enum[],
          TRUE
        )
    `)

    // add runner name field
    await queryRunner.query(`ALTER TABLE "runner" ADD "name" character varying`)
    await queryRunner.query(`UPDATE "runner" SET "name" = "id"`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "name" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ADD CONSTRAINT "runner_region_name_unique" UNIQUE ("region", "name")`)

    // create new index for runner
    await queryRunner.query(
      `CREATE INDEX "runner_state_unschedulable_region_index" ON "runner" ("state", "unschedulable", "region") `,
    )

    // add region proxy and ssh gateway fields
    await queryRunner.query(`ALTER TABLE "region" ADD "proxyUrl" character varying`)
    await queryRunner.query(`ALTER TABLE "region" ADD "toolboxProxyUrl" character varying`)
    await queryRunner.query(`ALTER TABLE "region" ADD "proxyApiKeyHash" character varying`)
    await queryRunner.query(`ALTER TABLE "region" ADD "sshGatewayUrl" character varying`)
    await queryRunner.query(`ALTER TABLE "region" ADD "sshGatewayApiKeyHash" character varying`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "proxyUrl" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "region" DROP DEFAULT`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // remove region proxy and ssh gateway fields
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "sshGatewayApiKeyHash"`)
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "sshGatewayUrl"`)
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "proxyApiKeyHash"`)
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "toolboxProxyUrl"`)
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "proxyUrl"`)

    // drop region type field
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "regionType"`)
    await queryRunner.query(`DROP TYPE "public"."region_regiontype_enum"`)

    // remove infrastructure admin role
    await queryRunner.query(
      `DELETE FROM "organization_role" WHERE "id" = '${GlobalOrganizationRolesIds.INFRASTRUCTURE_ADMIN}'`,
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

    // drop runner name field
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "name"`)

    // drop new index for runner
    await queryRunner.query(`DROP INDEX "public"."runner_state_unschedulable_region_index"`)
  }
}
