import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1760411389246 implements MigrationInterface {
  name = 'Migration1760411389246'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "public"."disk_state_enum" AS ENUM('fresh', 'pulling', 'ready', 'attached', 'detached', 'pending_push', 'pushing', 'stored', 'pending_delete', 'deleting', 'deleted', 'error', 'forking', 'locked')`,
    )
    await queryRunner.query(
      `CREATE TABLE "disk" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "organizationId" uuid NOT NULL, "name" character varying NOT NULL, "size" integer NOT NULL, "state" "public"."disk_state_enum" NOT NULL DEFAULT 'fresh', "baseDiskId" uuid, "runnerId" uuid, "sandboxId" uuid, "errorReason" character varying, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "disk_organizationId_name_unique" UNIQUE ("organizationId", "name"), CONSTRAINT "disk_id_pk" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ADD "disks" uuid array NOT NULL DEFAULT '{}'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE "disk"`)
    await queryRunner.query(`DROP TYPE "public"."disk_state_enum"`)
    await queryRunner.query(`ALTER TABLE "sandbox" DROP COLUMN "disks"`)
  }
}
