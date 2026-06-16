import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1781597992117 implements MigrationInterface {
  name = 'Migration1781597992117'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "sandbox_usage_periods_archive" ADD "regionType" character varying NOT NULL DEFAULT 'shared'`,
    )
    await queryRunner.query(
      `ALTER TABLE "sandbox_usage_periods" ADD "regionType" character varying NOT NULL DEFAULT 'shared'`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "organization" ADD "preview_warning_enabled" boolean NOT NULL DEFAULT false`)
    await queryRunner.query(
      `CREATE INDEX "idx_sandbox_last_activity_at" ON "sandbox_last_activity" ("lastActivityAt") `,
    )
  }
}
