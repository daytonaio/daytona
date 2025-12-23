/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as pathe from 'pathe'
import { quote, parse as parseShellQuote } from 'shell-quote'
import { DaytonaError } from './errors/DaytonaError'
import { dynamicImport } from './utils/Import'

const SUPPORTED_PYTHON_SERIES = ['3.9', '3.10', '3.11', '3.12', '3.13'] as const
type SupportedPythonSeries = (typeof SUPPORTED_PYTHON_SERIES)[number]
const LATEST_PYTHON_MICRO_VERSIONS = ['3.9.22', '3.10.17', '3.11.12', '3.12.10', '3.13.3']

/**
 * Represents a context file to be added to the image.
 *
 * @interface
 * @property {string} sourcePath - The path to the source file or directory.
 * @property {string} archivePath - The path inside the archive file in object storage.
 */
export interface Context {
  sourcePath: string
  archivePath: string
}

/**
 * Internal state used during image building.
 *
 * @interface
 * @property {string[]} dockerfile - The Dockerfile content as an array of lines.
 * @property {Context[]} contextList - The list of context files to be added to the image.
 */
interface ImgState {
  dockerfile: string[]
  contextList: Context[]
}

/**
 * Represents a built image with its Dockerfile content and context files.
 * This is the immutable result of building an ImageBuilder.
 *
 * @class
 * @property {string} dockerfile - The complete Dockerfile content.
 * @property {Context[]} contextList - The list of context files to be added to the image.
 */
export class Image {
  constructor(
    public readonly dockerfile: string,
    public readonly contextList: Context[],
  ) {}

  /**
   * Creates an ImageBuilder from an existing base image.
   *
   * @param {string} image - The base image to use.
   * @returns {ImageBuilder} The ImageBuilder instance.
   *
   * @example
   * const image = await Image.base('python:3.12-slim-bookworm')
   */
  static base(image: string): ImageBuilder {
    return ImageBuilder.base(image)
  }

  /**
   * Creates an ImageBuilder with a Debian slim base image.
   *
   * @param {string} pythonVersion - The Python version to use.
   * @returns {ImageBuilder} The ImageBuilder instance.
   *
   * @example
   * const image = await Image.debianSlim('3.12')
   */
  static debianSlim(pythonVersion?: SupportedPythonSeries): ImageBuilder {
    return ImageBuilder.debianSlim(pythonVersion)
  }

  /**
   * Creates an ImageBuilder from an existing Dockerfile.
   *
   * @param {string} path - The path to the Dockerfile.
   * @returns {ImageBuilder} The ImageBuilder instance.
   *
   * @example
   * const image = await Image.fromDockerfile('Dockerfile')
   */
  static fromDockerfile(path: string): ImageBuilder {
    return ImageBuilder.fromDockerfile(path)
  }
}

/**
 * Options for the pip install command.
 *
 * @interface
 * @property {string[]} findLinks - The find-links to use for the pip install command.
 * @property {string} indexUrl - The index URL to use for the pip install command.
 * @property {string[]} extraIndexUrls - The extra index URLs to use for the pip install command.
 * @property {boolean} pre - Whether to install pre-release versions.
 * @property {string} extraOptions - The extra options to use for the pip install command. Given string is passed directly to the pip install command.
 */
export interface PipInstallOptions {
  findLinks?: string[]
  indexUrl?: string
  extraIndexUrls?: string[]
  pre?: boolean
  extraOptions?: string
}

/**
 * Options for the pip install command from a pyproject.toml file.
 *
 * @interface
 * @property {string[]} optionalDependencies - The optional dependencies to install.
 *
 * @extends {PipInstallOptions}
 */
export interface PyprojectOptions extends PipInstallOptions {
  optionalDependencies?: string[]
}

/**
 * A builder for creating Docker images for Daytona sandboxes.
 * Do not construct this class directly. Instead use static factory methods on the Image class,
 * such as `Image.base()`, `Image.debianSlim()` or `Image.fromDockerfile()`.
 *
 * ## Execution Model
 * - Operations record transformation steps without executing immediately
 * - Steps are executed sequentially when the builder is awaited
 * - You **MUST** `await` the builder after the final method call to get the built Image
 * - The builder resolves to an immutable Image object
 *
 * @class
 *
 * @example
 * // Correct usage - await after the chain to get Image
 * const image = await Image.debianSlim('3.12')
 *   .pipInstall(['numpy'])
 *   .addLocalFile('file.txt', '/dest')
 *   .env({ FOO: 'bar' })
 *
 * await daytona.snapshot.create({ name: 'my-snapshot', image })
 */
export class ImageBuilder implements PromiseLike<Image> {
  private steps: Array<(s: ImgState) => ImgState | Promise<ImgState>> = []

  private constructor(private initial: ImgState) {}

  /**
   * Builds the final Image by applying all recorded transformation steps.
   *
   * @returns {Promise<Image>} A promise that resolves to the built Image.
   */
  private async build(): Promise<Image> {
    let state = this.initial
    for (const step of this.steps) {
      state = await step(state)
    }
    return new Image(state.dockerfile.join('\n'), state.contextList)
  }

  /**
   * Makes the ImageBuilder thenable, allowing it to be used with `await`.
   *
   * Steps are recorded as they are called, but execution is deferred until the builder is awaited.
   * Awaiting the builder builds the final Image by applying all transformation steps sequentially.
   *
   * **Important**: You MUST await the builder after the method chain to get the built Image
   * before passing it to SDK methods.
   *
   * @param {function} onfulfilled - Called when the Image is resolved.
   * @param {function} onrejected - Called when the build is rejected.
   * @returns {PromiseLike} A promise that resolves to the built Image.
   *
   * @example
   * const image = await Image.debianSlim('3.12')
   *   .pipInstall('numpy')
   *   .addLocalFile('file.txt', '/dest')
   */
  then<TResult1 = Image, TResult2 = never>(
    onfulfilled?: ((value: Image) => TResult1 | PromiseLike<TResult1>) | null | undefined,
    onrejected?: ((reason: any) => TResult2 | PromiseLike<TResult2>) | null | undefined,
  ): PromiseLike<TResult1 | TResult2> {
    return this.build().then(onfulfilled, onrejected)
  }

  /**
   * Adds commands to install packages using pip.
   *
   * @param {string | string[]} packages - The packages to install.
   * @param {Object} options - The options for the pip install command.
   * @param {string[]} options.findLinks - The find-links to use for the pip install command.
   * @returns {this} The ImageBuilder instance.
   *
   * @example
   * const image = await Image.debianSlim('3.12').pipInstall('numpy', { findLinks: ['https://pypi.org/simple'] })
   */
  pipInstall(packages: string | string[], options?: PipInstallOptions): this {
    const pkgs = this.flattenStringArgs('pipInstall', 'packages', packages)
    if (!pkgs.length) return this

    const extraArgs = this.formatPipInstallArgs(options)
    this.steps.push((s) => ({
      ...s,
      dockerfile: [...s.dockerfile, `RUN python -m pip install ${quote(pkgs.sort())}${extraArgs}`],
    }))

    return this
  }

  /**
   * Installs dependencies from a requirements.txt file.
   *
   * @param {string} requirementsTxt - The path to the requirements.txt file.
   * @param {PipInstallOptions} options - The options for the pip install command.
   * @returns {this} The ImageBuilder instance.
   *
   * @example
   * const image = Image.debianSlim('3.12')
   * image.pipInstallFromRequirements('requirements.txt', { findLinks: ['https://pypi.org/simple'] })
   */
  pipInstallFromRequirements(requirementsTxt: string, options?: PipInstallOptions): this {
    this.steps.push(async (s) => {
      const importErrorPrefix = '"pipInstallFromRequirements" is not supported: '
      const expandTilde = await dynamicImport('expand-tilde', importErrorPrefix)
      const fs = await dynamicImport('fs', importErrorPrefix)

      const expandedPath = expandTilde(requirementsTxt)
      if (!fs.existsSync(expandedPath)) {
        throw new Error(`Requirements file ${requirementsTxt} does not exist`)
      }

      const extraArgs = this.formatPipInstallArgs(options)

      return {
        ...s,
        contextList: [...s.contextList, { sourcePath: expandedPath, archivePath: expandedPath }],
        dockerfile: [
          ...s.dockerfile,
          `COPY ${expandedPath} /.requirements.txt`,
          `RUN python -m pip install -r /.requirements.txt${extraArgs}`,
        ],
      }
    })
    return this
  }

  /**
   * Installs dependencies from a pyproject.toml file.
   *
   * @param {string} pyprojectToml - The path to the pyproject.toml file.
   * @param {PyprojectOptions} options - The options for the pip install command.
   * @returns {this} The ImageBuilder instance.
   *
   * @example
   * const image = Image.debianSlim('3.12')
   * image.pipInstallFromPyproject('pyproject.toml', { optionalDependencies: ['dev'] })
   */
  pipInstallFromPyproject(pyprojectToml: string, options?: PyprojectOptions): this {
    this.steps.push(async (s) => {
      const importErrorPrefix = '"pipInstallFromPyproject" is not supported: '
      const expandTilde = await dynamicImport('expand-tilde', importErrorPrefix)
      const toml = await dynamicImport('@iarna/toml', importErrorPrefix)
      const fs = await dynamicImport('fs', importErrorPrefix)

      const tomlData = toml.parse(fs.readFileSync(expandTilde(pyprojectToml), 'utf-8')) as any
      const dependencies: string[] = []

      if (!tomlData || !tomlData.project || !Array.isArray(tomlData.project.dependencies)) {
        const msg =
          'No [project.dependencies] section in pyproject.toml file. ' +
          'See https://packaging.python.org/en/latest/guides/writing-pyproject-toml ' +
          'for further file format guidelines.'
        throw new DaytonaError(msg)
      }

      dependencies.push(...tomlData.project.dependencies)

      if (options?.optionalDependencies && tomlData.project['optional-dependencies']) {
        const optionalGroups = tomlData.project['optional-dependencies'] as Record<string, string[]>
        for (const group of options.optionalDependencies) {
          const deps = optionalGroups[group]
          if (Array.isArray(deps)) {
            dependencies.push(...deps)
          }
        }
      }

      const pkgs = this.flattenStringArgs('pipInstall', 'packages', dependencies)
      if (!pkgs.length) return s

      const extraArgs = this.formatPipInstallArgs(options)
      return {
        ...s,
        dockerfile: [...s.dockerfile, `RUN python -m pip install ${quote(pkgs.sort())}${extraArgs}`],
      }
    })
    return this
  }

  /**
   * Adds a local file to the image.
   *
   * @param {string} localPath - The path to the local file.
   * @param {string} remotePath - The path of the file in the image.
   * @returns {this} The ImageBuilder instance.
   *
   * @example
   * const image = Image
   *  .debianSlim('3.12')
   *  .addLocalFile('requirements.txt', '/home/daytona/requirements.txt')
   */
  addLocalFile(localPath: string, remotePath: string): this {
    this.steps.push(async (s) => {
      const importErrorPrefix = '"addLocalFile" is not supported: '
      const expandTilde = await dynamicImport('expand-tilde', importErrorPrefix)

      let finalRemotePath = remotePath
      if (remotePath.endsWith('/')) {
        finalRemotePath = remotePath + pathe.basename(localPath)
      }

      const expandedPath = expandTilde(localPath)
      return {
        ...s,
        contextList: [...s.contextList, { sourcePath: expandedPath, archivePath: expandedPath }],
        dockerfile: [...s.dockerfile, `COPY ${expandedPath} ${finalRemotePath}`],
      }
    })
    return this
  }

  /**
   * Adds a local directory to the image.
   *
   * @param {string} localPath - The path to the local directory.
   * @param {string} remotePath - The path of the directory in the image.
   * @returns {this} The ImageBuilder instance.
   *
   * @example
   * const image = Image
   *  .debianSlim('3.12')
   *  .addLocalDir('src', '/home/daytona/src')
   */
  addLocalDir(localPath: string, remotePath: string): this {
    this.steps.push(async (s) => {
      const importErrorPrefix = '"addLocalDir" is not supported: '
      const expandTilde = await dynamicImport('expand-tilde', importErrorPrefix)

      const expandedPath = expandTilde(localPath)

      return {
        ...s,
        contextList: [...s.contextList, { sourcePath: expandedPath, archivePath: expandedPath }],
        dockerfile: [...s.dockerfile, `COPY ${expandedPath} ${remotePath}`],
      }
    })
    return this
  }

  /**
   * Runs commands in the image.
   *
   * @param {string | string[]} commands - The commands to run.
   * @returns {this} The ImageBuilder instance.
   *
   * @example
   * const image = Image
   *  .debianSlim('3.12')
   *  .runCommands(
   *    'echo "Hello, world!"',
   *    ['bash', '-c', 'echo Hello, world, again!']
   *  )
   */
  runCommands(...commands: (string | string[])[]): this {
    this.steps.push((s) => {
      const newLines: string[] = []
      for (const command of commands) {
        if (Array.isArray(command)) {
          newLines.push(`RUN ${command.map((c) => `"${c.replace(/"/g, '\\\\\\"').replace(/'/g, "\\'")}"`).join(' ')}`)
        } else {
          newLines.push(`RUN ${command}`)
        }
      }
      return {
        ...s,
        dockerfile: [...s.dockerfile, ...newLines],
      }
    })

    return this
  }

  /**
   * Sets environment variables in the image.
   *
   * @param {Record<string, string>} envVars - The environment variables to set.
   * @returns {this} The ImageBuilder instance.
   *
   * @example
   * const image = Image
   *  .debianSlim('3.12')
   *  .env({ FOO: 'bar' })
   */
  env(envVars: Record<string, string>): this {
    const nonStringKeys = Object.entries(envVars)
      .filter(([, value]) => typeof value !== 'string')
      .map(([key]) => key)

    if (nonStringKeys.length) {
      throw new Error(`Image ENV variables must be strings. Invalid keys: ${nonStringKeys}`)
    }

    this.steps.push((s) => {
      const newLines: string[] = []
      for (const [key, val] of Object.entries(envVars)) {
        newLines.push(`ENV ${key}=${quote([val])}`)
      }
      return {
        ...s,
        dockerfile: [...s.dockerfile, ...newLines],
      }
    })

    return this
  }

  /**
   * Sets the working directory in the image.
   *
   * @param {string} dirPath - The path to the working directory.
   * @returns {this} The ImageBuilder instance.
   *
   * @example
   * const image = Image
   *  .debianSlim('3.12')
   *  .workdir('/home/daytona')
   */
  workdir(dirPath: string): this {
    this.steps.push((s) => ({
      ...s,
      dockerfile: [...s.dockerfile, `WORKDIR ${quote([dirPath])}`],
    }))
    return this
  }

  /**
   * Sets the entrypoint for the image.
   *
   * @param {string[]} entrypointCommands - The commands to set as the entrypoint.
   * @returns {this} The ImageBuilder instance.
   *
   * @example
   * const image = Image
   *  .debianSlim('3.12')
   *  .entrypoint(['/bin/bash'])
   */
  entrypoint(entrypointCommands: string[]): this {
    if (!Array.isArray(entrypointCommands) || !entrypointCommands.every((x) => typeof x === 'string')) {
      throw new Error('entrypoint_commands must be a list of strings')
    }

    this.steps.push((s) => {
      const argsStr = entrypointCommands.map((arg) => `"${arg}"`).join(', ')
      return {
        ...s,
        dockerfile: [...s.dockerfile, `ENTRYPOINT [${argsStr}]`],
      }
    })

    return this
  }

  /**
   * Sets the default command for the image.
   *
   * @param {string[]} cmd - The command to set as the default command.
   * @returns {this} The ImageBuilder instance.
   *
   * @example
   * const image = Image
   *  .debianSlim('3.12')
   *  .cmd(['/bin/bash'])
   */
  cmd(cmd: string[]): this {
    if (!Array.isArray(cmd) || !cmd.every((x) => typeof x === 'string')) {
      throw new Error('Image CMD must be a list of strings')
    }

    this.steps.push((s) => {
      const cmdStr = cmd.map((arg) => `"${arg}"`).join(', ')
      return {
        ...s,
        dockerfile: [...s.dockerfile, `CMD [${cmdStr}]`],
      }
    })

    return this
  }

  /**
   * Extends an image with arbitrary Dockerfile-like commands.
   *
   * @param {string | string[]} dockerfileCommands - The commands to add to the Dockerfile.
   * @param {string} contextDir - The path to the context directory.
   * @returns {this} The ImageBuilder instance.
   *
   * @example
   * const image = Image
   *  .debianSlim('3.12')
   *  .dockerfileCommands(['RUN echo "Hello, world!"'])
   */
  dockerfileCommands(dockerfileCommands: string[], contextDir?: string): this {
    this.steps.push(async (s) => {
      if (contextDir) {
        const importErrorPrefix = '"dockerfileCommands" is not supported: '
        const expandTilde = await dynamicImport('expand-tilde', importErrorPrefix)
        const fs = await dynamicImport('fs', importErrorPrefix)

        const expandedPath = expandTilde(contextDir)
        if (!fs.existsSync(expandedPath) || !fs.statSync(expandedPath).isDirectory()) {
          throw new Error(`Context directory ${contextDir} does not exist`)
        }
      }

      const newContextList: Context[] = []
      for (const [contextPath, originalPath] of await ImageBuilder.extractCopySources(
        dockerfileCommands.join('\n'),
        contextDir || '',
      )) {
        let archiveBasePath = contextPath
        if (contextDir && !originalPath.startsWith(contextDir)) {
          archiveBasePath = contextPath.substring(contextDir.length)
          // Remove leading separators
          // eslint-disable-next-line no-useless-escape
          archiveBasePath = archiveBasePath.replace(/^[\/\\]+/, '')
        }
        newContextList.push({ sourcePath: contextPath, archivePath: archiveBasePath })
      }

      return {
        ...s,
        contextList: [...s.contextList, ...newContextList],
        dockerfile: [...s.dockerfile, ...dockerfileCommands],
      }
    })
    return this
  }

  /**
   * Creates an ImageBuilder from an existing Dockerfile.
   *
   * @param {string} path - The path to the Dockerfile.
   * @returns {ImageBuilder} The ImageBuilder instance.
   *
   * @example
   * const image = await Image.fromDockerfile('Dockerfile')
   */
  static fromDockerfile(path: string): ImageBuilder {
    const img = new ImageBuilder({ dockerfile: [], contextList: [] })
    img.steps.push(async (s) => {
      const importErrorPrefix = '"fromDockerfile" is not supported: '
      const expandTilde = await dynamicImport('expand-tilde', importErrorPrefix)
      const fs = await dynamicImport('fs', importErrorPrefix)

      const expandedPath = pathe.resolve(expandTilde(path))
      if (!fs.existsSync(expandedPath)) {
        throw new Error(`Dockerfile ${path} does not exist`)
      }

      const dockerfileContent = fs.readFileSync(expandedPath, 'utf-8')

      // Remove dockerfile filename from path to get the path prefix
      const pathPrefix = pathe.dirname(expandedPath) + pathe.sep

      const newContextList: Context[] = []
      for (const [contextPath, originalPath] of await ImageBuilder.extractCopySources(dockerfileContent, pathPrefix)) {
        let archiveBasePath = contextPath
        if (!originalPath.startsWith(pathPrefix)) {
          // Remove the path prefix from the context path to get the archive path
          archiveBasePath = contextPath.substring(pathPrefix.length)
          // Remove leading separators
          // eslint-disable-next-line no-useless-escape
          archiveBasePath = archiveBasePath.replace(/^[\/\\]+/, '')
        }
        newContextList.push({ sourcePath: contextPath, archivePath: archiveBasePath })
      }

      return {
        ...s,
        dockerfile: dockerfileContent.split('\n').filter((line) => line.trim()),
        contextList: [...s.contextList, ...newContextList],
      }
    })
    return img
  }

  /**
   * Creates an ImageBuilder from an existing base image.
   *
   * @param {string} image - The base image to use.
   * @returns {ImageBuilder} The ImageBuilder instance.
   *
   * @example
   * const image = await Image.base('python:3.12-slim-bookworm')
   */
  static base(image: string): ImageBuilder {
    return new ImageBuilder({
      dockerfile: [`FROM ${image}`],
      contextList: [],
    })
  }

  /**
   * Creates an ImageBuilder with a Debian slim base image.
   *
   * @param {string} pythonVersion - The Python version to use.
   * @returns {ImageBuilder} The ImageBuilder instance.
   *
   * @example
   * const image = await Image.debianSlim('3.12')
   */
  static debianSlim(pythonVersion?: SupportedPythonSeries): ImageBuilder {
    const version = ImageBuilder.processPythonVersion(pythonVersion)

    const commands = [
      `FROM python:${version}-slim-bookworm`,
      'RUN apt-get update',
      'RUN apt-get install -y gcc gfortran build-essential',
      'RUN pip install --upgrade pip',
      // Set debian front-end to non-interactive to avoid users getting stuck with input prompts.
      "RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections",
    ]

    return new ImageBuilder({
      dockerfile: commands,
      contextList: [],
    })
  }

  /**
   * Formats pip install arguments in a single string.
   *
   * @param {PipInstallOptions} options - The options for the pip install command.
   * @returns {string} The formatted pip install arguments.
   */
  private formatPipInstallArgs(options?: PipInstallOptions): string {
    if (!options) return ''

    let extraArgs = ''

    if (options.findLinks) {
      for (const findLink of options.findLinks) {
        extraArgs += ` --find-links ${quote([findLink])}`
      }
    }

    if (options.indexUrl) {
      extraArgs += ` --index-url ${quote([options.indexUrl])}`
    }

    if (options.extraIndexUrls) {
      for (const extraIndexUrl of options.extraIndexUrls) {
        extraArgs += ` --extra-index-url ${quote([extraIndexUrl])}`
      }
    }

    if (options.pre) {
      extraArgs += ' --pre'
    }

    if (options.extraOptions) {
      extraArgs += ` ${options.extraOptions.trim()}`
    }

    return extraArgs
  }

  /**
   * Flattens a string argument.
   *
   * @param {string} functionName - The name of the function.
   * @param {string} argName - The name of the argument.
   * @param {any} args - The argument to flatten.
   * @returns {string[]} The flattened argument.
   */
  private flattenStringArgs(functionName: string, argName: string, args: any): string[] {
    const result: string[] = []

    const flatten = (arg: any) => {
      if (typeof arg === 'string') {
        result.push(arg)
      } else if (Array.isArray(arg)) {
        for (const item of arg) {
          flatten(item)
        }
      } else {
        throw new Error(`${functionName}: ${argName} must only contain strings`)
      }
    }

    flatten(args)
    return result
  }

  /**
   * Processes the Python version.
   *
   * @param {string} pythonVersion - The Python version to use.
   * @returns {string} The processed Python version.
   */
  private static processPythonVersion(pythonVersion?: SupportedPythonSeries): string {
    if (!pythonVersion) {
      // Default to latest
      pythonVersion = SUPPORTED_PYTHON_SERIES[SUPPORTED_PYTHON_SERIES.length - 1]
    }

    if (!SUPPORTED_PYTHON_SERIES.includes(pythonVersion)) {
      throw new Error(
        `Unsupported Python version: ${pythonVersion}. ` +
          `Daytona supports the following series: ${SUPPORTED_PYTHON_SERIES.join(', ')}`,
      )
    }

    // Map series to latest micro version
    const seriesMap = Object.fromEntries(
      LATEST_PYTHON_MICRO_VERSIONS.map((v) => {
        const [major, minor, micro] = v.split('.')
        return [`${major}.${minor}`, micro]
      }),
    )

    const micro = seriesMap[pythonVersion]
    return `${pythonVersion}.${micro}`
  }

  /**
   * Extracts source files from COPY commands in a Dockerfile.
   *
   * @param {string} dockerfileContent - The content of the Dockerfile.
   * @param {string} pathPrefix - The path prefix to use for the sources.
   * @returns {Array<[string, string]>} The list of the actual file path and its corresponding COPY-command source path.
   */
  private static async extractCopySources(
    dockerfileContent: string,
    pathPrefix = '',
  ): Promise<Array<[string, string]>> {
    const sources: Array<[string, string]> = []
    const lines = dockerfileContent.split('\n')

    for (const line of lines) {
      // Skip empty lines and comments
      if (!line.trim() || line.trim().startsWith('#')) {
        continue
      }

      // Check if the line contains a COPY command
      if (/^\s*COPY\s+(?!.*--from=)/i.test(line)) {
        const importErrorPrefix = '"extractCopySources" is not supported: '
        const fg = await dynamicImport('fast-glob', importErrorPrefix)

        const commandParts = this.parseCopyCommand(line)
        if (commandParts) {
          // Get source paths from the parsed command parts
          for (const source of commandParts.sources) {
            // Handle absolute and relative paths differently
            const fullPathPattern = pathe.isAbsolute(source) ? source : pathe.join(pathPrefix, source)

            const matchingFiles = fg.sync([fullPathPattern], { dot: true })
            if (matchingFiles.length > 0) {
              for (const matchingFile of matchingFiles) {
                sources.push([matchingFile, source])
              }
            } else {
              sources.push([fullPathPattern, source])
            }
          }
        }
      }
    }

    return sources
  }

  /**
   * Parses a COPY command to extract sources and destination.
   *
   * @param {string} line - The line to parse.
   * @returns {Object} The parsed sources and destination.
   */
  private static parseCopyCommand(line: string): { sources: string[]; dest: string } | null {
    // Remove initial "COPY" and strip whitespace
    const parts = line.trim().substring(4).trim()

    // Handle JSON array format: COPY ["src1", "src2", "dest"]
    if (parts.startsWith('[')) {
      try {
        // Parse the JSON-like array format
        const elements = parseShellQuote(parts.replace('[', '').replace(']', '')).filter(
          (x): x is string => typeof x === 'string',
        )

        if (elements.length < 2) {
          return null
        }

        return {
          sources: elements.slice(0, -1),
          dest: elements[elements.length - 1],
        }
      } catch {
        return null
      }
    }

    // Handle regular format with possible flags
    const splitParts = parseShellQuote(parts).filter((x): x is string => typeof x === 'string')

    // Extract flags like --chown, --chmod, --from
    let sourcesStartIdx = 0
    for (let i = 0; i < splitParts.length; i++) {
      const part = splitParts[i]
      if (part.startsWith('--')) {
        // Skip the flag and its value if it has one
        if (!part.includes('=') && i + 1 < splitParts.length && !splitParts[i + 1].startsWith('--')) {
          sourcesStartIdx = i + 2
        } else {
          sourcesStartIdx = i + 1
        }
      } else {
        break
      }
    }

    // After skipping flags, we need at least one source and one destination
    if (splitParts.length - sourcesStartIdx < 2) {
      return null
    }

    return {
      sources: splitParts.slice(sourcesStartIdx, -1),
      dest: splitParts[splitParts.length - 1],
    }
  }
}
