import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1752067215297 implements MigrationInterface {
  name = 'Migration1752067215297'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD "version" character varying NOT NULL DEFAULT '0'`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "version"`)
  }
}
