/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1741088883000 implements MigrationInterface {
  name = 'Migration1741088883000'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // organizations
    await queryRunner.query(
      `CREATE TABLE "organization" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "name" character varying NOT NULL, "createdBy" character varying NOT NULL, "personal" boolean NOT NULL DEFAULT false, "telemetryEnabled" boolean NOT NULL DEFAULT true, "total_cpu_quota" integer NOT NULL DEFAULT '10', "total_memory_quota" integer NOT NULL DEFAULT '40', "total_disk_quota" integer NOT NULL DEFAULT '100', "max_cpu_per_workspace" integer NOT NULL DEFAULT '2', "max_memory_per_workspace" integer NOT NULL DEFAULT '4', "max_disk_per_workspace" integer NOT NULL DEFAULT '10', "max_concurrent_workspaces" integer NOT NULL DEFAULT '10', "workspace_quota" integer NOT NULL DEFAULT '0', "image_quota" integer NOT NULL DEFAULT '5', "max_image_size" integer NOT NULL DEFAULT '2', "total_image_size" integer NOT NULL DEFAULT '5', "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "organization_id_pk" PRIMARY KEY ("id"))`,
    )

    // organization users
    await queryRunner.query(`CREATE TYPE "public"."organization_user_role_enum" AS ENUM('owner', 'member')`)
    await queryRunner.query(
      `CREATE TABLE "organization_user" ("organizationId" uuid NOT NULL, "userId" character varying NOT NULL, "role" "public"."organization_user_role_enum" NOT NULL DEFAULT 'member', "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "organization_user_organizationId_userId_pk" PRIMARY KEY ("organizationId", "userId"))`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_user" ADD CONSTRAINT "organization_user_organizationId_fk" FOREIGN KEY ("organizationId") REFERENCES "organization"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )

    // organization invitations
    await queryRunner.query(`CREATE TYPE "public"."organization_invitation_role_enum" AS ENUM('owner', 'member')`)
    await queryRunner.query(
      `CREATE TYPE "public"."organization_invitation_status_enum" AS ENUM('pending', 'accepted', 'declined', 'cancelled')`,
    )
    await queryRunner.query(
      `CREATE TABLE "organization_invitation" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "organizationId" uuid NOT NULL, "email" character varying NOT NULL, "role" "public"."organization_invitation_role_enum" NOT NULL DEFAULT 'member', "expiresAt" TIMESTAMP NOT NULL, "status" "public"."organization_invitation_status_enum" NOT NULL DEFAULT 'pending', "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "organization_invitation_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_invitation" ADD CONSTRAINT "organization_invitation_organizationId_fk" FOREIGN KEY ("organizationId") REFERENCES "organization"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )

    // organization roles
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:images', 'delete:images', 'write:sandboxes', 'delete:sandboxes')`,
    )
    await queryRunner.query(
      `CREATE TABLE "organization_role" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "name" character varying NOT NULL, "description" character varying NOT NULL, "permissions" "public"."organization_role_permissions_enum" array NOT NULL, "isGlobal" boolean NOT NULL DEFAULT false, "organizationId" uuid, "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "organization_role_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ADD CONSTRAINT "organization_role_organizationId_fk" FOREIGN KEY ("organizationId") REFERENCES "organization"("id") ON DELETE CASCADE ON UPDATE NO ACTION`,
    )

    // organization role assignments for members
    await queryRunner.query(
      `CREATE TABLE "organization_role_assignment" ("organizationId" uuid NOT NULL, "userId" character varying NOT NULL, "roleId" uuid NOT NULL, CONSTRAINT "organization_role_assignment_organizationId_userId_roleId_pk" PRIMARY KEY ("organizationId", "userId", "roleId"))`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" ADD CONSTRAINT "organization_role_assignment_organizationId_userId_fk" FOREIGN KEY ("organizationId", "userId") REFERENCES "organization_user"("organizationId","userId") ON DELETE CASCADE ON UPDATE CASCADE`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" ADD CONSTRAINT "organization_role_assignment_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
    )
    await queryRunner.query(
      `CREATE INDEX "organization_role_assignment_organizationId_userId_index" ON "organization_role_assignment" ("organizationId", "userId") `,
    )
    await queryRunner.query(
      `CREATE INDEX "organization_role_assignment_roleId_index" ON "organization_role_assignment" ("roleId") `,
    )

    // organization role assignments for invitations
    await queryRunner.query(
      `CREATE TABLE "organization_role_assignment_invitation" ("invitationId" uuid NOT NULL, "roleId" uuid NOT NULL, CONSTRAINT "organization_role_assignment_invitation_invitationId_roleId_pk" PRIMARY KEY ("invitationId", "roleId"))`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" ADD CONSTRAINT "organization_role_assignment_invitation_invitationId_fk" FOREIGN KEY ("invitationId") REFERENCES "organization_invitation"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" ADD CONSTRAINT "organization_role_assignment_invitation_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
    )
    await queryRunner.query(
      `CREATE INDEX "organization_role_assignment_invitation_invitationId_index" ON "organization_role_assignment_invitation" ("invitationId") `,
    )
    await queryRunner.query(
      `CREATE INDEX "organization_role_assignment_invitation_roleId_index" ON "organization_role_assignment_invitation" ("roleId") `,
    )

    // create personal organizations
    await queryRunner.query(`
        INSERT INTO "organization" (
          name, 
          personal,
          "createdBy", 
          total_cpu_quota,
          total_memory_quota,
          total_disk_quota,
          max_cpu_per_workspace,
          max_memory_per_workspace,
          max_disk_per_workspace,
          max_concurrent_workspaces,
          workspace_quota,
          image_quota,
          max_image_size,
          total_image_size
        )
        SELECT 
          'Personal',
          true,
          u.id,
          u.total_cpu_quota,
          u.total_memory_quota,
          u.total_disk_quota,
          u.max_cpu_per_workspace,
          u.max_memory_per_workspace,
          u.max_disk_per_workspace,
          u.max_concurrent_workspaces,
          u.workspace_quota,
          u.image_quota,
          u.max_image_size,
          u.total_image_size
        FROM "user" u
    `)
    await queryRunner.query(`
        INSERT INTO "organization_user" ("organizationId", "userId", role)
        SELECT 
          o.id,
          o."createdBy",
          'owner'
        FROM "organization" o
        WHERE o.personal = true
    `)

    // drop user quotas
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "total_cpu_quota"`)
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "total_memory_quota"`)
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "total_disk_quota"`)
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "max_cpu_per_workspace"`)
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "max_memory_per_workspace"`)
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "max_disk_per_workspace"`)
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "max_concurrent_workspaces"`)
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "workspace_quota"`)
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "image_quota"`)
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "max_image_size"`)
    await queryRunner.query(`ALTER TABLE "user" DROP COLUMN "total_image_size"`)

    // move existing api keys to corresponding personal organizations
    await queryRunner.query(`ALTER TABLE "api_key" ADD "organizationId" uuid NULL`)
    await queryRunner.query(`
        UPDATE "api_key" ak
        SET "organizationId" = (
          SELECT o.id 
          FROM "organization" o
          WHERE o."createdBy" = ak."userId" 
          AND o.personal = true
          LIMIT 1
        )
    `)
    await queryRunner.query(`ALTER TABLE "api_key" ALTER COLUMN "organizationId" SET NOT NULL`)

    // update api key primary key
    await queryRunner.query(`
        DO $$
        DECLARE
            constraint_name text;
        BEGIN
            SELECT tc.constraint_name INTO constraint_name
            FROM information_schema.table_constraints tc
            WHERE tc.table_name = 'api_key' 
            AND tc.constraint_type = 'PRIMARY KEY';
            IF constraint_name IS NOT NULL THEN
            EXECUTE format('ALTER TABLE "api_key" DROP CONSTRAINT "%s"', constraint_name);
            END IF;
        END $$;
    `)
    await queryRunner.query(
      `ALTER TABLE "api_key" ADD CONSTRAINT "api_key_userId_name_organizationId_pk" PRIMARY KEY ("userId", "name", "organizationId")`,
    )

    // api key permissions
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:images', 'delete:images', 'write:sandboxes', 'delete:sandboxes')`,
    )
    await queryRunner.query(`ALTER TABLE "api_key" ADD "permissions" "public"."api_key_permissions_enum" array NULL`)
    await queryRunner.query(`
      UPDATE api_key
      SET permissions = ARRAY[
        'write:registries',
        'delete:registries', 
        'write:images',
        'delete:images',
        'write:sandboxes',
        'delete:sandboxes'
      ]::api_key_permissions_enum[]
    `)
    await queryRunner.query(`ALTER TABLE "api_key" ALTER COLUMN "permissions" SET NOT NULL`)

    // modify docker registry type enum
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "registryType" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TYPE "public"."docker_registry_registrytype_enum" RENAME TO "docker_registry_registrytype_enum_old"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."docker_registry_registrytype_enum" AS ENUM('internal', 'organization', 'public', 'transient')`,
    )
    await queryRunner.query(`
      CREATE OR REPLACE FUNCTION migrate_registry_type(old_type text) 
      RETURNS "public"."docker_registry_registrytype_enum" AS $$
      BEGIN
        IF old_type = 'user' THEN
          RETURN 'organization'::"public"."docker_registry_registrytype_enum";
        ELSE
          RETURN old_type::"public"."docker_registry_registrytype_enum";
        END IF;
      END;
      $$ LANGUAGE plpgsql;
    `)
    await queryRunner.query(`
      ALTER TABLE "docker_registry" 
      ALTER COLUMN "registryType" TYPE "public"."docker_registry_registrytype_enum" 
      USING migrate_registry_type("registryType"::text)
    `)
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "registryType" SET DEFAULT 'internal'`)
    await queryRunner.query(`DROP TYPE "public"."docker_registry_registrytype_enum_old"`)
    await queryRunner.query(`DROP FUNCTION migrate_registry_type`)

    // move existing docker registries to corresponding personal organizations
    await queryRunner.query(`ALTER TABLE "docker_registry" ADD "organizationId" uuid NULL`)
    await queryRunner.query(`UPDATE "docker_registry" SET "organizationId" = NULL WHERE "userId" = 'system'`)
    await queryRunner.query(`
        UPDATE "docker_registry" dr
        SET "organizationId" = (
          SELECT o.id 
          FROM "organization" o
          WHERE o."createdBy" = dr."userId" 
          AND o.personal = true
          LIMIT 1
        )
    `)
    await queryRunner.query(`ALTER TABLE "docker_registry" DROP COLUMN "userId"`)

    // move existing images to corresponding personal organizations
    await queryRunner.query(`ALTER TABLE "image" ADD "organizationId" uuid NULL`)
    await queryRunner.query(`
        UPDATE "image" i
        SET "organizationId" = (
          SELECT o.id 
          FROM "organization" o
          WHERE o."createdBy" = i."userId" 
          AND o.personal = true
          LIMIT 1
        )
    `)

    // update image unique constraint
    await queryRunner.query(`
        DO $$
        DECLARE
            constraint_name text;
        BEGIN
            SELECT tc.constraint_name INTO constraint_name
            FROM information_schema.table_constraints tc
            WHERE tc.table_name = 'image' 
            AND tc.constraint_type = 'UNIQUE' 
            AND tc.constraint_name LIKE '%name%';
            
            IF constraint_name IS NOT NULL THEN
            EXECUTE format('ALTER TABLE "image" DROP CONSTRAINT "%s"', constraint_name);
            END IF;
        END $$;
    `)
    await queryRunner.query(
      `ALTER TABLE "image" ADD CONSTRAINT "image_organizationId_name_unique" UNIQUE ("organizationId", "name")`,
    )
    await queryRunner.query(`ALTER TABLE "image" DROP COLUMN "userId"`)

    // move existing workspaces to corresponding personal organizations
    await queryRunner.query(`ALTER TABLE "workspace" ADD "organizationId" uuid NULL`)
    await queryRunner.query(`
        UPDATE "workspace" w
        SET "organizationId" = (
          SELECT o.id 
          FROM "organization" o
          WHERE o."createdBy" = w."userId" 
          AND o.personal = true
          LIMIT 1
        )
        WHERE w."userId" != 'unassigned'
    `)
    await queryRunner.query(`
        UPDATE "workspace" w
        SET "organizationId" = '00000000-0000-0000-0000-000000000000'
        WHERE w."userId" = 'unassigned'
    `)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "organizationId" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "userId"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // workspaces
    await queryRunner.query(`ALTER TABLE "workspace" ADD "userId" character varying NULL`)
    await queryRunner.query(`
        UPDATE "workspace" w
        SET "userId" = 'unassigned'
        WHERE w."organizationId" = '00000000-0000-0000-0000-000000000000'
    `)
    await queryRunner.query(`
        UPDATE "workspace" w
        SET "userId" = (
          SELECT o."createdBy" 
          FROM "organization" o
          WHERE o.id = w."organizationId"
        )
        WHERE w."organizationId" != '00000000-0000-0000-0000-000000000000'
    `)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "userId" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "organizationId"`)

    // images
    await queryRunner.query(`ALTER TABLE "image" ADD "userId" character varying NULL`)
    await queryRunner.query(`
        UPDATE "image" i
        SET "userId" = (
          SELECT o."createdBy" 
          FROM "organization" o
          WHERE o.id = i."organizationId"
        )
    `)
    await queryRunner.query(`ALTER TABLE "image" ALTER COLUMN "userId" SET NOT NULL`)
    await queryRunner.query(`ALTER TABLE "image" DROP CONSTRAINT "image_organizationId_name_unique"`)
    await queryRunner.query(`ALTER TABLE "image" ADD CONSTRAINT "image_userId_name_unique" UNIQUE ("userId", "name")`)
    await queryRunner.query(`ALTER TABLE "image" DROP COLUMN "organizationId"`)

    // docker registries
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "registryType" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TYPE "public"."docker_registry_registrytype_enum" RENAME TO "docker_registry_registrytype_enum_old"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."docker_registry_registrytype_enum" AS ENUM('internal', 'user', 'public', 'transient')`,
    )
    await queryRunner.query(`
    CREATE OR REPLACE FUNCTION rollback_registry_type(new_type text) 
    RETURNS "public"."docker_registry_registrytype_enum" AS $$
    BEGIN
      IF new_type = 'organization' THEN
        RETURN 'user'::"public"."docker_registry_registrytype_enum";
      ELSE
        RETURN new_type::"public"."docker_registry_registrytype_enum";
      END IF;
    END;
    $$ LANGUAGE plpgsql;
  `)
    await queryRunner.query(`
    ALTER TABLE "docker_registry" 
    ALTER COLUMN "registryType" TYPE "public"."docker_registry_registrytype_enum" 
    USING rollback_registry_type("registryType"::text)
  `)
    await queryRunner.query(`ALTER TABLE "docker_registry" ALTER COLUMN "registryType" SET DEFAULT 'internal'`)
    await queryRunner.query(`DROP TYPE "public"."docker_registry_registrytype_enum_old"`)
    await queryRunner.query(`DROP FUNCTION rollback_registry_type`)
    await queryRunner.query(`ALTER TABLE "docker_registry" ADD "userId" character varying NULL`)
    await queryRunner.query(`
        UPDATE "docker_registry" dr
        SET "userId" = (
          SELECT o."createdBy" 
          FROM "organization" o
          WHERE o.id = dr."organizationId"
        )
    `)
    await queryRunner.query(`ALTER TABLE "docker_registry" DROP COLUMN "organizationId"`)

    // api keys
    await queryRunner.query(`ALTER TABLE "api_key" DROP CONSTRAINT "api_key_userId_name_organizationId_pk"`)
    await queryRunner.query(
      `ALTER TABLE "api_key" ADD CONSTRAINT "api_key_userId_name_pk" PRIMARY KEY ("userId", "name")`,
    )
    await queryRunner.query(`ALTER TABLE "api_key" DROP COLUMN "organizationId"`)
    await queryRunner.query(`ALTER TABLE "api_key" DROP COLUMN "permissions"`)
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum"`)

    // user quotas
    await queryRunner.query(`ALTER TABLE "user" ADD "total_cpu_quota" integer NOT NULL DEFAULT '10'`)
    await queryRunner.query(`ALTER TABLE "user" ADD "total_memory_quota" integer NOT NULL DEFAULT '40'`)
    await queryRunner.query(`ALTER TABLE "user" ADD "total_disk_quota" integer NOT NULL DEFAULT '100'`)
    await queryRunner.query(`ALTER TABLE "user" ADD "max_cpu_per_workspace" integer NOT NULL DEFAULT '2'`)
    await queryRunner.query(`ALTER TABLE "user" ADD "max_memory_per_workspace" integer NOT NULL DEFAULT '4'`)
    await queryRunner.query(`ALTER TABLE "user" ADD "max_disk_per_workspace" integer NOT NULL DEFAULT '10'`)
    await queryRunner.query(`ALTER TABLE "user" ADD "max_concurrent_workspaces" integer NOT NULL DEFAULT '10'`)
    await queryRunner.query(`ALTER TABLE "user" ADD "workspace_quota" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "user" ADD "image_quota" integer NOT NULL DEFAULT '5'`)
    await queryRunner.query(`ALTER TABLE "user" ADD "max_image_size" integer NOT NULL DEFAULT '2'`)
    await queryRunner.query(`ALTER TABLE "user" ADD "total_image_size" integer NOT NULL DEFAULT '5'`)
    await queryRunner.query(`
        UPDATE "user" u
        SET 
          total_cpu_quota = (
            SELECT o.total_cpu_quota
            FROM "organization" o
            WHERE o."createdBy" = u.id
            AND o.personal = true
            LIMIT 1
          ),
          total_memory_quota = (
            SELECT o.total_memory_quota
            FROM "organization" o
            WHERE o."createdBy" = u.id
            AND o.personal = true
            LIMIT 1
          ),
          total_disk_quota = (
            SELECT o.total_disk_quota
            FROM "organization" o
            WHERE o."createdBy" = u.id
            AND o.personal = true
            LIMIT 1
          ),
          max_cpu_per_workspace = (
            SELECT o.max_cpu_per_workspace
            FROM "organization" o
            WHERE o."createdBy" = u.id
            AND o.personal = true
            LIMIT 1
          ),
          max_memory_per_workspace = (
            SELECT o.max_memory_per_workspace
            FROM "organization" o
            WHERE o."createdBy" = u.id
            AND o.personal = true
            LIMIT 1
          ),
          max_disk_per_workspace = (
            SELECT o.max_disk_per_workspace
            FROM "organization" o
            WHERE o."createdBy" = u.id
            AND o.personal = true
            LIMIT 1
          ),
          max_concurrent_workspaces = (
            SELECT o.max_concurrent_workspaces
            FROM "organization" o
            WHERE o."createdBy" = u.id
            AND o.personal = true
            LIMIT 1
          ),
          workspace_quota = (
            SELECT o.workspace_quota
            FROM "organization" o
            WHERE o."createdBy" = u.id
            AND o.personal = true
            LIMIT 1
          ),
          image_quota = (
            SELECT o.image_quota
            FROM "organization" o
            WHERE o."createdBy" = u.id
            AND o.personal = true
            LIMIT 1
          ),
          max_image_size = (
            SELECT o.max_image_size
            FROM "organization" o
            WHERE o."createdBy" = u.id
            AND o.personal = true
            LIMIT 1
          ),
          total_image_size = (
            SELECT o.total_image_size
            FROM "organization" o
            WHERE o."createdBy" = u.id
            AND o.personal = true
            LIMIT 1
          )
    `)

    // drop organization tables and related constraints
    await queryRunner.query(`DROP INDEX "organization_role_assignment_invitation_roleId_index"`)
    await queryRunner.query(`DROP INDEX "organization_role_assignment_invitation_invitationId_index"`)
    await queryRunner.query(`DROP TABLE "organization_role_assignment_invitation"`)
    await queryRunner.query(`DROP INDEX "organization_role_assignment_roleId_index"`)
    await queryRunner.query(`DROP INDEX "organization_role_assignment_organizationId_userId_index"`)
    await queryRunner.query(`DROP TABLE "organization_role_assignment"`)
    await queryRunner.query(`DROP TABLE "organization_role"`)
    await queryRunner.query(`DROP TYPE "organization_role_permissions_enum"`)
    await queryRunner.query(`DROP TABLE "organization_invitation"`)
    await queryRunner.query(`DROP TYPE "organization_invitation_status_enum"`)
    await queryRunner.query(`DROP TYPE "organization_invitation_role_enum"`)
    await queryRunner.query(`DROP TABLE "organization_user"`)
    await queryRunner.query(`DROP TYPE "organization_user_role_enum"`)
    await queryRunner.query(`DROP TABLE "organization"`)
  }
}
