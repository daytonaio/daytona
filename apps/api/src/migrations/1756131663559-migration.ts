import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1756131663559 implements MigrationInterface {
  name = 'Migration1756131663559'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "region" ("code" character varying NOT NULL, "name" character varying NOT NULL, "organizationId" uuid NOT NULL, "dockerRegistryId" uuid, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "region_organizationId_name_unique" UNIQUE ("organizationId", "name"), CONSTRAINT "region_code_pk" PRIMARY KEY ("code"))`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "sshPass"`)
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum" RENAME TO "api_key_permissions_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'read:regions', 'write:regions', 'delete:regions', 'read:audit_logs')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum"[] USING "permissions"::"text"::"public"."api_key_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum_old"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "authToken" SET DEFAULT MD5(random()::text)`)
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum" RENAME TO "organization_role_permissions_enum_old"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'read:regions', 'write:regions', 'delete:regions', 'read:audit_logs')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum_old"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum_old" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'read:audit_logs')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum_old"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum"`)
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum_old" RENAME TO "organization_role_permissions_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "authToken" SET DEFAULT md5((random()))`)
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum_old" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'read:audit_logs')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum_old"[] USING "permissions"::"text"::"public"."api_key_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum_old" RENAME TO "api_key_permissions_enum"`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ADD "sshPass" character varying(32) NOT NULL DEFAULT replace((uuid_generate_v4()), '-', '')`,
    )
    await queryRunner.query(`DROP TABLE "region"`)
  }
}
