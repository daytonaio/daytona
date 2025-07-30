/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  Configuration,
  SnapshotsApi,
  ObjectStorageApi,
  SandboxApi,
  SandboxState,
  ToolboxApi,
  VolumesApi,
  SandboxVolume,
} from '@daytonaio/api-client'
import axios, { AxiosError } from 'axios'
import { SandboxPythonCodeToolbox } from './code-toolbox/SandboxPythonCodeToolbox'
import { SandboxTsCodeToolbox } from './code-toolbox/SandboxTsCodeToolbox'
import { DaytonaError, DaytonaNotFoundError } from './errors/DaytonaError'
import { Image } from './Image'
import { Sandbox } from './Sandbox'
import { SnapshotService } from './Snapshot'
import { VolumeService } from './Volume'
import * as packageJson from '../package.json'
import { processStreamingResponse } from './utils/Stream'
import { getEnvVar, RUNTIME, Runtime } from './utils/Runtime'

/**
 * Represents a volume mount for a Sandbox.
 *
 * @interface
 * @property {string} volumeId - ID of the Volume to mount
 * @property {string} mountPath - Path on the Sandbox to mount the Volume
 */

export interface VolumeMount extends SandboxVolume {
  volumeId: string
  mountPath: string
}

/**
 * Configuration options for initializing the Daytona client.
 *
 * @interface
 * @property {string} apiKey - API key for authentication with the Daytona API
 * @property {string} jwtToken - JWT token for authentication with the Daytona API. If not set, it must be provided
 * via the environment variable `DAYTONA_JWT_TOKEN`, or an API key must be provided instead.
 * @property {string} organizationId - Organization ID used for JWT-based authentication. Required if a JWT token
 * is provided, and must be set either here or in the environment variable `DAYTONA_ORGANIZATION_ID`.
 * @property {string} apiUrl - URL of the Daytona API. Defaults to 'https://app.daytona.io/api'
 * if not set here and not set in environment variable DAYTONA_API_URL.
 * @property {string} target - Target location for Sandboxes
 *
 * @example
 * const config: DaytonaConfig = {
 *     apiKey: "your-api-key",
 *     apiUrl: "https://your-api.com",
 *     target: "us"
 * };
 * const daytona = new Daytona(config);
 */
export interface DaytonaConfig {
  /** API key for authentication with the Daytona API */
  apiKey?: string
  /** JWT token for authentication with the Daytona API */
  jwtToken?: string
  /** Organization ID for authentication with the Daytona API */
  organizationId?: string
  /** URL of the Daytona API.
   */
  apiUrl?: string
  /**
   * @deprecated Use `apiUrl` instead. This property will be removed in future versions.
   */
  serverUrl?: string
  /** Target environment for sandboxes */
  target?: string
}

/**
 * Supported programming languages for code execution
 */
export enum CodeLanguage {
  PYTHON = 'python',
  TYPESCRIPT = 'typescript',
  JAVASCRIPT = 'javascript',
}

/**
 * Resource allocation for a Sandbox.
 *
 * @interface
 * @property {number} [cpu] - CPU allocation for the Sandbox in cores
 * @property {number} [gpu] - GPU allocation for the Sandbox in units
 * @property {number} [memory] - Memory allocation for the Sandbox in GiB
 * @property {number} [disk] - Disk space allocation for the Sandbox in GiB
 *
 * @example
 * const resources: SandboxResources = {
 *     cpu: 2,
 *     memory: 4,  // 4GiB RAM
 *     disk: 20    // 20GiB disk
 * };
 */
export interface Resources {
  /** CPU allocation for the Sandbox */
  cpu?: number
  /** GPU allocation for the Sandbox */
  gpu?: number
  /** Memory allocation for the Sandbox in GiB */
  memory?: number
  /** Disk space allocation for the Sandbox in GiB */
  disk?: number
}

/**
 * Base parameters for creating a new Sandbox.
 *
 * @interface
 * @property {string} [user] - Optional os user to use for the Sandbox
 * @property {CodeLanguage | string} [language] - Programming language for direct code execution
 * @property {Record<string, string>} [envVars] - Optional environment variables to set in the Sandbox
 * @property {Record<string, string>} [labels] - Sandbox labels
 * @property {boolean} [public] - Is the Sandbox port preview public
 * @property {number} [autoStopInterval] - Auto-stop interval in minutes (0 means disabled). Default is 15 minutes.
 * @property {number} [autoArchiveInterval] - Auto-archive interval in minutes (0 means the maximum interval will be used). Default is 7 days.
 * @property {number} [autoDeleteInterval] - Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping). By default, auto-delete is disabled.
 * @property {VolumeMount[]} [volumes] - Optional array of volumes to mount to the Sandbox
 */
export type CreateSandboxBaseParams = {
  user?: string
  language?: CodeLanguage | string
  envVars?: Record<string, string>
  labels?: Record<string, string>
  public?: boolean
  autoStopInterval?: number
  autoArchiveInterval?: number
  autoDeleteInterval?: number
  volumes?: VolumeMount[]
}

/**
 * Parameters for creating a new Sandbox.
 *
 * @interface
 * @property {string | Image} [image] - Custom Docker image to use for the Sandbox. If an Image object is provided,
 * the image will be dynamically built.
 * @property {Resources} [resources] - Resource allocation for the Sandbox. If not provided, sandbox will
 * have default resources.
 */
export type CreateSandboxFromImageParams = CreateSandboxBaseParams & {
  image: string | Image
  resources?: Resources
}

/**
 * Parameters for creating a new Sandbox from a snapshot.
 *
 * @interface
 * @property {string} [snapshot] - Name of the snapshot to use for the Sandbox.
 */
export type CreateSandboxFromSnapshotParams = CreateSandboxBaseParams & {
  snapshot?: string
}

/**
 * Filter for Sandboxes.
 *
 * @interface
 * @property {string} [id] - The ID of the Sandbox to retrieve
 * @property {Record<string, string>} [labels] - Labels to filter Sandboxes
 */
export type SandboxFilter = {
  id?: string
  labels?: Record<string, string>
}

/**
 * Main class for interacting with the Daytona API.
 * Provides methods for creating, managing, and interacting with Daytona Sandboxes.
 * Can be initialized either with explicit configuration or using environment variables.
 *
 * @property {VolumeService} volume - Service for managing Daytona Volumes
 * @property {SnapshotService} snapshot - Service for managing Daytona Snapshots
 *
 * @example
 * // Using environment variables
 * // Uses DAYTONA_API_KEY, DAYTONA_API_URL, DAYTONA_TARGET
 * const daytona = new Daytona();
 * const sandbox = await daytona.create();
 *
 * @example
 * // Using explicit configuration
 * const config: DaytonaConfig = {
 *     apiKey: "your-api-key",
 *     apiUrl: "https://your-api.com",
 *     target: "us"
 * };
 * const daytona = new Daytona(config);
 *
 * @class
 */
export class Daytona {
  private readonly clientConfig: Configuration
  private readonly sandboxApi: SandboxApi
  private readonly toolboxApi: ToolboxApi
  private readonly objectStorageApi: ObjectStorageApi
  private readonly target?: string
  private readonly apiKey?: string
  private readonly jwtToken?: string
  private readonly organizationId?: string
  private readonly apiUrl: string
  public readonly volume: VolumeService
  public readonly snapshot: SnapshotService

  /**
   * Creates a new Daytona client instance.
   *
   * @param {DaytonaConfig} [config] - Configuration options
   * @throws {DaytonaError} - `DaytonaError` - When API key is missing
   */
  constructor(config?: DaytonaConfig) {
    let apiUrl: string | undefined
    if (config) {
      this.apiKey = !config?.apiKey && config?.jwtToken ? undefined : config?.apiKey
      this.jwtToken = config?.jwtToken
      this.organizationId = config?.organizationId
      apiUrl = config?.apiUrl || config?.serverUrl
      this.target = config?.target
    }

    if (
      (!config ||
        (!(this.apiKey && apiUrl && this.target) &&
          !(this.jwtToken && this.organizationId && apiUrl && this.target))) &&
      RUNTIME !== Runtime.BROWSER
    ) {
      if (RUNTIME === Runtime.NODE) {
        const dotenv = require('dotenv')
        dotenv.config({ quiet: true })
        dotenv.config({ path: '.env.local', override: true, quiet: true })
      }
      this.apiKey = this.apiKey || (this.jwtToken ? undefined : getEnvVar('DAYTONA_API_KEY'))
      this.jwtToken = this.jwtToken || getEnvVar('DAYTONA_JWT_TOKEN')
      this.organizationId = this.organizationId || getEnvVar('DAYTONA_ORGANIZATION_ID')
      apiUrl = apiUrl || getEnvVar('DAYTONA_API_URL') || getEnvVar('DAYTONA_SERVER_URL')
      this.target = this.target || getEnvVar('DAYTONA_TARGET')

      if (getEnvVar('DAYTONA_SERVER_URL') && !getEnvVar('DAYTONA_API_URL')) {
        console.warn(
          '[Deprecation Warning] Environment variable `DAYTONA_SERVER_URL` is deprecated and will be removed in future versions. Use `DAYTONA_API_URL` instead.',
        )
      }
    }

    this.apiUrl = apiUrl || 'https://app.daytona.io/api'

    const orgHeader: Record<string, string> = {}
    if (!this.apiKey) {
      if (!this.organizationId) {
        throw new DaytonaError('Organization ID is required when using JWT token')
      }
      orgHeader['X-Daytona-Organization-ID'] = this.organizationId
    }

    const configuration = new Configuration({
      basePath: this.apiUrl,
      baseOptions: {
        headers: {
          Authorization: `Bearer ${this.apiKey || this.jwtToken}`,
          'X-Daytona-Source': 'typescript-sdk',
          'X-Daytona-SDK-Version': packageJson.version,
          ...orgHeader,
        },
      },
    })

    const axiosInstance = axios.create({
      timeout: 24 * 60 * 60 * 1000, // 24 hours
    })
    axiosInstance.interceptors.response.use(
      (response) => {
        return response
      },
      (error) => {
        let errorMessage: string

        if (error instanceof AxiosError && error.message.includes('timeout of')) {
          errorMessage = 'Operation timed out'
        } else {
          errorMessage = error.response?.data?.message || error.response?.data || error.message || String(error)
        }

        try {
          errorMessage = JSON.stringify(errorMessage)
        } catch {
          errorMessage = String(errorMessage)
        }

        switch (error.response?.data?.statusCode) {
          case 404:
            throw new DaytonaNotFoundError(errorMessage)
          default:
            throw new DaytonaError(errorMessage)
        }
      },
    )

    this.sandboxApi = new SandboxApi(configuration, '', axiosInstance)
    this.toolboxApi = new ToolboxApi(configuration, '', axiosInstance)
    this.objectStorageApi = new ObjectStorageApi(configuration, '', axiosInstance)
    this.volume = new VolumeService(new VolumesApi(configuration, '', axiosInstance))
    this.snapshot = new SnapshotService(
      configuration,
      new SnapshotsApi(configuration, '', axiosInstance),
      this.objectStorageApi,
    )
    this.clientConfig = configuration
  }

  /**
   * Creates Sandboxes from specified or default snapshot. You can specify various parameters,
   * including language, image, environment variables, and volumes.
   *
   * @param {CreateSandboxFromSnapshotParams} [params] - Parameters for Sandbox creation from snapshot
   * @param {object} [options] - Options for the create operation
   * @param {number} [options.timeout] - Timeout in seconds (0 means no timeout, default is 60)
   * @returns {Promise<Sandbox>} The created Sandbox instance
   *
   * @example
   * const sandbox = await daytona.create();
   *
   * @example
   * // Create a custom sandbox
   * const params: CreateSandboxFromSnapshotParams = {
   *     language: 'typescript',
   *     snapshot: 'my-snapshot-id',
   *     envVars: {
   *         NODE_ENV: 'development',
   *         DEBUG: 'true'
   *     },
   *     autoStopInterval: 60,
   *     autoArchiveInterval: 60,
   *     autoDeleteInterval: 120
   * };
   * const sandbox = await daytona.create(params, { timeout: 100 });
   */
  public async create(params?: CreateSandboxFromSnapshotParams, options?: { timeout?: number }): Promise<Sandbox>
  /**
   * Creates Sandboxes from specified image available on some registry or declarative Daytona Image. You can specify various parameters,
   * including resources, language, image, environment variables, and volumes. Daytona creates snapshot from
   * provided image and uses it to create Sandbox.
   *
   * @param {CreateSandboxFromImageParams} [params] - Parameters for Sandbox creation from image
   * @param {object} [options] - Options for the create operation
   * @param {number} [options.timeout] - Timeout in seconds (0 means no timeout, default is 60)
   * @param {function} [options.onSnapshotCreateLogs] - Callback function to handle snapshot creation logs.
   * @returns {Promise<Sandbox>} The created Sandbox instance
   *
   * @example
   * const sandbox = await daytona.create({ image: 'debian:12.9' }, { timeout: 90, onSnapshotCreateLogs: console.log });
   *
   * @example
   * // Create a custom sandbox
   * const image = Image.base('alpine:3.18').pipInstall('numpy');
   * const params: CreateSandboxFromImageParams = {
   *     language: 'typescript',
   *     image,
   *     envVars: {
   *         NODE_ENV: 'development',
   *         DEBUG: 'true'
   *     },
   *     resources: {
   *         cpu: 2,
   *         memory: 4 // 4GB RAM
   *     },
   *     autoStopInterval: 60,
   *     autoArchiveInterval: 60,
   *     autoDeleteInterval: 120
   * };
   * const sandbox = await daytona.create(params, { timeout: 100, onSnapshotCreateLogs: console.log });
   */
  public async create(
    params?: CreateSandboxFromImageParams,
    options?: { onSnapshotCreateLogs?: (chunk: string) => void; timeout?: number },
  ): Promise<Sandbox>
  public async create(
    params?: CreateSandboxFromSnapshotParams | CreateSandboxFromImageParams,
    options: { onSnapshotCreateLogs?: (chunk: string) => void; timeout?: number } = { timeout: 60 },
  ): Promise<Sandbox> {
    const startTime = Date.now()

    options = typeof options === 'number' ? { timeout: options } : { ...options }
    if (options.timeout == undefined || options.timeout == null) {
      options.timeout = 60
    }

    if (params == null) {
      params = { language: 'python' }
    }

    const labels = params.labels || {}
    if (params.language) {
      labels['code-toolbox-language'] = params.language
    }

    if (options.timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    if (
      params.autoStopInterval !== undefined &&
      (!Number.isInteger(params.autoStopInterval) || params.autoStopInterval < 0)
    ) {
      throw new DaytonaError('autoStopInterval must be a non-negative integer')
    }

    if (
      params.autoArchiveInterval !== undefined &&
      (!Number.isInteger(params.autoArchiveInterval) || params.autoArchiveInterval < 0)
    ) {
      throw new DaytonaError('autoArchiveInterval must be a non-negative integer')
    }

    const codeToolbox = this.getCodeToolbox(params.language as CodeLanguage)

    try {
      let buildInfo: any | undefined
      let snapshot: string | undefined
      let resources: Resources | undefined

      if ('snapshot' in params) {
        snapshot = params.snapshot
      }

      if ('image' in params) {
        if (typeof params.image === 'string') {
          buildInfo = {
            dockerfileContent: Image.base(params.image).dockerfile,
          }
        } else if (params.image instanceof Image) {
          const contextHashes = await SnapshotService.processImageContext(this.objectStorageApi, params.image)
          buildInfo = {
            contextHashes,
            dockerfileContent: params.image.dockerfile,
          }
        }
      }

      if ('resources' in params) {
        resources = params.resources
      }

      const response = await this.sandboxApi.createSandbox(
        {
          snapshot: snapshot,
          buildInfo,
          user: params.user,
          env: params.envVars || {},
          labels: labels,
          public: params.public,
          target: this.target,
          cpu: resources?.cpu,
          gpu: resources?.gpu,
          memory: resources?.memory,
          disk: resources?.disk,
          autoStopInterval: params.autoStopInterval,
          autoArchiveInterval: params.autoArchiveInterval,
          autoDeleteInterval: params.autoDeleteInterval,
          volumes: params.volumes,
        },
        undefined,
        {
          timeout: options.timeout * 1000,
        },
      )

      let sandboxInstance = response.data

      if (sandboxInstance.state === SandboxState.PENDING_BUILD && options.onSnapshotCreateLogs) {
        const terminalStates: SandboxState[] = [
          SandboxState.STARTED,
          SandboxState.STARTING,
          SandboxState.ERROR,
          SandboxState.BUILD_FAILED,
        ]

        while (sandboxInstance.state === SandboxState.PENDING_BUILD) {
          await new Promise((resolve) => setTimeout(resolve, 1000))
          sandboxInstance = (await this.sandboxApi.getSandbox(sandboxInstance.id)).data
        }

        const url = `${this.clientConfig.basePath}/sandbox/${sandboxInstance.id}/build-logs?follow=true`

        await processStreamingResponse(
          () => fetch(url, { method: 'GET', headers: this.clientConfig.baseOptions.headers }),
          (chunk) => options.onSnapshotCreateLogs?.(chunk.trimEnd()),
          async () => {
            sandboxInstance = (await this.sandboxApi.getSandbox(sandboxInstance.id)).data
            return sandboxInstance.state !== undefined && terminalStates.includes(sandboxInstance.state)
          },
        )
      }

      const sandbox = new Sandbox(sandboxInstance, this.clientConfig, this.sandboxApi, this.toolboxApi, codeToolbox)

      if (sandbox.state !== 'started') {
        const timeElapsed = Date.now() - startTime
        await sandbox.waitUntilStarted(options.timeout ? options.timeout - timeElapsed / 1000 : 0)
      }

      return sandbox
    } catch (error) {
      if (error instanceof DaytonaError && error.message.includes('Operation timed out')) {
        const errMsg = `Failed to create and start sandbox within ${options.timeout} seconds. Operation timed out.`
        throw new DaytonaError(errMsg)
      }
      throw error
    }
  }

  /**
   * Gets a Sandbox by its ID.
   *
   * @param {string} sandboxId - The ID of the Sandbox to retrieve
   * @returns {Promise<Sandbox>} The Sandbox
   *
   * @example
   * const sandbox = await daytona.get('my-sandbox-id');
   * console.log(`Sandbox state: ${sandbox.state}`);
   */
  public async get(sandboxId: string): Promise<Sandbox> {
    const response = await this.sandboxApi.getSandbox(sandboxId)
    const sandboxInstance = response.data
    const language = sandboxInstance.labels && sandboxInstance.labels['code-toolbox-language']
    const codeToolbox = this.getCodeToolbox(language as CodeLanguage)

    return new Sandbox(sandboxInstance, this.clientConfig, this.sandboxApi, this.toolboxApi, codeToolbox)
  }

  /**
   * Finds a Sandbox by its ID or labels.
   *
   * @param {SandboxFilter} filter - Filter for Sandboxes
   * @returns {Promise<Sandbox>} First Sandbox that matches the ID or labels.
   *
   * @example
   * const sandbox = await daytona.findOne({ labels: { 'my-label': 'my-value' } });
   * console.log(`Sandbox ID: ${sandbox.id}, State: ${sandbox.state}`);
   */
  public async findOne(filter: SandboxFilter): Promise<Sandbox> {
    if (filter.id) {
      return this.get(filter.id)
    }

    const sandboxes = await this.list(filter.labels)
    if (sandboxes.length === 0) {
      const errMsg = `No sandbox found with labels ${JSON.stringify(filter.labels)}`
      throw new DaytonaError(errMsg)
    }
    return sandboxes[0]
  }

  /**
   * Lists all Sandboxes filtered by labels.
   *
   * @param {Record<string, string>} [labels] - Labels to filter Sandboxes
   * @returns {Promise<Sandbox[]>} Array of Sandboxes that match the labels.
   *
   * @example
   * const sandboxes = await daytona.list({ 'my-label': 'my-value' });
   * for (const sandbox of sandboxes) {
   *     console.log(`${sandbox.id}: ${sandbox.state}`);
   * }
   */
  public async list(labels?: Record<string, string>): Promise<Sandbox[]> {
    const response = await this.sandboxApi.listSandboxes(
      undefined,
      undefined,
      labels ? JSON.stringify(labels) : undefined,
    )
    return response.data.map((sandbox) => {
      const language = sandbox.labels?.['code-toolbox-language'] as CodeLanguage

      return new Sandbox(sandbox, this.clientConfig, this.sandboxApi, this.toolboxApi, this.getCodeToolbox(language))
    })
  }

  /**
   * Starts a Sandbox and waits for it to be ready.
   *
   * @param {Sandbox} sandbox - The Sandbox to start
   * @param {number} [timeout] - Optional timeout in seconds (0 means no timeout)
   * @returns {Promise<void>}
   *
   * @example
   * const sandbox = await daytona.get('my-sandbox-id');
   * // Wait up to 60 seconds for the sandbox to start
   * await daytona.start(sandbox, 60);
   */
  public async start(sandbox: Sandbox, timeout?: number) {
    await sandbox.start(timeout)
  }

  /**
   * Stops a Sandbox.
   *
   * @param {Sandbox} sandbox - The Sandbox to stop
   * @returns {Promise<void>}
   *
   * @example
   * const sandbox = await daytona.get('my-sandbox-id');
   * await daytona.stop(sandbox);
   */
  public async stop(sandbox: Sandbox) {
    await sandbox.stop()
  }

  /**
   * Deletes a Sandbox.
   *
   * @param {Sandbox} sandbox - The Sandbox to delete
   * @param {number} timeout - Timeout in seconds (0 means no timeout, default is 60)
   * @returns {Promise<void>}
   *
   * @example
   * const sandbox = await daytona.get('my-sandbox-id');
   * await daytona.delete(sandbox);
   */
  public async delete(sandbox: Sandbox, timeout = 60) {
    await sandbox.delete(timeout)
  }

  /**
   * Gets the appropriate code toolbox based on language.
   *
   * @private
   * @param {CodeLanguage} [language] - Programming language for the toolbox
   * @returns {SandboxCodeToolbox} The appropriate code toolbox instance
   * @throws {DaytonaError} - `DaytonaError` - When an unsupported language is specified
   */
  private getCodeToolbox(language?: CodeLanguage) {
    switch (language) {
      case CodeLanguage.JAVASCRIPT:
      case CodeLanguage.TYPESCRIPT:
        return new SandboxTsCodeToolbox()
      case CodeLanguage.PYTHON:
      case undefined:
        return new SandboxPythonCodeToolbox()
      default: {
        const errMsg = `Unsupported language: ${language}, supported languages: ${Object.values(CodeLanguage).join(', ')}`
        throw new DaytonaError(errMsg)
      }
    }
  }
}
