/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { GlobalOrganizationRolesIds } from '../organization/constants/global-organization-roles.constant'
import { OrganizationResourcePermission } from '../organization/enums/organization-resource-permission.enum'

export class Migration1751456907334 implements MigrationInterface {
  name = 'Migration1751456907334'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // update enums
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum" RENAME TO "api_key_permissions_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'read:audit_logs')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum"[] USING "permissions"::"text"::"public"."api_key_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum_old"`)
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum" RENAME TO "organization_role_permissions_enum_old"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'read:audit_logs')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum_old"`)

    // add auditor role
    await queryRunner.query(`
        INSERT INTO "organization_role" 
          ("id", "name", "description", "permissions", "isGlobal")
        VALUES 
          (
            '${GlobalOrganizationRolesIds.AUDITOR}',    
            'Auditor', 
            'Grants access to audit logs in the organization', 
            ARRAY[
              '${OrganizationResourcePermission.READ_AUDIT_LOGS}'
            ]::organization_role_permissions_enum[],
            TRUE
          )
      `)

    // update organization role foreign keys
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" DROP CONSTRAINT "organization_role_assignment_roleId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" ADD CONSTRAINT "organization_role_assignment_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" DROP CONSTRAINT "organization_role_assignment_invitation_roleId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" ADD CONSTRAINT "organization_role_assignment_invitation_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // delete auditor role
    await queryRunner.query(`DELETE FROM "organization_role" WHERE "id" = '${GlobalOrganizationRolesIds.AUDITOR}'`)

    // remove read:audit_logs permission from api keys and organization roles
    await queryRunner.query(
      `UPDATE "api_key" SET "permissions" = array_remove("permissions", '${OrganizationResourcePermission.READ_AUDIT_LOGS}')`,
    )
    await queryRunner.query(
      `UPDATE "organization_role" SET "permissions" = array_remove("permissions", '${OrganizationResourcePermission.READ_AUDIT_LOGS}')`,
    )

    // revert enums
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum_old" AS ENUM('delete:registries', 'delete:sandboxes', 'delete:snapshots', 'delete:volumes', 'read:volumes', 'write:registries', 'write:sandboxes', 'write:snapshots', 'write:volumes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum_old"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum"`)
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum_old" RENAME TO "organization_role_permissions_enum"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum_old" AS ENUM('delete:registries', 'delete:sandboxes', 'delete:snapshots', 'delete:volumes', 'read:volumes', 'write:registries', 'write:sandboxes', 'write:snapshots', 'write:volumes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum_old"[] USING "permissions"::"text"::"public"."api_key_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum_old" RENAME TO "api_key_permissions_enum"`)
  }
}
