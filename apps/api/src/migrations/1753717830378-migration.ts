import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1753717830378 implements MigrationInterface {
  name = 'Migration1753717830378'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "audit_log" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "actorId" character varying NOT NULL, "actorEmail" character varying NOT NULL DEFAULT '', "organizationId" character varying, "action" character varying NOT NULL, "targetType" character varying, "targetId" character varying, "statusCode" integer, "errorMessage" character varying, "ipAddress" character varying, "userAgent" text, "source" character varying, "metadata" jsonb, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "audit_log_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `CREATE INDEX "audit_log_organizationId_createdAt_index" ON "audit_log" ("organizationId", "createdAt") `,
    )
    await queryRunner.query(`CREATE INDEX "audit_log_createdAt_index" ON "audit_log" ("createdAt") `)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "memory"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "disk"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "networkAllowAll" boolean NOT NULL DEFAULT true`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "networkAllowList" character varying`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "proxyUrl" character varying NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "memoryGiB" integer NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "diskGiB" integer NOT NULL`)
    await queryRunner.query(
      `ALTER TABLE "runner" ADD "currentCpuUsagePercentage" double precision NOT NULL DEFAULT '0'`,
    )
    await queryRunner.query(
      `ALTER TABLE "runner" ADD "currentMemoryUsagePercentage" double precision NOT NULL DEFAULT '0'`,
    )
    await queryRunner.query(
      `ALTER TABLE "runner" ADD "currentDiskUsagePercentage" double precision NOT NULL DEFAULT '0'`,
    )
    await queryRunner.query(`ALTER TABLE "runner" ADD "currentAllocatedCpu" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "currentAllocatedMemoryGiB" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "currentAllocatedDiskGiB" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "currentSnapshotCount" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "availabilityScore" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "version" character varying NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum" RENAME TO "api_key_permissions_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'read:audit_logs')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum"[] USING "permissions"::"text"::"public"."api_key_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum_old"`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" DROP COLUMN "startAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" ADD "startAt" TIMESTAMP WITH TIME ZONE NOT NULL`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" DROP COLUMN "endAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" ADD "endAt" TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "warm_pool" DROP COLUMN "target"`)
    await queryRunner.query(`DROP TYPE "public"."warm_pool_target_enum"`)
    await queryRunner.query(`ALTER TABLE "warm_pool" ADD "target" character varying NOT NULL DEFAULT 'us'`)
    await queryRunner.query(`ALTER TABLE "warm_pool" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "warm_pool" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "warm_pool" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "warm_pool" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "volume" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "volume" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "snapshot_runner" DROP COLUMN "createdAt"`)
    await queryRunner.query(
      `ALTER TABLE "snapshot_runner" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot_runner" DROP COLUMN "updatedAt"`)
    await queryRunner.query(
      `ALTER TABLE "snapshot_runner" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "region"`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_region_enum"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "region" character varying NOT NULL DEFAULT 'us'`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "lastBackupAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "lastBackupAt" TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "lastActivityAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "lastActivityAt" TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "authToken" SET DEFAULT MD5(random()::text)`)
    await queryRunner.query(`ALTER TABLE "build_info" DROP COLUMN "lastUsedAt"`)
    await queryRunner.query(`ALTER TABLE "build_info" ADD "lastUsedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "build_info" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "build_info" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "build_info" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "build_info" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "region"`)
    await queryRunner.query(`DROP TYPE "public"."runner_region_enum"`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "region" character varying NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "lastChecked"`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "lastChecked" TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "organization_user" DROP COLUMN "createdAt"`)
    await queryRunner.query(
      `ALTER TABLE "organization_user" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
    await queryRunner.query(`ALTER TABLE "organization_user" DROP COLUMN "updatedAt"`)
    await queryRunner.query(
      `ALTER TABLE "organization_user" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
    await queryRunner.query(`ALTER TABLE "organization_invitation" DROP COLUMN "expiresAt"`)
    await queryRunner.query(`ALTER TABLE "organization_invitation" ADD "expiresAt" TIMESTAMP WITH TIME ZONE NOT NULL`)
    await queryRunner.query(`ALTER TABLE "organization_invitation" DROP COLUMN "createdAt"`)
    await queryRunner.query(
      `ALTER TABLE "organization_invitation" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
    await queryRunner.query(`ALTER TABLE "organization_invitation" DROP COLUMN "updatedAt"`)
    await queryRunner.query(
      `ALTER TABLE "organization_invitation" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "suspendedAt"`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "suspendedAt" TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "suspendedUntil"`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "suspendedUntil" TIMESTAMP WITH TIME ZONE`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "createdAt"`)
    await queryRunner.query(
      `ALTER TABLE "organization" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "updatedAt"`)
    await queryRunner.query(
      `ALTER TABLE "organization" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
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
    await queryRunner.query(`ALTER TABLE "organization_role" DROP COLUMN "createdAt"`)
    await queryRunner.query(
      `ALTER TABLE "organization_role" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
    await queryRunner.query(`ALTER TABLE "organization_role" DROP COLUMN "updatedAt"`)
    await queryRunner.query(
      `ALTER TABLE "organization_role" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
    await queryRunner.query(`ALTER TABLE "docker_registry" DROP COLUMN "createdAt"`)
    await queryRunner.query(
      `ALTER TABLE "docker_registry" ADD "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
    await queryRunner.query(`ALTER TABLE "docker_registry" DROP COLUMN "updatedAt"`)
    await queryRunner.query(
      `ALTER TABLE "docker_registry" ADD "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "docker_registry" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "docker_registry" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "docker_registry" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "docker_registry" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "organization_role" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "organization_role" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "organization_role" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "organization_role" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum_old" AS ENUM('write:registries', 'delete:registries', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'write:snapshots', 'delete:snapshots')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum_old"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum"`)
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum_old" RENAME TO "organization_role_permissions_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "suspendedUntil"`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "suspendedUntil" TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "suspendedAt"`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "suspendedAt" TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "organization_invitation" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "organization_invitation" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "organization_invitation" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "organization_invitation" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "organization_invitation" DROP COLUMN "expiresAt"`)
    await queryRunner.query(`ALTER TABLE "organization_invitation" ADD "expiresAt" TIMESTAMP NOT NULL`)
    await queryRunner.query(`ALTER TABLE "organization_user" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "organization_user" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "organization_user" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "organization_user" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "lastChecked"`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "lastChecked" TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "region"`)
    await queryRunner.query(`CREATE TYPE "public"."runner_region_enum" AS ENUM('eu', 'us', 'asia')`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "region" "public"."runner_region_enum" NOT NULL`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "build_info" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "build_info" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "build_info" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "build_info" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "build_info" DROP COLUMN "lastUsedAt"`)
    await queryRunner.query(`ALTER TABLE "build_info" ADD "lastUsedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "authToken" SET DEFAULT md5((random()))`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "lastActivityAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "lastActivityAt" TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "lastBackupAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "lastBackupAt" TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "region"`)
    await queryRunner.query(`CREATE TYPE "public"."sandbox_region_enum" AS ENUM('eu', 'us', 'asia')`)
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "region" "public"."sandbox_region_enum" NOT NULL DEFAULT 'eu'`)
    await queryRunner.query(`ALTER TABLE "snapshot_runner" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "snapshot_runner" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "snapshot_runner" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "snapshot_runner" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "volume" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "volume" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "volume" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "warm_pool" DROP COLUMN "updatedAt"`)
    await queryRunner.query(`ALTER TABLE "warm_pool" ADD "updatedAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "warm_pool" DROP COLUMN "createdAt"`)
    await queryRunner.query(`ALTER TABLE "warm_pool" ADD "createdAt" TIMESTAMP NOT NULL DEFAULT now()`)
    await queryRunner.query(`ALTER TABLE "warm_pool" DROP COLUMN "target"`)
    await queryRunner.query(`CREATE TYPE "public"."warm_pool_target_enum" AS ENUM('eu', 'us', 'asia')`)
    await queryRunner.query(
      `ALTER TABLE "warm_pool" ADD "target" "public"."warm_pool_target_enum" NOT NULL DEFAULT 'eu'`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" DROP COLUMN "endAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" ADD "endAt" TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" DROP COLUMN "startAt"`)
    await queryRunner.query(`ALTER TABLE "sandbox_usage_periods" ADD "startAt" TIMESTAMP NOT NULL`)
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum_old" AS ENUM('write:registries', 'delete:registries', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes', 'write:snapshots', 'delete:snapshots')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum_old"[] USING "permissions"::"text"::"public"."api_key_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum_old" RENAME TO "api_key_permissions_enum"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "version"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "availabilityScore"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentSnapshotCount"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentAllocatedDiskGiB"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentAllocatedMemoryGiB"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentAllocatedCpu"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentDiskUsagePercentage"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentMemoryUsagePercentage"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentCpuUsagePercentage"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "diskGiB"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "memoryGiB"`)
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "proxyUrl"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "networkAllowList"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "networkAllowAll"`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "disk" integer NOT NULL`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "memory" integer NOT NULL`)
    await queryRunner.query(`DROP INDEX "public"."audit_log_createdAt_index"`)
    await queryRunner.query(`DROP INDEX "public"."audit_log_organizationId_createdAt_index"`)
    await queryRunner.query(`DROP TABLE "audit_log"`)
  }
}
