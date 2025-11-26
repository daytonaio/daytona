import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1764162937691 implements MigrationInterface {
  name = 'Migration1764162937691'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "region" ADD "proxyUrl" character varying`)
    await queryRunner.query(`ALTER TABLE "region" ADD "toolboxProxyUrl" character varying`)
    await queryRunner.query(`ALTER TABLE "region" ADD "proxyApiKeyHash" character varying`)
    await queryRunner.query(`ALTER TABLE "region" ADD "sshGatewayUrl" character varying`)
    await queryRunner.query(`ALTER TABLE "region" ADD "sshGatewayApiKeyHash" character varying`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "proxyUrl" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "region" DROP DEFAULT`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "sshGatewayApiKeyHash"`)
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "sshGatewayUrl"`)
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "proxyApiKeyHash"`)
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "toolboxProxyUrl"`)
    await queryRunner.query(`ALTER TABLE "region" DROP COLUMN "proxyUrl"`)
  }
}
