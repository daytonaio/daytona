import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1748017789450 implements MigrationInterface {
  name = 'Migration1748017789450'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "workspace" DROP CONSTRAINT "workspace_buildInfoImageRef_fk"`)
    await queryRunner.query(`ALTER TABLE "build_info" RENAME COLUMN "imageRef" TO "snapshotRef"`)
    await queryRunner.query(
      `ALTER TABLE "build_info" RENAME CONSTRAINT "build_info_imageRef_pk" TO "build_info_snapshotRef_pk"`,
    )
    await queryRunner.query(`ALTER TABLE "warm_pool" RENAME COLUMN "image" TO "snapshot"`)
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_runner_state_enum" AS ENUM('pulling_snapshot', 'building_snapshot', 'ready', 'error', 'removing')`,
    )
    await queryRunner.query(
      `CREATE TABLE "snapshot_runner" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "state" "public"."snapshot_runner_state_enum" NOT NULL DEFAULT 'pulling_snapshot', "errorReason" character varying, "snapshotRef" character varying NOT NULL DEFAULT '', "runnerId" character varying NOT NULL, "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "snapshot_runner_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_state_enum" AS ENUM('build_pending', 'building', 'pending', 'pulling_image', 'pending_validation', 'validating', 'active', 'error', 'removing')`,
    )
    await queryRunner.query(
      `CREATE TABLE "snapshot" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "organizationId" uuid, "general" boolean NOT NULL DEFAULT false, "name" character varying NOT NULL, "internalName" character varying, "enabled" boolean NOT NULL DEFAULT true, "state" "public"."snapshot_state_enum" NOT NULL DEFAULT 'pending', "errorReason" character varying, "size" double precision, "entrypoint" text array, "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), "lastUsedAt" TIMESTAMP, "buildRunnerId" character varying, "buildInfoSnapshotRef" character varying, CONSTRAINT "snapshot_organizationId_name_unique" UNIQUE ("organizationId", "name"), CONSTRAINT "snapshot_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `CREATE TABLE "runner" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "domain" character varying NOT NULL, "apiUrl" character varying NOT NULL, "apiKey" character varying NOT NULL, "cpu" integer NOT NULL, "memory" integer NOT NULL, "disk" integer NOT NULL, "gpu" integer NOT NULL, "gpuType" character varying NOT NULL, "class" "public"."runner_class_enum" NOT NULL DEFAULT 'small', "used" integer NOT NULL DEFAULT '0', "capacity" integer NOT NULL, "region" "public"."runner_region_enum" NOT NULL, "state" "public"."runner_state_enum" NOT NULL DEFAULT 'initializing', "lastChecked" TIMESTAMP, "unschedulable" boolean NOT NULL DEFAULT false, "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "runner_domain_unique" UNIQUE ("domain"), CONSTRAINT "runner_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "nodeId"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "prevNodeId"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "image"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "buildInfoImageRef"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "snapshotRegistryId"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "snapshotImage"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "lastSnapshotAt"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "snapshotState"`)
    await queryRunner.query(`DROP TYPE "public"."workspace_backupstate_enum"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "existingSnapshotImages"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "image_quota"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "max_image_size"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_image_size"`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "runnerId" uuid`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "prevRunnerId" uuid`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "snapshot" character varying`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "backupRegistryId" character varying`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "backupSnapshot" character varying`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "lastBackupAt" TIMESTAMP`)
    await queryRunner.query(
      `ALTER TABLE "workspace" ADD "backupState" "public"."workspace_backupstate_enum" NOT NULL DEFAULT 'None'`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ADD "existingBackupSnapshots" jsonb NOT NULL DEFAULT '[]'`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "buildInfoSnapshotRef" character varying`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "snapshot_quota" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "max_snapshot_size" integer NOT NULL DEFAULT '2'`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "total_snapshot_size" integer NOT NULL DEFAULT '5'`)
    await queryRunner.query(`ALTER TYPE "public"."workspace_state_enum" RENAME TO "workspace_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."workspace_state_enum" AS ENUM('creating', 'restoring', 'destroyed', 'destroying', 'started', 'stopped', 'starting', 'stopping', 'error', 'pending_build', 'building_snapshot', 'unknown', 'pulling_snapshot', 'archiving', 'archived')`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "workspace" ALTER COLUMN "state" TYPE "public"."workspace_state_enum" USING "state"::"text"::"public"."workspace_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "state" SET DEFAULT 'unknown'`)
    await queryRunner.query(`DROP TYPE "public"."workspace_state_enum_old"`)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "authToken" SET DEFAULT MD5(random()::text)`)
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum" RENAME TO "api_key_permissions_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum"[] USING "permissions"::"text"::"public"."api_key_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum_old"`)
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum" RENAME TO "organization_role_permissions_enum_old"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum" AS ENUM('write:registries', 'delete:registries', 'write:snapshots', 'delete:snapshots', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum_old"`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ADD CONSTRAINT "snapshot_buildInfoSnapshotRef_fk" FOREIGN KEY ("buildInfoSnapshotRef") REFERENCES "build_info"("snapshotRef") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
    await queryRunner.query(
      `ALTER TABLE "workspace" ADD CONSTRAINT "workspace_buildInfoSnapshotRef_fk" FOREIGN KEY ("buildInfoSnapshotRef") REFERENCES "build_info"("snapshotRef") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "workspace" DROP CONSTRAINT "workspace_buildInfoSnapshotRef_fk"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP CONSTRAINT "snapshot_buildInfoSnapshotRef_fk"`)
    await queryRunner.query(
      `CREATE TYPE "public"."organization_role_permissions_enum_old" AS ENUM('write:registries', 'delete:registries', 'write:images', 'delete:images', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role" ALTER COLUMN "permissions" TYPE "public"."organization_role_permissions_enum_old"[] USING "permissions"::"text"::"public"."organization_role_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."organization_role_permissions_enum"`)
    await queryRunner.query(
      `ALTER TYPE "public"."organization_role_permissions_enum_old" RENAME TO "organization_role_permissions_enum"`,
    )
    await queryRunner.query(
      `CREATE TYPE "public"."api_key_permissions_enum_old" AS ENUM('write:registries', 'delete:registries', 'write:images', 'delete:images', 'write:sandboxes', 'delete:sandboxes', 'read:volumes', 'write:volumes', 'delete:volumes')`,
    )
    await queryRunner.query(
      `ALTER TABLE "api_key" ALTER COLUMN "permissions" TYPE "public"."api_key_permissions_enum_old"[] USING "permissions"::"text"::"public"."api_key_permissions_enum_old"[]`,
    )
    await queryRunner.query(`DROP TYPE "public"."api_key_permissions_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."api_key_permissions_enum_old" RENAME TO "api_key_permissions_enum"`)
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "authToken" SET DEFAULT md5((random()))`)
    await queryRunner.query(
      `CREATE TYPE "public"."workspace_state_enum_old" AS ENUM('creating', 'restoring', 'destroyed', 'destroying', 'started', 'stopped', 'starting', 'stopping', 'error', 'pending_build', 'building_image', 'unknown', 'pulling_image', 'archiving', 'archived')`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "workspace" ALTER COLUMN "state" TYPE "public"."workspace_state_enum_old" USING "state"::"text"::"public"."workspace_state_enum_old"`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ALTER COLUMN "state" SET DEFAULT 'unknown'`)
    await queryRunner.query(`DROP TYPE "public"."workspace_state_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."workspace_state_enum_old" RENAME TO "workspace_state_enum"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "total_snapshot_size"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "max_snapshot_size"`)
    await queryRunner.query(`ALTER TABLE "organization" DROP COLUMN "snapshot_quota"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "buildInfoSnapshotRef"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "existingBackupSnapshots"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "backupState"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "lastBackupAt"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "backupSnapshot"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "backupRegistryId"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "snapshot"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "prevRunnerId"`)
    await queryRunner.query(`ALTER TABLE "workspace" DROP COLUMN "runnerId"`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "total_image_size" integer NOT NULL DEFAULT '5'`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "max_image_size" integer NOT NULL DEFAULT '2'`)
    await queryRunner.query(`ALTER TABLE "organization" ADD "image_quota" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "existingSnapshotImages" jsonb NOT NULL DEFAULT '[]'`)
    await queryRunner.query(
      `CREATE TYPE "public"."workspace_backupstate_enum" AS ENUM('None', 'Pending', 'InProgress', 'Completed', 'Error')`,
    )
    await queryRunner.query(
      `ALTER TABLE "workspace" ADD "snapshotState" "public"."workspace_backupstate_enum" NOT NULL DEFAULT 'None'`,
    )
    await queryRunner.query(`ALTER TABLE "workspace" ADD "lastSnapshotAt" TIMESTAMP`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "snapshotImage" character varying`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "snapshotRegistryId" character varying`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "buildInfoImageRef" character varying`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "image" character varying`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "prevNodeId" uuid`)
    await queryRunner.query(`ALTER TABLE "workspace" ADD "nodeId" uuid`)
    await queryRunner.query(`DROP TABLE "runner"`)
    await queryRunner.query(`DROP TABLE "snapshot"`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_state_enum"`)
    await queryRunner.query(`DROP TABLE "snapshot_runner"`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_runner_state_enum"`)
    await queryRunner.query(`ALTER TABLE "warm_pool" RENAME COLUMN "snapshot" TO "image"`)
    await queryRunner.query(
      `ALTER TABLE "build_info" RENAME CONSTRAINT "build_info_snapshotRef_pk" TO "build_info_imageRef_pk"`,
    )
    await queryRunner.query(`ALTER TABLE "build_info" RENAME COLUMN "snapshotRef" TO "imageRef"`)
    await queryRunner.query(
      `ALTER TABLE "workspace" ADD CONSTRAINT "workspace_buildInfoImageRef_fk" FOREIGN KEY ("buildInfoImageRef") REFERENCES "build_info"("imageRef") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
  }
}
