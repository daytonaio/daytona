import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1759328077156 implements MigrationInterface {
  name = 'Migration1759328077156'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "snapshot" ADD "experimental" boolean NOT NULL DEFAULT false`)
    await queryRunner.query(`ALTER TABLE "runner" ADD "experimental" boolean NOT NULL DEFAULT false`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "runner" DROP COLUMN "experimental"`)
    await queryRunner.query(`ALTER TABLE "snapshot" DROP COLUMN "experimental"`)
  }
}
