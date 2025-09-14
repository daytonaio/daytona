import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1757891323408 implements MigrationInterface {
  name = 'Migration1757891323408'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "experimental" boolean NOT NULL DEFAULT false`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "experimental"`)
  }
}
