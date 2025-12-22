import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1766422504882 implements MigrationInterface {
  name = 'Migration1766422504882'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD "currentCpuLoadAverage" double precision NOT NULL DEFAULT '0'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "currentCpuLoadAverage"`)
  }
}
