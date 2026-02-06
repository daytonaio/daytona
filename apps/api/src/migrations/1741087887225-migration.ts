/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MigrationInterface, QueryRunner } from 'typeorm'

export class Migration1741087887225 implements MigrationInterface {
  name = 'Migration1741087887225'

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`CREATE SCHEMA IF NOT EXISTS "public"`)
    await queryRunner.query(
      `CREATE TABLE "user" ("id" character varying NOT NULL, "name" character varying NOT NULL, "keyPair" text, "publicKeys" text NOT NULL, "total_cpu_quota" integer NOT NULL DEFAULT '10', "total_memory_quota" integer NOT NULL DEFAULT '40', "total_disk_quota" integer NOT NULL DEFAULT '100', "max_cpu_per_workspace" integer NOT NULL DEFAULT '2', "max_memory_per_workspace" integer NOT NULL DEFAULT '4', "max_disk_per_workspace" integer NOT NULL DEFAULT '10', "max_concurrent_workspaces" integer NOT NULL DEFAULT '10', "workspace_quota" integer NOT NULL DEFAULT '0', "image_quota" integer NOT NULL DEFAULT '5', "max_image_size" integer NOT NULL DEFAULT '2', "total_image_size" integer NOT NULL DEFAULT '5', CONSTRAINT "PK_cace4a159ff9f2512dd42373760" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `CREATE TABLE "team" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "name" character varying NOT NULL, CONSTRAINT "PK_f57d8293406df4af348402e4b74" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `CREATE TABLE "workspace_usage_periods" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "workspaceId" character varying NOT NULL, "startAt" TIMESTAMP NOT NULL, "endAt" TIMESTAMP, "cpu" double precision NOT NULL, "gpu" double precision NOT NULL, "mem" double precision NOT NULL, "disk" double precision NOT NULL, "storage" double precision NOT NULL, "region" character varying NOT NULL, CONSTRAINT "PK_b8d71f79ee638064397f678e877" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(`CREATE TYPE "node_class_enum" AS ENUM('small', 'medium', 'large')`)
    await queryRunner.query(`CREATE TYPE "node_region_enum" AS ENUM('eu', 'us', 'asia')`)
    await queryRunner.query(
      `CREATE TYPE "node_state_enum" AS ENUM('initializing', 'ready', 'disabled', 'decommissioned', 'unresponsive')`,
    )
    await queryRunner.query(
      `CREATE TABLE "node" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "domain" character varying NOT NULL, "apiUrl" character varying NOT NULL, "apiKey" character varying NOT NULL, "cpu" integer NOT NULL, "memory" integer NOT NULL, "disk" integer NOT NULL, "gpu" integer NOT NULL, "gpuType" character varying NOT NULL, "class" "node_class_enum" NOT NULL DEFAULT 'small', "used" integer NOT NULL DEFAULT '0', "capacity" integer NOT NULL, "region" "node_region_enum" NOT NULL, "state" "node_state_enum" NOT NULL DEFAULT 'initializing', "lastChecked" TIMESTAMP, "unschedulable" boolean NOT NULL DEFAULT false, "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "UQ_330d74ac3d0e349b4c73c62ad6d" UNIQUE ("domain"), CONSTRAINT "PK_8c8caf5f29d25264abe9eaf94dd" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `CREATE TYPE "image_node_state_enum" AS ENUM('pulling_image', 'ready', 'error', 'removing')`,
    )
    await queryRunner.query(
      `CREATE TABLE "image_node" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "state" "image_node_state_enum" NOT NULL DEFAULT 'pulling_image', "errorReason" character varying, "image" character varying NOT NULL, "internalImageName" character varying NOT NULL DEFAULT '', "nodeId" character varying NOT NULL, "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "PK_6c66fc8bd2b9fb41362a50fddd0" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `CREATE TYPE "image_state_enum" AS ENUM('pending', 'pulling_image', 'pending_validation', 'validating', 'active', 'error', 'removing')`,
    )
    await queryRunner.query(
      `CREATE TABLE "image" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "userId" character varying NOT NULL, "general" boolean NOT NULL DEFAULT false, "name" character varying NOT NULL, "internalName" character varying, "enabled" boolean NOT NULL DEFAULT true, "state" "image_state_enum" NOT NULL DEFAULT 'pending', "errorReason" character varying, "size" double precision, "entrypoint" character varying, "internalRegistryId" character varying, "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), "lastUsedAt" TIMESTAMP, CONSTRAINT "UQ_9db6fbe71409d80375c32826db3" UNIQUE ("userId", "name"), CONSTRAINT "PK_d6db1ab4ee9ad9dbe86c64e4cc3" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(`CREATE TYPE "workspace_region_enum" AS ENUM('eu', 'us', 'asia')`)
    await queryRunner.query(`CREATE TYPE "workspace_class_enum" AS ENUM('small', 'medium', 'large')`)
    await queryRunner.query(
      `CREATE TYPE "workspace_state_enum" AS ENUM('creating', 'restoring', 'destroyed', 'destroying', 'started', 'stopped', 'starting', 'stopping', 'resizing', 'error', 'unknown', 'pulling_image', 'archiving', 'archived')`,
    )
    await queryRunner.query(
      `CREATE TYPE "workspace_desiredstate_enum" AS ENUM('destroyed', 'started', 'stopped', 'resized', 'archived')`,
    )
    await queryRunner.query(
      `CREATE TYPE "workspace_snapshotstate_enum" AS ENUM('None', 'Pending', 'InProgress', 'Completed', 'Error')`,
    )
    await queryRunner.query(
      `CREATE TABLE "workspace" ("id" character varying NOT NULL, "name" character varying NOT NULL, "userId" character varying NOT NULL, "region" "workspace_region_enum" NOT NULL DEFAULT 'eu', "nodeId" uuid, "prevNodeId" uuid, "class" "workspace_class_enum" NOT NULL DEFAULT 'small', "state" "workspace_state_enum" NOT NULL DEFAULT 'unknown', "desiredState" "workspace_desiredstate_enum" NOT NULL DEFAULT 'started', "image" character varying NOT NULL, "osUser" character varying NOT NULL, "errorReason" character varying, "env" text NOT NULL DEFAULT '{}', "public" boolean NOT NULL DEFAULT false, "labels" jsonb, "snapshotRegistryId" character varying, "snapshotImage" character varying, "lastSnapshotAt" TIMESTAMP, "snapshotState" "workspace_snapshotstate_enum" NOT NULL DEFAULT 'None', "existingSnapshotImages" jsonb NOT NULL DEFAULT '[]', "cpu" integer NOT NULL DEFAULT '2', "gpu" integer NOT NULL DEFAULT '0', "mem" integer NOT NULL DEFAULT '4', "disk" integer NOT NULL DEFAULT '10', "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), "lastActivityAt" TIMESTAMP, "autoStopInterval" integer NOT NULL DEFAULT '15', CONSTRAINT "PK_ca86b6f9b3be5fe26d307d09b49" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `CREATE TABLE "docker_registry" ("id" uuid NOT NULL DEFAULT uuid_generate_v4(), "name" character varying NOT NULL, "url" character varying NOT NULL, "username" character varying NOT NULL, "password" character varying NOT NULL, "isDefault" boolean NOT NULL DEFAULT false, "project" character varying NOT NULL, "userId" character varying, "createdAt" TIMESTAMP NOT NULL DEFAULT now(), "updatedAt" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "PK_4ad72294240279415eb57799798" PRIMARY KEY ("id"))`,
    )
    await queryRunner.query(
      `CREATE TABLE "api_key" ("userId" character varying NOT NULL, "name" character varying NOT NULL, "value" character varying NOT NULL, "createdAt" TIMESTAMP NOT NULL, CONSTRAINT "UQ_4b0873b633484d5de20b2d8f852" UNIQUE ("value"), CONSTRAINT "PK_1df0337a701df00e9b2a16c8a0b" PRIMARY KEY ("userId", "name"))`,
    )
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE "api_key"`)
    await queryRunner.query(`DROP TABLE "docker_registry"`)
    await queryRunner.query(`DROP TABLE "workspace"`)
    await queryRunner.query(`DROP TYPE "workspace_snapshotstate_enum"`)
    await queryRunner.query(`DROP TYPE "workspace_desiredstate_enum"`)
    await queryRunner.query(`DROP TYPE "workspace_state_enum"`)
    await queryRunner.query(`DROP TYPE "workspace_class_enum"`)
    await queryRunner.query(`DROP TYPE "workspace_region_enum"`)
    await queryRunner.query(`DROP TABLE "image"`)
    await queryRunner.query(`DROP TYPE "image_state_enum"`)
    await queryRunner.query(`DROP TABLE "image_node"`)
    await queryRunner.query(`DROP TYPE "image_node_state_enum"`)
    await queryRunner.query(`DROP TABLE "node"`)
    await queryRunner.query(`DROP TYPE "node_state_enum"`)
    await queryRunner.query(`DROP TYPE "node_region_enum"`)
    await queryRunner.query(`DROP TYPE "node_class_enum"`)
    await queryRunner.query(`DROP TABLE "workspace_usage_periods"`)
    await queryRunner.query(`DROP TABLE "team"`)
    await queryRunner.query(`DROP TABLE "user"`)
  }
}
