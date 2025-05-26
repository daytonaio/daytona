import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1748337927898 implements MigrationInterface {
  name = 'Migration1748337927898'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "imageName" character varying NOT NULL DEFAULT ''`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "cpu" integer NOT NULL DEFAULT '1'`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "gpu" integer NOT NULL DEFAULT '0'`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "mem" integer NOT NULL DEFAULT '1'`)
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "disk" integer NOT NULL DEFAULT '3'`)

    // Update existing rows to set image_name to the value of name
    await queryRunner.query(`UPDATE "snapshot" SET "imageName" = "name"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "disk"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "mem"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "gpu"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "cpu"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "imageName"`)
  }
}
