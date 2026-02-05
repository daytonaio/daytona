import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1770211874119 implements MigrationInterface {
  name = 'Migration1770211874119'

  public async up(queryRunner: QueryRunner): Promise<void> {
    // Update snapshot_runnerclass_enum
    await queryRunner.query(`ALTER TYPE "public"."snapshot_runnerclass_enum" RENAME TO "snapshot_runnerclass_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_runnerclass_enum" AS ENUM('linux', 'linux-exp', 'windows-exp', 'android-exp')`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "runnerClass" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ALTER COLUMN "runnerClass" TYPE "public"."snapshot_runnerclass_enum" USING "runnerClass"::"text"::"public"."snapshot_runnerclass_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "runnerClass" SET DEFAULT 'linux'`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_runnerclass_enum_old"`)

    // Update runner_class_enum
    await queryRunner.query(`ALTER TYPE "public"."runner_class_enum" RENAME TO "runner_class_enum_old"`)
    await queryRunner.query(
      `CREATE TYPE "public"."runner_class_enum" AS ENUM('linux', 'linux-exp', 'windows-exp', 'android-exp')`,
    )
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "class" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "runner" ALTER COLUMN "class" TYPE "public"."runner_class_enum" USING "class"::"text"::"public"."runner_class_enum"`,
    )
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "class" SET DEFAULT 'linux'`)
    await queryRunner.query(`DROP TYPE "public"."runner_class_enum_old"`)
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    // Revert runner_class_enum
    await queryRunner.query(`CREATE TYPE "public"."runner_class_enum_old" AS ENUM('linux', 'linux-exp', 'windows-exp')`)
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "class" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "runner" ALTER COLUMN "class" TYPE "public"."runner_class_enum_old" USING "class"::"text"::"public"."runner_class_enum_old"`,
    )
    await queryRunner.query(`ALTER TABLE "runner" ALTER COLUMN "class" SET DEFAULT 'linux'`)
    await queryRunner.query(`DROP TYPE "public"."runner_class_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."runner_class_enum_old" RENAME TO "runner_class_enum"`)

    // Revert snapshot_runnerclass_enum
    await queryRunner.query(
      `CREATE TYPE "public"."snapshot_runnerclass_enum_old" AS ENUM('linux', 'linux-exp', 'windows-exp')`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "runnerClass" DROP DEFAULT`)
    await queryRunner.query(
      `ALTER TABLE "snapshot" ALTER COLUMN "runnerClass" TYPE "public"."snapshot_runnerclass_enum_old" USING "runnerClass"::"text"::"public"."snapshot_runnerclass_enum_old"`,
    )
    await queryRunner.query(`ALTER TABLE "snapshot" ALTER COLUMN "runnerClass" SET DEFAULT 'linux'`)
    await queryRunner.query(`DROP TYPE "public"."snapshot_runnerclass_enum"`)
    await queryRunner.query(`ALTER TYPE "public"."snapshot_runnerclass_enum_old" RENAME TO "snapshot_runnerclass_enum"`)
  }
}
