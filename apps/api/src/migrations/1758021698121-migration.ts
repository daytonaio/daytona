import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1758021698121 implements MigrationInterface {
  name = 'Migration1758021698121'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" ADD "experimental" boolean NOT NULL DEFAULT false`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "experimental"`)
  }
}
