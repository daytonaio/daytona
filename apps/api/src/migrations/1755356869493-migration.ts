import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1755356869493 implements MigrationInterface {
  name = 'Migration1755356869493'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "webhook_initialization" ("organizationId" character varying NOT NULL, "endpointsCreated" boolean NOT NULL DEFAULT false, "svixApplicationCreated" boolean NOT NULL DEFAULT false, "endpointIds" jsonb, "svixApplicationId" character varying, "lastError" text, "retryCount" integer NOT NULL DEFAULT '0', "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "webhook_initialization_organizationId_pk" PRIMARY KEY ("organizationId"))`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE "webhook_initialization"`)
  }
}
