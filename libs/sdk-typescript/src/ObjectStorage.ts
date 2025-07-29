/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ListObjectsV2Command, S3Client } from '@aws-sdk/client-s3'
import { Upload } from '@aws-sdk/lib-storage'
import * as crypto from 'crypto'
import * as pathe from 'pathe'
import { DaytonaError } from './errors/DaytonaError'
import { dynamicImport } from './utils/Import'

/**
 * Configuration for the ObjectStorage class.
 *
 * @interface
 * @property {string} endpointUrl - The endpoint URL for the object storage service.
 * @property {string} accessKeyId - The access key ID for the object storage service.
 * @property {string} secretAccessKey - The secret access key for the object storage service.
 * @property {string} [sessionToken] - The session token for the object storage service. Used for temporary credentials.
 * @property {string} [bucketName] - The name of the bucket to use.
 */
export interface ObjectStorageConfig {
  endpointUrl: string
  accessKeyId: string
  secretAccessKey: string
  sessionToken?: string
  bucketName?: string
}

/**
 * ObjectStorage class for interacting with object storage services.
 *
 * @class
 * @param {ObjectStorageConfig} config - The configuration for the object storage service.
 */
export class ObjectStorage {
  private bucketName: string
  private s3Client: S3Client

  constructor(config: ObjectStorageConfig) {
    this.bucketName = config.bucketName || 'daytona-volume-builds'
    this.s3Client = new S3Client({
      region: this.extractAwsRegion(config.endpointUrl) || 'us-east-1',
      endpoint: config.endpointUrl,
      credentials: {
        accessKeyId: config.accessKeyId,
        secretAccessKey: config.secretAccessKey,
        sessionToken: config.sessionToken,
      },
      forcePathStyle: true,
    })
  }

  /**
   * Upload a file or directory to object storage.
   *
   * @param {string} path - The path to the file or directory to upload.
   * @param {string} organizationId - The organization ID to use for the upload.
   * @param {string} archiveBasePath - The base path to use for the archive.
   * @returns {Promise<string>} The hash of the uploaded file or directory.
   */
  async upload(path: string, organizationId: string, archiveBasePath: string): Promise<string> {
    const fs = await dynamicImport('fs', '"upload" is not supported: ')

    if (!fs.existsSync(path)) {
      const errMsg = `Path does not exist: ${path}`
      throw new DaytonaError(errMsg)
    }

    // Compute hash for the path
    const pathHash = await this.computeHashForPathMd5(path, archiveBasePath)

    // Define the S3 prefix
    const prefix = `${organizationId}/${pathHash}/`
    const s3Key = `${prefix}context.tar`

    // Check if it already exists in S3
    if (await this.folderExistsInS3(prefix)) {
      return pathHash
    }

    // Upload to S3
    await this.uploadAsTar(s3Key, path, archiveBasePath)

    return pathHash
  }

  /**
   * Compute a hash for a file or directory.
   *
   * @param {string} pathStr - The path to the file or directory to hash.
   * @param {string} archiveBasePath - The base path to use for the archive.
   * @returns {Promise<string>} The hash of the file or directory.
   */
  private async computeHashForPathMd5(pathStr: string, archiveBasePath: string): Promise<string> {
    const fs = await dynamicImport('fs', '"computeHashForPathMd5" is not supported: ')

    const md5Hasher = crypto.createHash('md5')
    const absPathStr = pathe.resolve(pathStr)

    md5Hasher.update(archiveBasePath)

    if (fs.statSync(absPathStr).isFile()) {
      // For files, hash the content
      await this.hashFile(absPathStr, md5Hasher)
    } else {
      // For directories, recursively hash all files and their paths
      await this.hashDirectory(absPathStr, pathStr, md5Hasher)
    }

    return md5Hasher.digest('hex')
  }

  /**
   * Recursively hash a directory and its contents.
   *
   * @param {string} dirPath - The path to the directory to hash.
   * @param {string} basePath - The base path to use for the hash.
   * @param {crypto.Hash} hasher - The hasher to use for the hash.
   * @returns {Promise<void>} A promise that resolves when the directory has been hashed.
   */
  private async hashDirectory(dirPath: string, basePath: string, hasher: crypto.Hash): Promise<void> {
    const fs = await dynamicImport('fs', '"hashDirectory" is not supported: ')

    const entries = fs.readdirSync(dirPath, { withFileTypes: true })
    const hasSubdirs = entries.some((e) => e.isDirectory())
    const hasFiles = entries.some((e) => e.isFile())

    if (!hasSubdirs && !hasFiles) {
      // Empty directory
      const relDir = pathe.relative(basePath, dirPath)
      hasher.update(relDir)
    }

    for (const entry of entries) {
      const fullPath = pathe.join(dirPath, entry.name)

      if (entry.isDirectory()) {
        await this.hashDirectory(fullPath, basePath, hasher)
      } else if (entry.isFile()) {
        const relPath = pathe.relative(basePath, fullPath)
        hasher.update(relPath)

        await this.hashFile(fullPath, hasher)
      }
    }
  }

  /**
   * Hash a file.
   *
   * @param {string} filePath - The path to the file to hash.
   * @param {crypto.Hash} hasher - The hasher to use for the hash.
   * @returns {Promise<void>} A promise that resolves when the file has been hashed.
   */
  private async hashFile(filePath: string, hasher: crypto.Hash): Promise<void> {
    const fs = await dynamicImport('fs', '"hashFile" is not supported: ')

    await new Promise<void>((resolve, reject) => {
      const stream = fs.createReadStream(filePath, { highWaterMark: 8192 })
      stream.on('data', (chunk) => hasher.update(chunk))
      stream.on('end', resolve)
      stream.on('error', reject)
    })
  }

  /**
   * Check if a prefix (folder) exists in S3.
   *
   * @param {string} prefix - The prefix to check.
   * @returns {Promise<boolean>} True if the prefix exists, false otherwise.
   */
  private async folderExistsInS3(prefix: string): Promise<boolean> {
    const response = await this.s3Client.send(
      new ListObjectsV2Command({
        Bucket: this.bucketName,
        Prefix: prefix,
        MaxKeys: 1,
      }),
    )

    return !!response.Contents && response.Contents.length > 0
  }

  /**
   * Create a tar archive of the specified path and upload it to S3.
   *
   * @param {string} s3Key - The key to use for the uploaded file.
   * @param {string} sourcePath - The path to the file or directory to upload.
   * @param {string} archiveBasePath - The base path to use for the archive.
   */
  private async uploadAsTar(s3Key: string, sourcePath: string, archiveBasePath: string) {
    const importErrorPrefix = '"uploadAsTar" is not supported: '
    const tar = await dynamicImport('tar', importErrorPrefix)
    const stream = await dynamicImport('stream', importErrorPrefix)

    sourcePath = pathe.resolve(sourcePath)
    const normalizedSourcePath = pathe.normalize(sourcePath)
    const normalizedArchiveBasePath = pathe.normalize(archiveBasePath)

    let basePrefix: string

    if (normalizedArchiveBasePath === '.') {
      // When archiveBasePath is empty (normalized to '.'), use the normalizedSourcePath as cwd and the '.' as target
      basePrefix = normalizedSourcePath
    } else {
      // Normal case: extract the base prefix by removing archiveBasePath from the end
      basePrefix = normalizedSourcePath.slice(0, normalizedSourcePath.length - normalizedArchiveBasePath.length)
    }

    const tarStream = tar.create(
      {
        cwd: basePrefix,
        portable: true,
        gzip: false,
      },
      [normalizedArchiveBasePath],
    )

    const pass = new stream.PassThrough()
    tarStream.pipe(pass)

    const uploader = new Upload({
      client: this.s3Client,
      params: {
        Bucket: this.bucketName,
        Key: s3Key,
        Body: pass,
      },
    })

    await uploader.done()
  }

  private extractAwsRegion(endpoint: string): string | undefined {
    const match = endpoint.match(/s3[.-]([a-z0-9-]+)\.amazonaws\.com/)
    return match?.[1]
  }
}
