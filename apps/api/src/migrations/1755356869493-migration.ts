import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1755356869493 implements MigrationInterface {
  name = 'Migration1755356869493'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" DROP CONSTRAINT "organization_role_assignment_invitation_roleId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" DROP CONSTRAINT "organization_role_assignment_roleId_fk"`,
    )
    await queryRunner.query(
      `CREATE TABLE "webhook_initialization" ("organizationId" character varying NOT NULL, "endpointsCreated" boolean NOT NULL DEFAULT false, "svixApplicationCreated" boolean NOT NULL DEFAULT false, "endpointIds" jsonb, "svixApplicationId" character varying, "lastError" text, "retryCount" integer NOT NULL DEFAULT '0', "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "webhook_initialization_organizationId_pk" PRIMARY KEY ("organizationId"))`,
    )
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "authToken" SET DEFAULT MD5(random()::text)`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "proxyUrl" DROP DEFAULT`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "region" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" ADD CONSTRAINT "organization_role_assignment_invitation_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" ADD CONSTRAINT "organization_role_assignment_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" DROP CONSTRAINT "organization_role_assignment_roleId_fk"`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" DROP CONSTRAINT "organization_role_assignment_invitation_roleId_fk"`,
    )
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "region" SET DEFAULT 'us'`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "proxyUrl" SET DEFAULT ''`)
    await queryRunner.query(`ALTER TABLE "sandbox" ALTER COLUMN "authToken" SET DEFAULT md5((random()))`)
    await queryRunner.query(`DROP TABLE "webhook_initialization"`)
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment" ADD CONSTRAINT "organization_role_assignment_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
    )
    await queryRunner.query(
      `ALTER TABLE "organization_role_assignment_invitation" ADD CONSTRAINT "organization_role_assignment_invitation_roleId_fk" FOREIGN KEY ("roleId") REFERENCES "organization_role"("id") ON DELETE CASCADE ON UPDATE CASCADE`,
    )
  }
}
