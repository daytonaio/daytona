import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1753185133352 implements MigrationInterface {
  name = 'Migration1753185133352'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_target_propagation_state_enum" AS ENUM('ready', 'propagating')`,
    )
    await queryRunner.query(`
      CREATE TABLE "snapshot_target_propagation" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "target" character varying NOT NULL, "snapshotId" uuid NOT NULL, "desiredConcurrentSandboxes" integer NOT NULL, "userOverride" integer DEFAULT NULL, "state" "public"."snapshot_target_propagation_state_enum" NOT NULL DEFAULT 'propagating', "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "snapshot_target_propagation_id_pk" PRIMARY KEY ("id"))`)
    await queryRunner.renameColumn('snapshot', 'internalName', 'ref')
    await queryRunner.renameColumn('snapshot', 'buildRunnerId', 'initialRunnerId')
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "parentRefChain" text array`)

    // Update sandbox states
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" RENAME TO "sandbox_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_state_enum" AS ENUM('creating', 'restoring', 'destroyed', 'destroying', 'started', 'stopped', 'starting', 'stopping', 'error', 'build_failed', 'pending_build', 'building_snapshot', 'unknown', 'pending_pull', 'pulling_snapshot', 'archiving', 'archived')`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "state" TYPE "public"."sandbox_state_enum" USING "state"::"text"::"public"."sandbox_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" SET DEFAULT 'unknown'`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_state_enum_old"`)

    // Update snapshot states
    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME TO "snapshot_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_state_enum" AS ENUM('pending', 'pulling', 'pending_validation', 'validating', 'active', 'inactive', 'building', 'warming_up', 'error', 'build_failed', 'removing')`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ALTER COLUMN "state" TYPE "public"."snapshot_state_enum" USING "state"::"text"::"public"."snapshot_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" SET DEFAULT 'pending'`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_state_enum_old"`)

    await queryRunner.query(
      `ALTER TABLE "snapshot_runner" ADD CONSTRAINT "snapshot_runner_runnerId_snapshotRef_unique" UNIQUE ("runnerId", "snapshotRef")`,
    )
    await queryRunner.query(
      `ALTER TABLE "snapshot_target_propagation" ADD CONSTRAINT "snapshot_target_propagation_snapshotId_fk" FOREIGN KEY ("snapshotId") REFERENCES "snapshot"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "snapshot_target_propagation" DROP CONSTRAINT "snapshot_target_propagation_snapshotId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "snapshot_runner" DROP CONSTRAINT "snapshot_runner_runnerId_snapshotRef_unique"`,
    )

    // Revert snapshot states
    await queryRunner.query(`ALTER TYPE "public"."snapshot_state_enum" RENAME TO "snapshot_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_state_enum" AS ENUM('pending', 'pulling', 'pending_validation', 'validating', 'active', 'inactive', 'building', 'warming_up', 'error', 'build_failed', 'removing')`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ALTER COLUMN "state" TYPE "public"."snapshot_state_enum" USING "state"::"text"::"public"."snapshot_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "state" SET DEFAULT 'pending'`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_state_enum_old"`)

    // Revert sandbox states
    await queryRunner.query(`ALTER TYPE "public"."sandbox_state_enum" RENAME TO "sandbox_state_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."sandbox_state_enum" AS ENUM('creating', 'restoring', 'destroyed', 'destroying', 'started', 'stopped', 'starting', 'stopping', 'error', 'build_failed', 'pending_build', 'building_snapshot', 'unknown', 'pulling_snapshot', 'archiving', 'archived')`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "sandbox" ALTER COLUMN "state" TYPE "public"."sandbox_state_enum" USING "state"::"text"::"public"."sandbox_state_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "state" SET DEFAULT 'unknown'`)
    await queryRunner.query(`DROP TYPE "public"."sandbox_state_enum_old"`)

    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "parentRefChain"`)
    await queryRunner.renameColumn('snapshot', 'initialRunnerId', 'buildRunnerId')
    await queryRunner.renameColumn('snapshot', 'ref', 'internalName')

    await queryRunner.query(`DROP TABLE "snapshot_target_propagation"`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_target_propagation_state_enum"`)
  }
}
