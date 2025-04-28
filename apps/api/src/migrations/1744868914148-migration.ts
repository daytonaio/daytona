/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'
import { GlobalOrganizationRolesIds } from '../organization/constants/global-organization-roles.constant'
import { OrganizationResourcePermission } from '../organization/enums/organization-resource-permission.enum'

export class Migration1744868914148 implements MigrationInterface {
  name = 'Migration1744868914148'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // update enums
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum" RENAME TO "api_key_permissions_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:images', 'delete:images', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum"[] USING "permissions"::"text"::"public"."api_key_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum_old"`)
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum" RENAME TO "organization_role_permissions_enum_old"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:images', 'delete:images', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum_old"`)

    // add volumes admin role
    await queryRunner.query(`
            INSERT INTO "organization_role" 
              ("id", "name", "description", "permissions", "isGlobal")
            VALUES 
              (
                '${GlobalOrganizationRolesIds.VOLUMES_ADMIN}',    
                'Volumes Admin', 
                'Grants admin access to volumes in the organization', 
                ARRAY[
                  '${OrganizationResourcePermission.READ_VOLUMES}',
                  '${OrganizationResourcePermission.WRITE_VOLUMES}',
                  '${OrganizationResourcePermission.DELETE_VOLUMES}'
                ]::organization_role_permissions_enum[],
                TRUE
              )
          `)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // remove volumes admin role
    await queryRunner.query(
      `DELETE FROM "organization_role" WHERE "id" = '${GlobalOrganizationRolesIds.VOLUMES_ADMIN}'`,
    )

    // revert enums
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum_old" AS ENUM('write:registries', 'delete:registries', 'write:images', 'delete:images', 'write:sandboxes', 'delete:sandboxes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum_old"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum"`)
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum_old" RENAME TO "organization_role_permissions_enum"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum_old" AS ENUM('write:registries', 'delete:registries', 'write:images', 'delete:images', 'write:sandboxes', 'delete:sandboxes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum_old"[] USING "permissions"::"text"::"public"."api_key_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum_old" RENAME TO "api_key_permissions_enum"`)
  }
}
