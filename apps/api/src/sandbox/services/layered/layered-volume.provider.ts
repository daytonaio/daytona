/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const LAYERED_VOLUME_PROVIDER = Symbol('LayeredVolumeProvider')

export type DiskMount =
  | {
      type: 's3'
      bucketName: string
      accessKeyId: string
      secretAccessKey: string
      sessionToken?: string
      bucketPrefix?: string
    }
  | {
      type: 's3-compatible'
      bucketName: string
      bucketEndpoint: string
      accessKeyId: string
      secretAccessKey: string
      bucketPrefix?: string
    }

export interface CreateDiskOptions {
  name: string
  region?: string
  mount: DiskMount
}

export interface CreateDiskResult {
  diskId: string
  region: string
  mountToken: string
}

export interface MintMountKeyOptions {
  diskId: string
  region: string
  nickname: string
}

export interface MintMountKeyResult {
  token: string
  identifier: string
}

export interface LayeredVolumeProvider {
  isConfigured(): boolean
  getDefaultRegion(): string
  createDisk(opts: CreateDiskOptions): Promise<CreateDiskResult>
  deleteDisk(diskId: string, region: string): Promise<void>
  mintMountKey(opts: MintMountKeyOptions): Promise<MintMountKeyResult>
  revokeMountKey(diskId: string, region: string, identifier: string): Promise<void>
}
