/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  SandboxState,
  SandboxApi,
  Sandbox as SandboxDto,
  PaginatedSandboxes as PaginatedSandboxesDto,
  PortPreviewUrl,
  SandboxVolume,
  BuildInfo,
  SandboxBackupStateEnum,
  Configuration,
  SshAccessDto,
  SshAccessValidationDto,
  SignedPortPreviewUrl,
  ResizeSandbox,
} from '@daytonaio/api-client'
import { Resources } from './Daytona'
import {
  FileSystemApi,
  GitApi,
  ProcessApi,
  LspApi,
  InfoApi,
  ComputerUseApi,
  InterpreterApi,
} from '@daytonaio/toolbox-api-client'
import { FileSystem } from './FileSystem'
import { Git } from './Git'
import { CodeRunParams, Process } from './Process'
import { LspLanguageId, LspServer } from './LspServer'
import { DaytonaError, DaytonaNotFoundError } from './errors/DaytonaError'
import { ComputerUse } from './ComputerUse'
import { AxiosInstance } from 'axios'
import { CodeInterpreter } from './CodeInterpreter'
import { WithInstrumentation } from './utils/otel.decorator'

const TOOLBOX_URL_PLACEHOLDER = 'dtn-placeholder'

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
 * @property {FileSystem} fs - File system operations interface
 * @property {Git} git - Git operations interface
 * @property {Process} process - Process execution interface
 * @property {CodeInterpreter} codeInterpreter - Stateful interpreter interface for executing code.
 *   Currently supports only Python. For other languages, use the `process.codeRun` method.
 * @property {ComputerUse} computerUse - Computer use operations interface for desktop automation
 * @property {string} id - Unique identifier for the Sandbox
 * @property {string} organizationId - Organization ID of the Sandbox
 * @property {string} [snapshot] - Daytona snapshot used to create the Sandbox
 * @property {string} user - OS user running in the Sandbox
 * @property {Record<string, string>} env - Environment variables set in the Sandbox
 * @property {Record<string, string>} labels - Custom labels attached to the Sandbox
 * @property {boolean} public - Whether the Sandbox is publicly accessible
 * @property {string} target - Target location of the runner where the Sandbox runs
 * @property {number} cpu - Number of CPUs allocated to the Sandbox
 * @property {number} gpu - Number of GPUs allocated to the Sandbox
 * @property {number} memory - Amount of memory allocated to the Sandbox in GiB
 * @property {number} disk - Amount of disk space allocated to the Sandbox in GiB
 * @property {SandboxState} state - Current state of the Sandbox (e.g., "started", "stopped")
 * @property {string} [errorReason] - Error message if Sandbox is in error state
 * @property {boolean} [recoverable] - Whether the Sandbox error is recoverable.
 * @property {SandboxBackupStateEnum} [backupState] - Current state of Sandbox backup
 * @property {string} [backupCreatedAt] - When the backup was created
 * @property {number} [autoStopInterval] - Auto-stop interval in minutes
 * @property {number} [autoArchiveInterval] - Auto-archive interval in minutes
 * @property {number} [autoDeleteInterval] - Auto-delete interval in minutes
 * @property {Array<SandboxVolume>} [volumes] - Volumes attached to the Sandbox
 * @property {BuildInfo} [buildInfo] - Build information for the Sandbox if it was created from dynamic build
 * @property {string} [createdAt] - When the Sandbox was created
 * @property {string} [updatedAt] - When the Sandbox was last updated
 * @property {boolean} networkBlockAll - Whether to block all network access for the Sandbox
 * @property {string} [networkAllowList] - Comma-separated list of allowed CIDR network addresses for the Sandbox
 *
 * @class
 */
export class Sandbox implements SandboxDto {
  public readonly fs: FileSystem
  public readonly git: Git
  public readonly process: Process
  public readonly computerUse: ComputerUse
  public readonly codeInterpreter: CodeInterpreter

  public id!: string
  public name!: string
  public organizationId!: string
  public snapshot?: string
  public user!: string
  public env!: Record<string, string>
  public labels!: Record<string, string>
  public public!: boolean
  public target!: string
  public cpu!: number
  public gpu!: number
  public memory!: number
  public disk!: number
  public state?: SandboxState
  public errorReason?: string
  public recoverable?: boolean
  public backupState?: SandboxBackupStateEnum
  public backupCreatedAt?: string
  public autoStopInterval?: number
  public autoArchiveInterval?: number
  public autoDeleteInterval?: number
  public volumes?: Array<SandboxVolume>
  public buildInfo?: BuildInfo
  public createdAt?: string
  public updatedAt?: string
  public networkBlockAll!: boolean
  public networkAllowList?: string

  private infoApi: InfoApi

  /**
   * Creates a new Sandbox instance
   *
   * @param {SandboxDto} sandboxDto - The API Sandbox instance
   * @param {SandboxApi} sandboxApi - API client for Sandbox operations
   * @param {InfoApi} infoApi - API client for info operations
   * @param {SandboxCodeToolbox} codeToolbox - Language-specific toolbox implementation
   */
  constructor(
    sandboxDto: SandboxDto,
    private readonly clientConfig: Configuration,
    private readonly axiosInstance: AxiosInstance,
    private readonly sandboxApi: SandboxApi,
    private readonly codeToolbox: SandboxCodeToolbox,
    private readonly getToolboxBaseUrl: (sandboxId: string, regionId: string) => Promise<string>,
  ) {
    this.processSandboxDto(sandboxDto)

    // Lazy load the base URL for the toolbox
    this.axiosInstance.defaults.baseURL = TOOLBOX_URL_PLACEHOLDER
    this.axiosInstance.interceptors.request.use(async (config) => {
      if (this.axiosInstance.defaults.baseURL === TOOLBOX_URL_PLACEHOLDER) {
        await this.ensureToolboxUrl()

        config.baseURL = this.axiosInstance.defaults.baseURL
      }
      return config
    })

    // Initialize Services
    const getPreviewToken = async () => (await this.getPreviewLink(1)).token

    this.fs = new FileSystem(
      this.clientConfig,
      new FileSystemApi(this.clientConfig, '', this.axiosInstance),
      this.ensureToolboxUrl.bind(this),
    )
    this.git = new Git(new GitApi(this.clientConfig, '', this.axiosInstance))
    this.process = new Process(
      this.clientConfig,
      this.codeToolbox,
      new ProcessApi(this.clientConfig, '', this.axiosInstance),
      getPreviewToken,
      this.ensureToolboxUrl.bind(this),
    )
    this.codeInterpreter = new CodeInterpreter(
      this.clientConfig,
      new InterpreterApi(this.clientConfig, '', this.axiosInstance),
      getPreviewToken,
      this.ensureToolboxUrl.bind(this),
    )
    this.computerUse = new ComputerUse(new ComputerUseApi(this.clientConfig, '', this.axiosInstance))
    this.infoApi = new InfoApi(this.clientConfig, '', this.axiosInstance)
  }

  /**
   * Gets the user's home directory path for the logged in user inside the Sandbox.
   *
   * @returns {Promise<string | undefined>} The absolute path to the Sandbox user's home directory for the logged in user
   *
   * @example
   * const userHomeDir = await sandbox.getUserHomeDir();
   * console.log(`Sandbox user home: ${userHomeDir}`);
   */
  @WithInstrumentation()
  public async getUserHomeDir(): Promise<string | undefined> {
    const response = await this.infoApi.getUserHomeDir()
    return response.data.dir
  }

  /**
   * @deprecated Use `getUserHomeDir` instead. This method will be removed in a future version.
   */
  @WithInstrumentation()
  public async getUserRootDir(): Promise<string | undefined> {
    return this.getUserHomeDir()
  }

  /**
   * Gets the working directory path inside the Sandbox.
   *
   * @returns {Promise<string | undefined>} The absolute path to the Sandbox working directory. Uses the WORKDIR specified
   * in the Dockerfile if present, or falling back to the user's home directory if not.
   *
   * @example
   * const workDir = await sandbox.getWorkDir();
   * console.log(`Sandbox working directory: ${workDir}`);
   */
  @WithInstrumentation()
  public async getWorkDir(): Promise<string | undefined> {
    const response = await this.infoApi.getWorkDir()
    return response.data.dir
  }

  /**
   * Creates a new Language Server Protocol (LSP) server instance.
   *
   * The LSP server provides language-specific features like code completion,
   * diagnostics, and more.
   *
   * @param {LspLanguageId} languageId - The language server type (e.g., "typescript")
   * @param {string} pathToProject - Path to the project root directory. Relative paths are resolved based on the sandbox working directory.
   * @returns {LspServer} A new LSP server instance configured for the specified language
   *
   * @example
   * const lsp = await sandbox.createLspServer('typescript', 'workspace/project');
   */
  @WithInstrumentation()
  public async createLspServer(languageId: LspLanguageId | string, pathToProject: string): Promise<LspServer> {
    return new LspServer(
      languageId as LspLanguageId,
      pathToProject,
      new LspApi(this.clientConfig, '', this.axiosInstance),
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
  @WithInstrumentation()
  public async setLabels(labels: Record<string, string>): Promise<Record<string, string>> {
    this.labels = (await this.sandboxApi.replaceLabels(this.id, { labels })).data.labels
    return this.labels
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
  @WithInstrumentation()
  public async start(timeout = 60): Promise<void> {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    const startTime = Date.now()
    const response = await this.sandboxApi.startSandbox(this.id, undefined, { timeout: timeout * 1000 })
    this.processSandboxDto(response.data)
    const timeElapsed = Date.now() - startTime
    await this.waitUntilStarted(timeout ? Math.max(0.001, timeout - timeElapsed / 1000) : timeout)
  }

  /**
   * Recover the Sandbox from a recoverable error and wait for it to be ready.
   *
   * @param {number} [timeout] - Maximum time to wait in seconds. 0 means no timeout.
   *                            Defaults to 60-second timeout.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - `DaytonaError` - If Sandbox fails to recover or times out
   *
   * @example
   * const sandbox = await daytona.get('my-sandbox-id');
   * await sandbox.recover();
   * console.log('Sandbox recovered successfully');
   */
  public async recover(timeout = 60): Promise<void> {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    const startTime = Date.now()
    const response = await this.sandboxApi.recoverSandbox(this.id, undefined, { timeout: timeout * 1000 })
    this.processSandboxDto(response.data)
    const timeElapsed = Date.now() - startTime
    await this.waitUntilStarted(timeout ? Math.max(0.001, timeout - timeElapsed / 1000) : timeout)
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
   * const sandbox = await daytona.get('my-sandbox-id');
   * await sandbox.stop();
   * console.log('Sandbox stopped successfully');
   */
  @WithInstrumentation()
  public async stop(timeout = 60): Promise<void> {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }
    const startTime = Date.now()
    await this.sandboxApi.stopSandbox(this.id, undefined, { timeout: timeout * 1000 })
    await this.refreshDataSafe()
    const timeElapsed = Date.now() - startTime
    await this.waitUntilStopped(timeout ? Math.max(0.001, timeout - timeElapsed / 1000) : timeout)
  }

  /**
   * Deletes the Sandbox.
   * @returns {Promise<void>}
   */
  @WithInstrumentation()
  public async delete(timeout = 60): Promise<void> {
    await this.sandboxApi.deleteSandbox(this.id, undefined, { timeout: timeout * 1000 })
    this.refreshDataSafe()
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
  @WithInstrumentation()
  public async waitUntilStarted(timeout = 60) {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    const checkInterval = 100 // Wait 100 ms between checks
    const startTime = Date.now()

    while (this.state !== 'started') {
      await this.refreshData()

      // @ts-expect-error this.refreshData() can modify this.state so this check is fine
      if (this.state === 'started') {
        return
      }

      if (this.state === 'error') {
        const errMsg = `Sandbox ${this.id} failed to start with status: ${this.state}, error reason: ${this.errorReason}`
        throw new DaytonaError(errMsg)
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
  @WithInstrumentation()
  public async waitUntilStopped(timeout = 60) {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    const checkInterval = 100 // Wait 100 ms between checks
    const startTime = Date.now()

    // Treat destroyed as stopped to cover ephemeral sandboxes that are automatically deleted after stopping
    while (this.state !== 'stopped' && this.state !== 'destroyed') {
      this.refreshDataSafe()

      // @ts-expect-error this.refreshData() can modify this.state so this check is fine
      if (this.state === 'stopped' || this.state === 'destroyed') {
        return
      }

      if (this.state === 'error') {
        const errMsg = `Sandbox failed to stop with status: ${this.state}, error reason: ${this.errorReason}`
        throw new DaytonaError(errMsg)
      }

      if (timeout !== 0 && Date.now() - startTime > timeout * 1000) {
        throw new DaytonaError('Sandbox failed to become stopped within the timeout period')
      }

      await new Promise((resolve) => setTimeout(resolve, checkInterval))
    }
  }

  /**
   * Refreshes the Sandbox data from the API.
   *
   * @returns {Promise<void>}
   *
   * @example
   * await sandbox.refreshData();
   * console.log(`Sandbox ${sandbox.id}:`);
   * console.log(`State: ${sandbox.state}`);
   * console.log(`Resources: ${sandbox.cpu} CPU, ${sandbox.memory} GiB RAM`);
   */
  @WithInstrumentation()
  public async refreshData(): Promise<void> {
    const response = await this.sandboxApi.getSandbox(this.id)
    this.processSandboxDto(response.data)
  }

  /**
   * Refreshes the sandbox activity to reset the timer for automated lifecycle management actions.
   *
   * This method updates the sandbox's last activity timestamp without changing its state.
   * It is useful for keeping long-running sessions alive while there is still user activity.
   *
   * @returns {Promise<void>}
   *
   * @example
   * // Keep sandbox activity alive
   * await sandbox.refreshActivity();
   */
  public async refreshActivity(): Promise<void> {
    await this.sandboxApi.updateLastActivity(this.id)
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
  @WithInstrumentation()
  public async setAutostopInterval(interval: number): Promise<void> {
    if (!Number.isInteger(interval) || interval < 0) {
      throw new DaytonaError('autoStopInterval must be a non-negative integer')
    }

    await this.sandboxApi.setAutostopInterval(this.id, interval)
    this.autoStopInterval = interval
  }

  /**
   * Set the auto-archive interval for the Sandbox.
   *
   * The Sandbox will automatically archive after being continuously stopped for the specified interval.
   *
   * @param {number} interval - Number of minutes after which a continuously stopped Sandbox will be auto-archived.
   *                           Set to 0 for the maximum interval. Default is 7 days.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - `DaytonaError` - If interval is not a non-negative integer
   *
   * @example
   * // Auto-archive after 1 hour
   * await sandbox.setAutoArchiveInterval(60);
   * // Or use the maximum interval
   * await sandbox.setAutoArchiveInterval(0);
   */
  @WithInstrumentation()
  public async setAutoArchiveInterval(interval: number): Promise<void> {
    if (!Number.isInteger(interval) || interval < 0) {
      throw new DaytonaError('autoArchiveInterval must be a non-negative integer')
    }
    await this.sandboxApi.setAutoArchiveInterval(this.id, interval)
    this.autoArchiveInterval = interval
  }

  /**
   * Set the auto-delete interval for the Sandbox.
   *
   * The Sandbox will automatically delete after being continuously stopped for the specified interval.
   *
   * @param {number} interval - Number of minutes after which a continuously stopped Sandbox will be auto-deleted.
   *                           Set to negative value to disable auto-delete. Set to 0 to delete immediately upon stopping.
   *                           By default, auto-delete is disabled.
   * @returns {Promise<void>}
   *
   * @example
   * // Auto-delete after 1 hour
   * await sandbox.setAutoDeleteInterval(60);
   * // Or delete immediately upon stopping
   * await sandbox.setAutoDeleteInterval(0);
   * // Or disable auto-delete
   * await sandbox.setAutoDeleteInterval(-1);
   */
  @WithInstrumentation()
  public async setAutoDeleteInterval(interval: number): Promise<void> {
    await this.sandboxApi.setAutoDeleteInterval(this.id, interval)
    this.autoDeleteInterval = interval
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
  @WithInstrumentation()
  public async getPreviewLink(port: number): Promise<PortPreviewUrl> {
    return (await this.sandboxApi.getPortPreviewUrl(this.id, port)).data
  }

  /**
   * Retrieves a signed preview url for the sandbox at the specified port.
   *
   * @param {number} port - The port to open the preview link on.
   * @param {number} [expiresInSeconds] - The number of seconds the signed preview url will be valid for. Defaults to 60 seconds.
   * @returns {Promise<SignedPortPreviewUrl>} The response object for the signed preview url.
   */
  public async getSignedPreviewUrl(port: number, expiresInSeconds?: number): Promise<SignedPortPreviewUrl> {
    return (await this.sandboxApi.getSignedPortPreviewUrl(this.id, port, undefined, expiresInSeconds)).data
  }

  /**
   * Expires a signed preview url for the sandbox at the specified port.
   *
   * @param {number} port - The port to expire the signed preview url on.
   * @param {string} token - The token to expire the signed preview url on.
   * @returns {Promise<void>}
   */
  public async expireSignedPreviewUrl(port: number, token: string): Promise<void> {
    await this.sandboxApi.expireSignedPortPreviewUrl(this.id, port, token)
  }

  /**
   * Archives the sandbox, making it inactive and preserving its state. When sandboxes are archived, the entire filesystem
   * state is moved to cost-effective object storage, making it possible to keep sandboxes available for an extended period.
   * The tradeoff between archived and stopped states is that starting an archived sandbox takes more time, depending on its size.
   * Sandbox must be stopped before archiving.
   */
  @WithInstrumentation()
  public async archive(): Promise<void> {
    await this.sandboxApi.archiveSandbox(this.id)
    await this.refreshData()
  }

  /**
   * Resizes the Sandbox resources.
   *
   * Changes the CPU, memory, or disk allocation for the Sandbox. Hot resize (on running
   * sandbox) only allows CPU/memory increases. Disk resize requires a stopped sandbox.
   *
   * @param {Resources} resources - New resource configuration. Only specified fields will be updated.
   *   - cpu: Number of CPU cores (minimum: 1). For hot resize, can only be increased.
   *   - memory: Memory in GiB (minimum: 1). For hot resize, can only be increased.
   *   - disk: Disk space in GiB (can only be increased, requires stopped sandbox).
   * @param {number} [timeout=60] - Timeout in seconds for the resize operation. 0 means no timeout.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - If hot resize constraints are violated, disk resize attempted on running sandbox,
   *   disk size decrease is attempted, no resource changes are specified, or resize operation times out.
   *
   * @example
   * // Increase CPU/memory on running sandbox (hot resize)
   * await sandbox.resize({ cpu: 4, memory: 8 });
   *
   * // Change disk (sandbox must be stopped)
   * await sandbox.stop();
   * await sandbox.resize({ cpu: 2, memory: 4, disk: 30 });
   */
  @WithInstrumentation()
  public async resize(resources: Resources, timeout = 60): Promise<void> {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    const startTime = Date.now()
    const resizeRequest: ResizeSandbox = {
      cpu: resources.cpu,
      memory: resources.memory,
      disk: resources.disk,
    }
    const response = await this.sandboxApi.resizeSandbox(this.id, resizeRequest, this.organizationId, {
      timeout: timeout * 1000,
    })
    this.processSandboxDto(response.data)
    const timeElapsed = Date.now() - startTime
    await this.waitForResizeComplete(timeout ? Math.max(0.001, timeout - timeElapsed / 1000) : timeout)
  }

  /**
   * Waits for the Sandbox resize operation to complete.
   *
   * This method polls the Sandbox status until the state is no longer 'resizing'.
   *
   * @param {number} [timeout=60] - Maximum time to wait in seconds. 0 means no timeout.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - If the sandbox ends up in an error state or resize times out.
   */
  @WithInstrumentation()
  public async waitForResizeComplete(timeout = 60): Promise<void> {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    const checkInterval = 100 // Wait 100 ms between checks
    const startTime = Date.now()

    while (this.state === SandboxState.RESIZING) {
      await this.refreshData()

      // @ts-expect-error this.refreshData() can modify this.state so this check is fine
      if (this.state === SandboxState.ERROR || this.state === SandboxState.BUILD_FAILED) {
        throw new DaytonaError(
          `Sandbox ${this.id} resize failed with state: ${this.state}, error reason: ${this.errorReason}`,
        )
      }

      if (this.state !== SandboxState.RESIZING) {
        return
      }

      if (timeout !== 0 && Date.now() - startTime > timeout * 1000) {
        throw new DaytonaError('Sandbox resize did not complete within the timeout period')
      }

      await new Promise((resolve) => setTimeout(resolve, checkInterval))
    }
  }

  /**
   * Creates an SSH access token for the sandbox.
   *
   * @param {number} expiresInMinutes - The number of minutes the SSH access token will be valid for.
   * @returns {Promise<SshAccessDto>} The SSH access token.
   */
  @WithInstrumentation()
  public async createSshAccess(expiresInMinutes?: number): Promise<SshAccessDto> {
    return (await this.sandboxApi.createSshAccess(this.id, undefined, expiresInMinutes)).data
  }

  /**
   * Revokes an SSH access token for the sandbox.
   *
   * @param {string} token - The token to revoke.
   * @returns {Promise<void>}
   */
  @WithInstrumentation()
  public async revokeSshAccess(token: string): Promise<void> {
    await this.sandboxApi.revokeSshAccess(this.id, undefined, token)
  }

  /**
   * Validates an SSH access token for the sandbox.
   *
   * @param {string} token - The token to validate.
   * @returns {Promise<SshAccessValidationDto>} The SSH access validation result.
   */
  @WithInstrumentation()
  public async validateSshAccess(token: string): Promise<SshAccessValidationDto> {
    return (await this.sandboxApi.validateSshAccess(token)).data
  }

  /**
   * Assigns the API sandbox data to the Sandbox object.
   *
   * @param {SandboxDto} sandboxDto - The API sandbox instance to assign data from
   * @returns {void}
   */
  private processSandboxDto(sandboxDto: SandboxDto) {
    this.id = sandboxDto.id
    this.name = sandboxDto.name
    this.organizationId = sandboxDto.organizationId
    this.snapshot = sandboxDto.snapshot
    this.user = sandboxDto.user
    this.env = sandboxDto.env
    this.labels = sandboxDto.labels
    this.public = sandboxDto.public
    this.target = sandboxDto.target
    this.cpu = sandboxDto.cpu
    this.gpu = sandboxDto.gpu
    this.memory = sandboxDto.memory
    this.disk = sandboxDto.disk
    this.state = sandboxDto.state
    this.errorReason = sandboxDto.errorReason
    this.recoverable = sandboxDto.recoverable
    this.backupState = sandboxDto.backupState
    this.backupCreatedAt = sandboxDto.backupCreatedAt
    this.autoStopInterval = sandboxDto.autoStopInterval
    this.autoArchiveInterval = sandboxDto.autoArchiveInterval
    this.autoDeleteInterval = sandboxDto.autoDeleteInterval
    this.volumes = sandboxDto.volumes
    this.buildInfo = sandboxDto.buildInfo
    this.createdAt = sandboxDto.createdAt
    this.updatedAt = sandboxDto.updatedAt
    this.networkBlockAll = sandboxDto.networkBlockAll
    this.networkAllowList = sandboxDto.networkAllowList
  }

  /**
   * Refreshes the Sandbox data from the API, but does not throw an error if the sandbox has been deleted.
   * Instead, it sets the state to destroyed.
   *
   * @returns {Promise<void>}
   */
  private async refreshDataSafe(): Promise<void> {
    try {
      await this.refreshData()
    } catch (error) {
      if (error instanceof DaytonaNotFoundError) {
        this.state = SandboxState.DESTROYED
      }
    }
  }

  private async ensureToolboxUrl(): Promise<void> {
    if (this.axiosInstance.defaults.baseURL !== TOOLBOX_URL_PLACEHOLDER) {
      return
    }
    this.axiosInstance.defaults.baseURL = await this.getToolboxBaseUrl(this.id, this.target)
    if (!this.axiosInstance.defaults.baseURL.endsWith('/')) {
      this.axiosInstance.defaults.baseURL += '/'
    }
    this.axiosInstance.defaults.baseURL += this.id
    this.clientConfig.basePath = this.axiosInstance.defaults.baseURL
  }
}

export interface PaginatedSandboxes extends Omit<PaginatedSandboxesDto, 'items'> {
  items: Sandbox[]
}
