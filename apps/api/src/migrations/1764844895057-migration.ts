/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { GlobalOrganizationRolesIds } from '../organization/constants/global-organization-roles.constant'
import { OrganizationResourcePermission } from '../organization/enums/organization-resource-permission.enum'

export class Migration1764844895057 implements MigrationInterface {
  name = 'Migration1764844895057'

  public async up(queryRunner: QueryRunner): Promise<void> {
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

    // create new index for runner
    await queryRunner.query(
      `CREATE INDEX "runner_state_unschedulable_region_index" ON "runner" ("state", "unschedulable", "region") `,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
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

    // remove infrastructure admin role
    await queryRunner.query(
      `DELETE FROM "organization_role" WHERE "id" = '${GlobalOrganizationRolesIds.INFRASTRUCTURE_ADMIN}'`,
    )

    // drop new index for runner
    await queryRunner.query(`DROP INDEX "public"."runner_state_unschedulable_region_index"`)
  }
}
