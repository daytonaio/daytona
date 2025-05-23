/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Configuration,
  ImagesApi,
  ImageState,
  ObjectStorageApi,
  WorkspaceApi as SandboxApi,
  WorkspaceState as SandboxState,
  CreateWorkspaceTargetEnum as SandboxTargetRegion,
  ToolboxApi,
  VolumesApi,
  WorkspaceVolume,
} from '@daytonaio/api-client'
import axios, { AxiosError } from 'axios'
import * as dotenv from 'dotenv'
import { SandboxPythonCodeToolbox } from './code-toolbox/SandboxPythonCodeToolbox'
import { SandboxTsCodeToolbox } from './code-toolbox/SandboxTsCodeToolbox'
import { DaytonaError, DaytonaNotFoundError } from './errors/DaytonaError'
import { Image } from './Image'
import { ObjectStorage } from './ObjectStorage'
import { Sandbox, SandboxInstance, Sandbox as Workspace } from './Sandbox'
import { processStreamingResponse } from './utils/Stream'
import { VolumeService } from './Volume'

/**
 * Represents a volume mount for a Sandbox.
 *
 * @interface
 * @property {string} volumeId - ID of the Volume to mount
 * @property {string} mountPath - Path on the Sandbox to mount the Volume
 */

export interface VolumeMount extends WorkspaceVolume {
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
 * @property {CreateSandboxTargetEnum} target - Target location for Sandboxes
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
  target?: SandboxTargetRegion
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
 * @property {number} [memory] - Memory allocation for the Sandbox in GB
 * @property {number} [disk] - Disk space allocation for the Sandbox in GB
 *
 * @example
 * const resources: SandboxResources = {
 *     cpu: 2,
 *     memory: 4,  // 4GB RAM
 *     disk: 20    // 20GB disk
 * };
 */
export interface SandboxResources {
  /** CPU allocation for the Sandbox */
  cpu?: number
  /** GPU allocation for the Sandbox */
  gpu?: number
  /** Memory allocation for the Sandbox in GB */
  memory?: number
  /** Disk space allocation for the Sandbox in GB */
  disk?: number
}

/**
 * Parameters for creating a new Sandbox.
 *
 * @interface
 * @property {string | Image} [image] - Optional Docker image to use for the Sandbox or an Image instance
 * @property {string} [user] - Optional os user to use for the Sandbox
 * @property {CodeLanguage | string} [language] - Programming language for direct code execution
 * @property {Record<string, string>} [envVars] - Optional environment variables to set in the Sandbox
 * @property {Record<string, string>} [labels] - Sandbox labels
 * @property {boolean} [public] - Is the Sandbox port preview public
 * @property {SandboxResources} [resources] - Resource allocation for the Sandbox
 * @property {boolean} [async] - If true, will not wait for the Sandbox to be ready before returning
 * @property {number} [timeout] - Timeout in seconds for the Sandbox to be ready (0 means no timeout)
 * @property {number} [autoStopInterval] - Auto-stop interval in minutes (0 means disabled)
 *
 * @example
 * const params: CreateSandboxParams = {
 *     language: 'typescript',
 *     envVars: { NODE_ENV: 'development' },
 *     resources: {
 *         cpu: 2,
 *         memory: 4 // 4GB RAM
 *     },
 *     autoStopInterval: 60  // Auto-stop after 1 hour of inactivity
 * };
 * const sandbox = await daytona.create(params, 50);
 */
export type CreateSandboxParams = {
  /** Optional Docker image to use for the Sandbox or an Image instance */
  image?: string | Image
  /** Optional os user to use for the Sandbox */
  user?: string
  /** Programming language for direct code execution */
  language?: CodeLanguage | string
  /** Optional environment variables to set in the sandbox */
  envVars?: Record<string, string>
  /** Sandbox labels */
  labels?: Record<string, string>
  /** Is the Sandbox port preview public */
  public?: boolean
  /** Resource allocation for the Sandbox */
  resources?: SandboxResources
  /** If true, will not wait for the Sandbox to be ready before returning */
  async?: boolean
  /**
   * Timeout in seconds, for the Sandbox to be ready (0 means no timeout)
   * @deprecated Use methods with `timeout` parameter instead
   */
  timeout?: number
  /** Auto-stop interval in minutes (0 means disabled) (must be a non-negative integer) */
  autoStopInterval?: number
  /** List of volumes to mount in the Sandbox */
  volumes?: VolumeMount[]
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
  private readonly sandboxApi: SandboxApi
  private readonly toolboxApi: ToolboxApi
  private readonly imagesApi: ImagesApi
  private readonly objectStorageApi: ObjectStorageApi
  private readonly target: SandboxTargetRegion
  private readonly apiKey?: string
  private readonly jwtToken?: string
  private readonly organizationId?: string
  private readonly apiUrl: string
  public readonly volume: VolumeService

  /**
   * Creates a new Daytona client instance.
   *
   * @param {DaytonaConfig} [config] - Configuration options
   * @throws {DaytonaError} - `DaytonaError` - When API key is missing
   */
  constructor(config?: DaytonaConfig) {
    this.remove = this.delete.bind(this)

    dotenv.config()
    dotenv.config({ path: '.env.local', override: true })
    const apiKey = !config?.apiKey && config?.jwtToken ? undefined : config?.apiKey || process?.env['DAYTONA_API_KEY']
    const jwtToken = config?.jwtToken || process?.env['DAYTONA_JWT_TOKEN']
    const organizationId = config?.organizationId || process?.env['DAYTONA_ORGANIZATION_ID']
    if (!apiKey && !jwtToken) {
      throw new DaytonaError('API key or JWT token is required')
    }
    const apiUrl =
      config?.apiUrl ||
      config?.serverUrl ||
      process?.env['DAYTONA_API_URL'] ||
      process?.env['DAYTONA_SERVER_URL'] ||
      'https://app.daytona.io/api'
    const envTarget = process?.env['DAYTONA_TARGET'] as SandboxTargetRegion
    const target = config?.target || envTarget || SandboxTargetRegion.US

    if (process?.env['DAYTONA_SERVER_URL'] && !process?.env['DAYTONA_API_URL']) {
      console.warn(
        '[Deprecation Warning] Environment variable `DAYTONA_SERVER_URL` is deprecated and will be removed in future versions. Use `DAYTONA_API_URL` instead.',
      )
    }

    this.apiKey = apiKey
    this.jwtToken = jwtToken
    this.organizationId = organizationId
    this.apiUrl = apiUrl
    this.target = target

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
    this.volume = new VolumeService(new VolumesApi(configuration, '', axiosInstance))
    this.imagesApi = new ImagesApi(configuration, '', axiosInstance)
    this.objectStorageApi = new ObjectStorageApi(configuration, '', axiosInstance)
  }

  /**
   * @deprecated Use `create` with `options` object instead. This method will be removed in a future version.
   *
   * Creates Sandboxes with default or custom configurations. You can specify various parameters,
   * including language, image, resources, environment variables, and volumes for the Sandbox.
   *
   * @param {CreateSandboxParams} [params] - Parameters for Sandbox creation
   * @param {number} [timeout] - Timeout in seconds (0 means no timeout, default is 60)
   * @returns {Promise<Sandbox>} The created Sandbox instance
   *
   * @example
   * // Create a default sandbox
   * const sandbox = await daytona.create();
   *
   * @example
   * // Create a custom sandbox
   * const params: CreateSandboxParams = {
   *     language: 'typescript',
   *     image: 'node:18',
   *     envVars: {
   *         NODE_ENV: 'development',
   *         DEBUG: 'true'
   *     },
   *     resources: {
   *         cpu: 2,
   *         memory: 4 // 4GB RAM
   *     },
   *     autoStopInterval: 60
   * };
   * const sandbox = await daytona.create(params, 40);
   */
  public async create(params?: CreateSandboxParams, options?: number): Promise<Sandbox>
  /**
   * Creates Sandboxes with default or custom configurations. You can specify various parameters,
   * including language, image, resources, environment variables, and volumes for the Sandbox.
   *
   * @param {CreateSandboxParams} [params] - Parameters for Sandbox creation
   * @param {object} [options] - Options for the create operation
   * @param {number} [options.timeout] - Timeout in seconds (0 means no timeout, default is 60)
   * @param {function} [options.onImageBuildLogs] - Callback function to handle image build logs.
   * It's invoked only when `params.image` is an instance of `Image` and there's no existing
   * image in Daytona with the same configuration.
   * @returns {Promise<Sandbox>} The created Sandbox instance
   *
   * @example
   * const image = Image.debianSlim('3.12').pipInstall('numpy');
   * const sandbox = await daytona.create({ image }, { timeout: 90, onImageBuildLogs: console.log });
   *
   * @example
   * // Create a custom sandbox
   * const image = Image.debianSlim('3.12').pipInstall('numpy');
   * const params: CreateSandboxParams = {
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
   *     autoStopInterval: 60
   * };
   * const sandbox = await daytona.create(params, { timeout: 100, onImageBuildLogs: console.log });
   */
  public async create(
    params?: CreateSandboxParams,
    options?: { onImageBuildLogs?: (chunk: string) => void; timeout?: number },
  ): Promise<Sandbox>
  public async create(
    params?: CreateSandboxParams,
    options: number | { onImageBuildLogs?: (chunk: string) => void; timeout?: number } = { timeout: 60 },
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

    // remove this when params.timeout is removed
    const effectiveTimeout = params.timeout || options.timeout
    if (effectiveTimeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    if (
      params.autoStopInterval !== undefined &&
      (!Number.isInteger(params.autoStopInterval) || params.autoStopInterval < 0)
    ) {
      throw new DaytonaError('autoStopInterval must be a non-negative integer')
    }

    const codeToolbox = this.getCodeToolbox(params.language as CodeLanguage)

    try {
      // Handle Image instance if provided
      let imageStr: string | undefined
      let buildInfo: any | undefined

      if (typeof params.image === 'string') {
        imageStr = params.image
      } else if (params.image instanceof Image) {
        const contextHashes = await this.processImageContext(params.image)
        buildInfo = {
          contextHashes,
          dockerfileContent: params.image.dockerfile,
        }
      }

      const response = await this.sandboxApi.createWorkspace(
        {
          image: imageStr,
          buildInfo,
          user: params.user,
          env: params.envVars || {},
          labels: params.labels,
          public: params.public,
          target: this.target,
          cpu: params.resources?.cpu,
          gpu: params.resources?.gpu,
          memory: params.resources?.memory,
          disk: params.resources?.disk,
          autoStopInterval: params.autoStopInterval,
          volumes: params.volumes,
        },
        undefined,
        {
          timeout: effectiveTimeout * 1000,
        },
      )

      let sandboxInstance = response.data

      if (sandboxInstance.state === SandboxState.PENDING_BUILD && options.onImageBuildLogs) {
        const terminalStates: SandboxState[] = [SandboxState.STARTED, SandboxState.STARTING, SandboxState.ERROR]

        while (sandboxInstance.state === SandboxState.PENDING_BUILD) {
          await new Promise((resolve) => setTimeout(resolve, 1000))
          sandboxInstance = (await this.sandboxApi.getWorkspace(sandboxInstance.id)).data
        }

        await processStreamingResponse(
          () => this.sandboxApi.getBuildLogs(sandboxInstance.id, undefined, true, { responseType: 'stream' }),
          options.onImageBuildLogs,
          async () => {
            sandboxInstance = (await this.sandboxApi.getWorkspace(sandboxInstance.id)).data
            return sandboxInstance.state !== undefined && terminalStates.includes(sandboxInstance.state)
          },
        )
      }

      const sandboxInfo = Sandbox.toSandboxInfo(sandboxInstance)
      sandboxInstance.info = {
        ...sandboxInfo,
        name: '',
      }

      const sandbox = new Sandbox(
        sandboxInstance.id,
        sandboxInstance as SandboxInstance,
        this.sandboxApi,
        this.toolboxApi,
        codeToolbox,
      )

      if (!params.async && sandbox.instance.state !== 'started') {
        const timeElapsed = Date.now() - startTime
        await sandbox.waitUntilStarted(effectiveTimeout ? effectiveTimeout - timeElapsed / 1000 : 0)
      }

      return sandbox
    } catch (error) {
      if (error instanceof DaytonaError && error.message.includes('Operation timed out')) {
        throw new DaytonaError(
          `Failed to create and start sandbox within ${effectiveTimeout} seconds. Operation timed out.`,
        )
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
   * console.log(`Sandbox state: ${sandbox.instance.state}`);
   */
  public async get(sandboxId: string): Promise<Sandbox> {
    const response = await this.sandboxApi.getWorkspace(sandboxId)
    const sandboxInstance = response.data
    const language = sandboxInstance.labels && sandboxInstance.labels['code-toolbox-language']
    const codeToolbox = this.getCodeToolbox(language as CodeLanguage)
    const sandboxInfo = Sandbox.toSandboxInfo(sandboxInstance)
    sandboxInstance.info = {
      ...sandboxInfo,
      name: '',
    }

    return new Sandbox(sandboxId, sandboxInstance as SandboxInstance, this.sandboxApi, this.toolboxApi, codeToolbox)
  }

  /**
   * Finds a Sandbox by its ID or labels.
   *
   * @param {SandboxFilter} filter - Filter for Sandboxes
   * @returns {Promise<Sandbox>} First Sandbox that matches the ID or labels.
   *
   * @example
   * const sandbox = await daytona.findOne({ labels: { 'my-label': 'my-value' } });
   * console.log(`Sandbox: ${await sandbox.info()}`);
   */
  public async findOne(filter: SandboxFilter): Promise<Sandbox> {
    if (filter.id) {
      return this.get(filter.id)
    }

    const sandboxes = await this.list(filter.labels)
    if (sandboxes.length === 0) {
      throw new DaytonaError(`No sandbox found with labels ${JSON.stringify(filter.labels)}`)
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
   *     console.log(`${sandbox.id}: ${sandbox.instance.state}`);
   * }
   */
  public async list(labels?: Record<string, string>): Promise<Sandbox[]> {
    const response = await this.sandboxApi.listWorkspaces(
      undefined,
      undefined,
      labels ? JSON.stringify(labels) : undefined,
    )
    return response.data.map((sandbox) => {
      const language = sandbox.labels?.['code-toolbox-language'] as CodeLanguage
      const sandboxInfo = Sandbox.toSandboxInfo(sandbox)
      sandbox.info = {
        ...sandboxInfo,
        name: '',
      }

      return new Sandbox(
        sandbox.id,
        sandbox as SandboxInstance,
        this.sandboxApi,
        this.toolboxApi,
        this.getCodeToolbox(language),
      )
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
    await this.sandboxApi.deleteWorkspace(sandbox.id, true, undefined, { timeout: timeout * 1000 })
  }

  /** @hidden */
  public remove!: (sandbox: Sandbox, timeout?: number) => Promise<void>

  /**
   * Gets the Sandbox by ID.
   *
   * @param {string} workspaceId - The ID of the Sandbox to retrieve
   * @returns {Promise<Workspace>} The Sandbox
   *
   * @deprecated Use `getCurrentSandbox` instead. This method will be removed in a future version.
   */
  public async getCurrentWorkspace(workspaceId: string): Promise<Workspace> {
    return await this.getCurrentSandbox(workspaceId)
  }

  /**
   * Gets the Sandbox by ID.
   *
   * @param {string} sandboxId - The ID of the Sandbox to retrieve
   * @returns {Promise<Sandbox>} The Sandbox
   *
   * @example
   * const sandbox = await daytona.getCurrentSandbox('my-sandbox-id');
   * console.log(`Current sandbox state: ${sandbox.instance.state}`);
   */
  public async getCurrentSandbox(sandboxId: string): Promise<Sandbox> {
    return await this.get(sandboxId)
  }

  /**
   * Creates and registers a new image from the given Image definition.
   *
   * @param {string} name - The name of the image to create.
   * @param {Image} image - The Image instance.
   * @param {object} options - Options for the create operation.
   * @param {boolean} options.verbose - Default is false. Whether to log progress information upon each state change of the image.
   * @param {number} options.timeout - Default is no timeout. Timeout in seconds (0 means no timeout).
   * @returns {Promise<void>}
   *
   * @example
   * const image = Image.debianSlim('3.12').pipInstall('numpy');
   * await daytona.createImage('my-python-image', image);
   */
  public async createImage(
    name: string,
    image: Image,
    options: { onLogs?: (chunk: string) => void; timeout?: number } = {},
  ): Promise<void> {
    const contextHashes = await this.processImageContext(image)
    let builtImage = (
      await this.imagesApi.buildImage(
        {
          name,
          buildInfo: {
            contextHashes,
            dockerfileContent: image.dockerfile,
          },
        },
        undefined,
        {
          timeout: (options.timeout || 0) * 1000,
        },
      )
    ).data

    const terminalStates: ImageState[] = [ImageState.ACTIVE, ImageState.ERROR]
    const imageRef = { builtImage }
    let streamPromise: Promise<void> | undefined
    const startLogStreaming = async () => {
      if (!streamPromise) {
        streamPromise = processStreamingResponse(
          () => this.imagesApi.getImageBuildLogs(builtImage.id, undefined, true, { responseType: 'stream' }),
          options.onLogs!,
          async () => terminalStates.includes(imageRef.builtImage.state),
        )
      }
    }

    if (options.onLogs) {
      options.onLogs(`Building image ${builtImage.name} (${builtImage.state})`)

      if (builtImage.state !== ImageState.BUILD_PENDING) {
        await startLogStreaming()
      }
    }

    let previousState = builtImage.state
    while (!terminalStates.includes(builtImage.state)) {
      if (options.onLogs && previousState !== builtImage.state) {
        if (builtImage.state !== ImageState.BUILD_PENDING && !streamPromise) {
          await startLogStreaming()
        }
        options.onLogs(`Building image ${builtImage.name} (${builtImage.state})`)
        previousState = builtImage.state
      }
      await new Promise((resolve) => setTimeout(resolve, 1000))
      builtImage = (await this.imagesApi.getImage(builtImage.id)).data
      imageRef.builtImage = builtImage
    }

    if (options.onLogs) {
      if (streamPromise) {
        await streamPromise
      }
      if (builtImage.state === ImageState.ACTIVE) {
        options.onLogs(`Built image ${builtImage.name} (${builtImage.state})`)
      }
    }

    if (builtImage.state === ImageState.ERROR) {
      throw new DaytonaError(
        `Failed to build image. Image ended in the ERROR state. name: ${builtImage.name}; error reason: ${builtImage.errorReason}`,
      )
    }
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
      default:
        throw new DaytonaError(
          `Unsupported language: ${language}, supported languages: ${Object.values(CodeLanguage).join(', ')}`,
        )
    }
  }

  /**
   * Processes the image contexts by uploading them to object storage
   *
   * @private
   * @param {Image} image - The Image instance.
   * @returns {Promise<string[]>} The list of context hashes stored in object storage.
   */
  private async processImageContext(image: Image): Promise<string[]> {
    if (!image.contextList || !image.contextList.length) {
      return []
    }

    const pushAccessCreds = (await this.objectStorageApi.getPushAccess()).data
    const objectStorage = new ObjectStorage({
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
