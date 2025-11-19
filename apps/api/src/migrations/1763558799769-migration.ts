import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1763558799769 implements MigrationInterface {
  name = 'Migration1763558799769'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD "name" character varying`)
    await queryRunner.query(`UPDATE "runner" SET "name" = "id"`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "name" SET NOT NULL`)
    await queryRunner.query(
      `ALTER TABLE "runner" ADD CONSTRAINT "runner_regionId_name_unique" UNIQUE ("regionId", "name")`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "name"`)
  }
}
