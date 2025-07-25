/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  ObjectStorageApi,
  SnapshotDto,
  SnapshotsApi,
  SnapshotState,
  CreateSnapshot,
  Configuration,
} from '@daytonaio/api-client'
import { DaytonaError } from './errors/DaytonaError'
import { Image } from './Image'
import { Resources } from './Daytona'
import { processStreamingResponse } from './utils/Stream'
import { dynamicImport } from './utils/Import'

const SNAPSHOTS_FETCH_LIMIT = 200

/**
 * Represents a Daytona Snapshot which is a pre-configured sandbox.
 *
 * @property {string} id - Unique identifier for the Snapshot.
 * @property {string} organizationId - Organization ID that owns the Snapshot.
 * @property {boolean} general - Whether the Snapshot is general.
 * @property {string} name - Name of the Snapshot.
 * @property {string} imageName - Name of the Image of the Snapshot.
 * @property {boolean} enabled - Whether the Snapshot is enabled.
 * @property {SnapshotState} state - Current state of the Snapshot.
 * @property {number} size - Size of the Snapshot.
 * @property {string[]} entrypoint - Entrypoint of the Snapshot.
 * @property {number} cpu - CPU of the Snapshot.
 * @property {number} gpu - GPU of the Snapshot.
 * @property {number} mem - Memory of the Snapshot in GiB.
 * @property {number} disk - Disk of the Snapshot in GiB.
 * @property {string} errorReason - Error reason of the Snapshot.
 * @property {Date} createdAt - Timestamp when the Snapshot was created.
 * @property {Date} updatedAt - Timestamp when the Snapshot was last updated.
 * @property {Date} lastUsedAt - Timestamp when the Snapshot was last used.
 */
export type Snapshot = SnapshotDto & { __brand: 'Snapshot' }

/**
 * Parameters for creating a new snapshot.
 *
 * @property {string} name - Name of the snapshot.
 * @property {string | Image} image - Image of the snapshot. If a string is provided, it should be available on some registry.
 * If an Image instance is provided, it will be used to create a new image in Daytona.
 * @property {Resources} resources - Resources of the snapshot.
 * @property {string[]} entrypoint - Entrypoint of the snapshot.
 */
export type CreateSnapshotParams = {
  name: string
  image: string | Image
  resources?: Resources
  entrypoint?: string[]
}

/**
 * Service for managing Daytona Snapshots. Can be used to list, get, create and delete Snapshots.
 *
 * @class
 */
export class SnapshotService {
  constructor(
    private clientConfig: Configuration,
    private snapshotsApi: SnapshotsApi,
    private objectStorageApi: ObjectStorageApi,
  ) {}

  /**
   * List all Snapshots.
   *
   * @returns {Promise<Snapshot[]>} List of all Snapshots accessible to the user
   *
   * @example
   * const daytona = new Daytona();
   * const snapshots = await daytona.snapshot.list();
   * console.log(`Found ${snapshots.length} snapshots`);
   * snapshots.forEach(snapshot => console.log(`${snapshot.name} (${snapshot.imageName})`));
   */
  async list(): Promise<Snapshot[]> {
    let response = await this.snapshotsApi.getAllSnapshots(undefined, SNAPSHOTS_FETCH_LIMIT)
    if (response.data.total > SNAPSHOTS_FETCH_LIMIT) {
      response = await this.snapshotsApi.getAllSnapshots(undefined, response.data.total)
    }
    return response.data.items as Snapshot[]
  }

  /**
   * Gets a Snapshot by its name.
   *
   * @param {string} name - Name of the Snapshot to retrieve
   * @returns {Promise<Snapshot>} The requested Snapshot
   * @throws {Error} If the Snapshot does not exist or cannot be accessed
   *
   * @example
   * const daytona = new Daytona();
   * const snapshot = await daytona.snapshot.get("snapshot-name");
   * console.log(`Snapshot ${snapshot.name} is in state ${snapshot.state}`);
   */
  async get(name: string): Promise<Snapshot> {
    const response = await this.snapshotsApi.getSnapshot(name)
    return response.data as Snapshot
  }

  /**
   * Deletes a Snapshot.
   *
   * @param {Snapshot} snapshot - Snapshot to delete
   * @returns {Promise<void>}
   * @throws {Error} If the Snapshot does not exist or cannot be deleted
   *
   * @example
   * const daytona = new Daytona();
   * const snapshot = await daytona.snapshot.get("snapshot-name");
   * await daytona.snapshot.delete(snapshot);
   * console.log("Snapshot deleted successfully");
   */
  async delete(snapshot: Snapshot): Promise<void> {
    await this.snapshotsApi.removeSnapshot(snapshot.id)
  }

  /**
   * Creates and registers a new snapshot from the given Image definition.
   *
   * @param {CreateSnapshotParams} params - Parameters for snapshot creation.
   * @param {object} options - Options for the create operation.
   * @param {boolean} options.onLogs - This callback function handles snapshot creation logs.
   * @param {number} options.timeout - Default is no timeout. Timeout in seconds (0 means no timeout).
   * @returns {Promise<void>}
   *
   * @example
   * const image = Image.debianSlim('3.12').pipInstall('numpy');
   * await daytona.snapshot.create({ name: 'my-snapshot', image: image }, { onLogs: console.log });
   */
  public async create(
    params: CreateSnapshotParams,
    options: { onLogs?: (chunk: string) => void; timeout?: number } = {},
  ): Promise<Snapshot> {
    const createSnapshotReq: CreateSnapshot = {
      name: params.name,
    }

    if (typeof params.image === 'string') {
      createSnapshotReq.imageName = params.image
      createSnapshotReq.entrypoint = params.entrypoint
    } else {
      const contextHashes = await SnapshotService.processImageContext(this.objectStorageApi, params.image)
      createSnapshotReq.buildInfo = {
        contextHashes,
        dockerfileContent: params.entrypoint
          ? params.image.entrypoint(params.entrypoint).dockerfile
          : params.image.dockerfile,
      }
    }

    if (params.resources) {
      createSnapshotReq.cpu = params.resources.cpu
      createSnapshotReq.gpu = params.resources.gpu
      createSnapshotReq.memory = params.resources.memory
      createSnapshotReq.disk = params.resources.disk
    }

    let createdSnapshot = (
      await this.snapshotsApi.createSnapshot(createSnapshotReq, undefined, {
        timeout: (options.timeout || 0) * 1000,
      })
    ).data

    if (!createdSnapshot) {
      throw new DaytonaError("Failed to create snapshot. Didn't receive a snapshot from the server API.")
    }

    const terminalStates: SnapshotState[] = [SnapshotState.ACTIVE, SnapshotState.ERROR, SnapshotState.BUILD_FAILED]
    const logTerminalStates: SnapshotState[] = [
      ...terminalStates,
      SnapshotState.PENDING_VALIDATION,
      SnapshotState.VALIDATING,
    ]
    const snapshotRef = { createdSnapshot: createdSnapshot }
    let streamPromise: Promise<void> | undefined
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    const startLogStreaming = async (onChunk: (chunk: string) => void = () => {}) => {
      if (!streamPromise) {
        const url = `${this.clientConfig.basePath}/snapshots/${createdSnapshot.id}/build-logs?follow=true`

        streamPromise = processStreamingResponse(
          () => fetch(url, { method: 'GET', headers: this.clientConfig.baseOptions.headers }),
          (chunk) => onChunk(chunk.trimEnd()),
          async () => logTerminalStates.includes(snapshotRef.createdSnapshot.state),
        )
      }
    }

    if (options.onLogs) {
      options.onLogs(`Creating snapshot ${createdSnapshot.name} (${createdSnapshot.state})`)

      if (createdSnapshot.state !== SnapshotState.BUILD_PENDING) {
        await startLogStreaming(options.onLogs)
      }
    }

    let previousState = createdSnapshot.state
    while (!terminalStates.includes(createdSnapshot.state)) {
      if (options.onLogs && previousState !== createdSnapshot.state) {
        if (createdSnapshot.state !== SnapshotState.BUILD_PENDING && !streamPromise) {
          await startLogStreaming(options.onLogs)
        }
        options.onLogs(`Creating snapshot ${createdSnapshot.name} (${createdSnapshot.state})`)
        previousState = createdSnapshot.state
      }
      await new Promise((resolve) => setTimeout(resolve, 1000))
      createdSnapshot = await this.get(createdSnapshot.name)
      snapshotRef.createdSnapshot = createdSnapshot
    }

    if (options.onLogs) {
      if (streamPromise) {
        await streamPromise
      }
      if (createdSnapshot.state === SnapshotState.ACTIVE) {
        options.onLogs(`Created snapshot ${createdSnapshot.name} (${createdSnapshot.state})`)
      }
    }

    if (createdSnapshot.state === SnapshotState.ERROR || createdSnapshot.state === SnapshotState.BUILD_FAILED) {
      const errMsg = `Failed to create snapshot. Name: ${createdSnapshot.name} Reason: ${createdSnapshot.errorReason}`
      throw new DaytonaError(errMsg)
    }

    return createdSnapshot as Snapshot
  }

  /**
   * Activates a snapshot.
   *
   * @param {Snapshot} snapshot - Snapshot to activate
   * @returns {Promise<Snapshot>} The activated Snapshot instance
   */
  async activate(snapshot: Snapshot): Promise<Snapshot> {
    return (await this.snapshotsApi.activateSnapshot(snapshot.id)).data as Snapshot
  }

  /**
   * Processes the image contexts by uploading them to object storage
   *
   * @private
   * @param {Image} image - The Image instance.
   * @returns {Promise<string[]>} The list of context hashes stored in object storage.
   */
  static async processImageContext(objectStorageApi: ObjectStorageApi, image: Image): Promise<string[]> {
    if (!image.contextList || !image.contextList.length) {
      return []
    }

    const ObjectStorageModule = await dynamicImport('ObjectStorage', '"processImageContext" is not supported: ')
    const pushAccessCreds = (await objectStorageApi.getPushAccess()).data
    const objectStorage = new ObjectStorageModule.ObjectStorage({
      endpointUrl: pushAccessCreds.storageUrl,
      accessKeyId: pushAccessCreds.accessKey,
      secretAccessKey: pushAccessCreds.secret,
      sessionToken: pushAccessCreds.sessionToken,
      bucketName: pushAccessCreds.bucket,
    })

    const contextHashes = []
    for (const context of image.contextList) {
      const contextHash = await objectStorage.upload(
        context.sourcePath,
        pushAccessCreds.organizationId,
        context.archivePath,
      )
      contextHashes.push(contextHash)
    }

    return contextHashes
  }
}
