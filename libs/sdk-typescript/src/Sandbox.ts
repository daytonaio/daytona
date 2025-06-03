/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  ToolboxApi,
  WorkspaceState as SandboxState,
  WorkspaceApi as SandboxApi,
  Workspace as ApiSandbox,
  WorkspaceInfo as ApiSandboxInfo,
  CreateNodeClassEnum as SandboxClass,
  CreateNodeRegionEnum as SandboxTargetRegion,
  Workspace as ApiWorkspace,
  PortPreviewUrl,
} from '@daytonaio/api-client'
import { FileSystem } from './FileSystem'
import { Git } from './Git'
import { CodeRunParams, Process } from './Process'
import { LspLanguageId, LspServer } from './LspServer'
import { DaytonaError } from './errors/DaytonaError'
import { prefixRelativePath } from './utils/Path'

/** @deprecated Use SandboxInfo instead. This type will be removed in a future version. */
type WorkspaceInfo = SandboxInfo

export interface SandboxInstance extends Omit<ApiSandbox, 'info'> {
  info?: SandboxInfo
}

/**
 * Resources allocated to a Sandbox
 *
 * @interface
 * @property {string} cpu - Number of CPU cores allocated (e.g., "1", "2")
 * @property {string | null} gpu - Number of GPUs allocated (e.g., "1") or null if no GPU
 * @property {string} memory - Amount of memory allocated with unit (e.g., "2Gi", "4Gi")
 * @property {string} disk - Amount of disk space allocated with unit (e.g., "10Gi", "20Gi")
 *
 * @example
 * const resources: SandboxResources = {
 *   cpu: "2",
 *   gpu: "1",
 *   memory: "4Gi",
 *   disk: "20Gi"
 * };
 */
export interface SandboxResources {
  /** CPU allocation */
  cpu: string
  /** GPU allocation */
  gpu: string | null
  /** Memory allocation */
  memory: string
  /** Disk allocation */
  disk: string
}

/**
 * Structured information about a Sandbox
 *
 * This interface provides detailed information about a Sandbox's configuration,
 * resources, and current state.
 *
 * @interface
 * @property {string} id - Unique identifier for the Sandbox
 * @property {string} [image] - Docker image used for the Sandbox
 * @property {string} user - OS user running in the Sandbox
 * @property {Record<string, string>} env - Environment variables set in the Sandbox
 * @property {Record<string, string>} labels - Custom labels attached to the Sandbox
 * @property {boolean} public - Whether the Sandbox is publicly accessible
 * @property {string} target - Target environment where the Sandbox runs
 * @property {SandboxResources} resources - Resource allocations for the Sandbox
 * @property {string} state - Current state of the Sandbox (e.g., "started", "stopped")
 * @property {string | null} errorReason - Error message if Sandbox is in error state
 * @property {string | null} snapshotState - Current state of Sandbox snapshot
 * @property {string | null} snapshotCreatedAt - When the snapshot was created
 * @property {string} nodeDomain - Domain name of the Sandbox node
 * @property {string} region - Region of the Sandbox node
 * @property {string} class - Sandbox class
 * @property {string} updatedAt - When the Sandbox was last updated
 * @property {string | null} lastSnapshot - When the last snapshot was created
 * @property {number} autoStopInterval - Auto-stop interval in minutes
 *
 * @example
 * const sandbox = await daytona.create();
 * const info = await sandbox.info();
 * console.log(`Sandbox ${info.id} is ${info.state}`);
 * console.log(`Resources: ${info.resources.cpu} CPU, ${info.resources.memory} RAM`);
 */
export interface SandboxInfo extends Omit<ApiSandboxInfo, 'name'> {
  /** Unique identifier */
  id: string
  /** Docker image */
  image?: string
  /** OS user */
  user: string
  /** Environment variables */
  env: Record<string, string>
  /** Sandbox labels */
  labels: Record<string, string>
  /** Public access flag */
  public: boolean
  /** Target location */
  target: SandboxTargetRegion | string
  /** Resource allocations */
  resources: SandboxResources
  /** Current state */
  state: SandboxState
  /** Error reason if any */
  errorReason: string | null
  /** Snapshot state */
  snapshotState: string | null
  /** Snapshot creation time */
  snapshotCreatedAt: string | null
  /** Node domain */
  nodeDomain: string
  /** Region */
  region: SandboxTargetRegion
  /** Class */
  class: SandboxClass
  /** Updated at */
  updatedAt: string
  /** Last snapshot */
  lastSnapshot: string | null
  /** Auto-stop interval in minutes*/
  autoStopInterval: number
  /**
   * @deprecated Use `state`, `nodeDomain`, `region`, `class`, `updatedAt`, `lastSnapshot`, `resources`, `autoStopInterval` instead.
   */
  providerMetadata?: string
}

/**
 * Interface defining methods that a code toolbox must implement
 * @interface
 */
export interface SandboxCodeToolbox {
  /** Generates a command to run the provided code */
  getRunCommand(code: string, params?: CodeRunParams): string
}

/**
 * Represents a Daytona Sandbox.
 *
 * @property {string} id - Unique identifier for the Sandbox
 * @property {SandboxInstance} instance - The underlying Sandbox instance
 * @property {SandboxApi} sandboxApi - API client for Sandbox operations
 * @property {ToolboxApi} toolboxApi - API client for toolbox operations
 * @property {SandboxCodeToolbox} codeToolbox - Language-specific toolbox implementation
 * @property {FileSystem} fs - File system operations interface
 * @property {Git} git - Git operations interface
 * @property {Process} process - Process execution interface
 *
 * @class
 */
export class Sandbox {
  /** File system operations for the Sandbox */
  public readonly fs: FileSystem
  /** Git operations for the Sandbox */
  public readonly git: Git
  /** Process and code execution operations */
  public readonly process: Process
  /** Default root directory for the Sandbox */
  private rootDir: string

  /**
   * Creates a new Sandbox instance
   *
   * @param {string} id - Unique identifier for the Sandbox
   * @param {SandboxInstance} instance - The underlying Sandbox instance
   * @param {SandboxApi} sandboxApi - API client for Sandbox operations
   * @param {ToolboxApi} toolboxApi - API client for toolbox operations
   * @param {SandboxCodeToolbox} codeToolbox - Language-specific toolbox implementation
   */
  constructor(
    public readonly id: string,
    public readonly instance: SandboxInstance,
    public readonly sandboxApi: SandboxApi,
    public readonly toolboxApi: ToolboxApi,
    private readonly codeToolbox: SandboxCodeToolbox,
  ) {
    this.rootDir = ''
    this.fs = new FileSystem(instance, this.toolboxApi, async () => await this.getRootDir())
    this.git = new Git(this, this.toolboxApi, instance, async () => await this.getRootDir())
    this.process = new Process(this.codeToolbox, this.toolboxApi, instance, async () => await this.getRootDir())
  }

  /**
   * Gets the root directory path for the logged in user inside the Sandbox.
   *
   * @returns {Promise<string | undefined>} The absolute path to the Sandbox root directory for the logged in user
   *
   * @example
   * const rootDir = await sandbox.getUserRootDir();
   * console.log(`Sandbox root: ${rootDir}`);
   */
  public async getUserRootDir(): Promise<string | undefined> {
    const response = await this.toolboxApi.getProjectDir(this.instance.id)
    return response.data.dir
  }

  /**
   * @deprecated Use `getUserRootDir` instead. This method will be removed in a future version.
   */
  public async getWorkspaceRootDir(): Promise<string | undefined> {
    return this.getUserRootDir()
  }

  /**
   * Creates a new Language Server Protocol (LSP) server instance.
   *
   * The LSP server provides language-specific features like code completion,
   * diagnostics, and more.
   *
   * @param {LspLanguageId} languageId - The language server type (e.g., "typescript")
   * @param {string} pathToProject - Path to the project root directory. Relative paths are resolved based on the user's
   * root directory.
   * @returns {LspServer} A new LSP server instance configured for the specified language
   *
   * @example
   * const lsp = await sandbox.createLspServer('typescript', 'workspace/project');
   */
  public async createLspServer(languageId: LspLanguageId | string, pathToProject: string): Promise<LspServer> {
    return new LspServer(
      languageId as LspLanguageId,
      prefixRelativePath(await this.getRootDir(), pathToProject),
      this.toolboxApi,
      this.instance,
    )
  }

  /**
   * Sets labels for the Sandbox.
   *
   * Labels are key-value pairs that can be used to organize and identify Sandboxes.
   *
   * @param {Record<string, string>} labels - Dictionary of key-value pairs representing Sandbox labels
   * @returns {Promise<void>}
   *
   * @example
   * // Set sandbox labels
   * await sandbox.setLabels({
   *   project: 'my-project',
   *   environment: 'development',
   *   team: 'backend'
   * });
   */
  public async setLabels(labels: Record<string, string>): Promise<void> {
    await this.sandboxApi.replaceLabels(this.instance.id, { labels })
  }

  /**
   * Start the Sandbox.
   *
   * This method starts the Sandbox and waits for it to be ready.
   *
   * @param {number} [timeout] - Maximum time to wait in seconds. 0 means no timeout.
   *                            Defaults to 60-second timeout.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - `DaytonaError` - If Sandbox fails to start or times out
   *
   * @example
   * const sandbox = await daytona.getCurrentSandbox('my-sandbox');
   * await sandbox.start(40);  // Wait up to 40 seconds
   * console.log('Sandbox started successfully');
   */
  public async start(timeout = 60): Promise<void> {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }
    const startTime = Date.now()
    await this.sandboxApi.startWorkspace(this.instance.id, undefined, { timeout: timeout * 1000 })
    const timeElapsed = Date.now() - startTime
    await this.waitUntilStarted(timeout ? timeout - timeElapsed / 1000 : 0)
  }

  /**
   * Stops the Sandbox.
   *
   * This method stops the Sandbox and waits for it to be fully stopped.
   *
   * @param {number} [timeout] - Maximum time to wait in seconds. 0 means no timeout.
   *                            Defaults to 60-second timeout.
   * @returns {Promise<void>}
   *
   * @example
   * const sandbox = await daytona.getCurrentSandbox('my-sandbox');
   * await sandbox.stop();
   * console.log('Sandbox stopped successfully');
   */
  public async stop(timeout = 60): Promise<void> {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }
    const startTime = Date.now()
    await this.sandboxApi.stopWorkspace(this.instance.id, undefined, { timeout: timeout * 1000 })
    const timeElapsed = Date.now() - startTime
    await this.waitUntilStopped(timeout ? timeout - timeElapsed / 1000 : 0)
  }

  /**
   * Deletes the Sandbox.
   * @returns {Promise<void>}
   */
  public async delete(): Promise<void> {
    await this.sandboxApi.deleteWorkspace(this.instance.id, true)
  }

  /**
   * Waits for the Sandbox to reach the 'started' state.
   *
   * This method polls the Sandbox status until it reaches the 'started' state
   * or encounters an error.
   *
   * @param {number} [timeout] - Maximum time to wait in seconds. 0 means no timeout.
   *                               Defaults to 60 seconds.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - `DaytonaError` - If the sandbox ends up in an error state or fails to start within the timeout period.
   */
  public async waitUntilStarted(timeout = 60) {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    const checkInterval = 100 // Wait 100 ms between checks
    const startTime = Date.now()

    let state: SandboxState | undefined = (await this.info()).state

    while (state !== 'started') {
      const response = await this.sandboxApi.getWorkspace(this.id)
      state = response.data.state

      if (state === 'error') {
        throw new DaytonaError(
          `Sandbox ${this.id} failed to start with status: ${state}, error reason: ${response.data.errorReason}`,
        )
      }

      if (timeout !== 0 && Date.now() - startTime > timeout * 1000) {
        throw new DaytonaError('Sandbox failed to become ready within the timeout period')
      }

      await new Promise((resolve) => setTimeout(resolve, checkInterval))
    }
  }

  /**
   * Wait for Sandbox to reach 'stopped' state.
   *
   * This method polls the Sandbox status until it reaches the 'stopped' state
   * or encounters an error.
   *
   * @param {number} [timeout] - Maximum time to wait in seconds. 0 means no timeout.
   *                               Defaults to 60 seconds.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - `DaytonaError` - If the sandbox fails to stop within the timeout period.
   */
  public async waitUntilStopped(timeout = 60) {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    const checkInterval = 100 // Wait 100 ms between checks
    const startTime = Date.now()

    let state: SandboxState | undefined = (await this.info()).state

    while (state !== 'stopped') {
      const response = await this.sandboxApi.getWorkspace(this.id)
      state = response.data.state

      if (state === 'error') {
        throw new DaytonaError(
          `Sandbox failed to stop with status: ${state}, error reason: ${response.data.errorReason}`,
        )
      }

      if (timeout !== 0 && Date.now() - startTime > timeout * 1000) {
        throw new DaytonaError('Sandbox failed to become stopped within the timeout period')
      }

      await new Promise((resolve) => setTimeout(resolve, checkInterval))
    }
  }

  /**
   * Gets structured information about the Sandbox.
   *
   * @returns {Promise<SandboxInfo>} Detailed information about the Sandbox including its
   *                                   configuration, resources, and current state
   *
   * @example
   * const info = await sandbox.info();
   * console.log(`Sandbox ${info.id}:`);
   * console.log(`State: ${info.state}`);
   * console.log(`Resources: ${info.resources.cpu} CPU, ${info.resources.memory} RAM`);
   */
  public async info(): Promise<SandboxInfo> {
    const response = await this.sandboxApi.getWorkspace(this.id)
    const instance = response.data
    return Sandbox.toSandboxInfo(instance)
  }

  /**
   * Converts an API workspace instance to a WorkspaceInfo object.
   *
   * @param {ApiWorkspace} instance - The API workspace instance to convert
   * @returns {WorkspaceInfo} The converted WorkspaceInfo object
   *
   * @deprecated Use `toSandboxInfo` instead. This method will be removed in a future version.
   */
  public static toWorkspaceInfo(instance: ApiWorkspace): WorkspaceInfo {
    return Sandbox.toSandboxInfo(instance)
  }
  /**
   * Converts an API sandbox instance to a SandboxInfo object.
   *
   * @param {ApiSandbox} instance - The API sandbox instance to convert
   * @returns {SandboxInfo} The converted SandboxInfo object
   */
  public static toSandboxInfo(instance: ApiSandbox): SandboxInfo {
    const providerMetadata = JSON.parse(instance.info?.providerMetadata || '{}')

    // Extract resources with defaults
    const resources: SandboxResources = {
      cpu: String(instance.cpu || '1'),
      gpu: instance.gpu ? String(instance.gpu) : null,
      memory: `${instance.memory ?? 2}Gi`,
      disk: `${instance.disk ?? 10}Gi`,
    }

    return {
      id: instance.id,
      image: instance.image,
      user: instance.user,
      env: instance.env || {},
      labels: instance.labels || {},
      public: instance.public || false,
      target: instance.target,
      resources,
      state: instance.state || SandboxState.UNKNOWN,
      errorReason: instance.errorReason || null,
      snapshotState: instance.snapshotState || null,
      snapshotCreatedAt: instance.snapshotCreatedAt || null,
      autoStopInterval: instance.autoStopInterval || 15,
      created: instance.info?.created || '',
      nodeDomain: providerMetadata.nodeDomain || '',
      region: providerMetadata.region || '',
      class: providerMetadata.class || '',
      updatedAt: providerMetadata.updatedAt || '',
      lastSnapshot: providerMetadata.lastSnapshot || null,
      providerMetadata: instance.info?.providerMetadata,
    }
  }

  /**
   * Set the auto-stop interval for the Sandbox.
   *
   * The Sandbox will automatically stop after being idle (no new events) for the specified interval.
   * Events include any state changes or interactions with the Sandbox through the sdk.
   * Interactions using Sandbox Previews are not included.
   *
   * @param {number} interval - Number of minutes of inactivity before auto-stopping.
   *                           Set to 0 to disable auto-stop. Default is 15 minutes.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - `DaytonaError` - If interval is not a non-negative integer
   *
   * @example
   * // Auto-stop after 1 hour
   * await sandbox.setAutostopInterval(60);
   * // Or disable auto-stop
   * await sandbox.setAutostopInterval(0);
   */
  public async setAutostopInterval(interval: number): Promise<void> {
    if (!Number.isInteger(interval) || interval < 0) {
      throw new DaytonaError('autoStopInterval must be a non-negative integer')
    }

    await this.sandboxApi.setAutostopInterval(this.id, interval)
    this.instance.autoStopInterval = interval
  }

  /**
   * Retrieves the preview link for the sandbox at the specified port. If the port is closed,
   * it will be opened automatically. For private sandboxes, a token is included to grant access
   * to the URL.
   *
   * @param {number} port - The port to open the preview link on.
   * @returns {PortPreviewUrl} The response object for the preview link, which includes the `url`
   * and the `token` (to access private sandboxes).
   *
   * @example
   * const previewLink = await sandbox.getPreviewLink(3000);
   * console.log(`Preview URL: ${previewLink.url}`);
   * console.log(`Token: ${previewLink.token}`);
   */
  public async getPreviewLink(port: number): Promise<PortPreviewUrl> {
    return (await this.sandboxApi.getPortPreviewUrl(this.id, port)).data
  }

  /**
   * Archives the sandbox, making it inactive and preserving its state. When sandboxes are archived, the entire filesystem
   * state is moved to cost-effective object storage, making it possible to keep sandboxes available for an extended period.
   * The tradeoff between archived and stopped states is that starting an archived sandbox takes more time, depending on its size.
   * Sandbox must be stopped before archiving.
   */
  public async archive(): Promise<void> {
    await this.sandboxApi.archiveWorkspace(this.id)
  }

  private async getRootDir(): Promise<string> {
    if (!this.rootDir) {
      this.rootDir = (await this.getUserRootDir()) || ''
    }
    return this.rootDir
  }
}
