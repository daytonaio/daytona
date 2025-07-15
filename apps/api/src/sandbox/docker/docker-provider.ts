/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import Docker from 'dockerode'
import { Inject, Injectable, OnModuleInit, Logger } from '@nestjs/common'
import { ConfigService } from '@nestjs/config'
import axios from 'axios'
import path from 'path'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'

@Injectable()
export class DockerProvider implements OnModuleInit {
  public docker: Docker

  private readonly logger = new Logger(DockerProvider.name)
  private readonly DAYTONA_BINARY_PATH = path.join(process.cwd(), '.tmp', 'binaries', 'daytona')
  private readonly daytonaBinaryUrl: string
  private readonly TERMINAL_BINARY_PATH = path.join(process.cwd(), '.tmp', 'binaries', 'terminal')
  private readonly terminalBinaryUrl: string

  constructor(
    @Inject(ConfigService)
    private readonly configService: ConfigService,
    @Inject(DockerRegistryService)
    private readonly dockerRegistryService: DockerRegistryService,
  ) {
    if (this.configService.get<string>('DOCKER_SSH_HOST')) {
      process.env.DOCKER_HOST = `ssh://${this.configService.get<string>('DOCKER_SSH_USERNAME')}@${this.configService.get<string>('DOCKER_SSH_HOST')}`
      this.docker = new Docker({})
    } else {
      this.docker = new Docker({ socketPath: '/var/run/docker.sock' })
    }
    this.daytonaBinaryUrl = this.configService.get<string>('DAYTONA_BINARY_URL')
    this.terminalBinaryUrl = this.configService.get<string>('TERMINAL_BINARY_URL')
  }

  async onModuleInit() {
    const binaryPromises = []

    try {
      await Promise.all(binaryPromises)
    } catch (error) {
      this.logger.error('Failed to download binaries during initialization:', error)
      // We don't throw here to allow the application to start even if the downloads fail
    }
  }

  public async startTerminalProcess(container: Docker.Container, port = 22222): Promise<void> {
    try {
      // First check if bash is available
      const execCheckBash = await container.exec({
        Cmd: ['which', 'bash'],
        AttachStdout: true,
        AttachStderr: true,
      })

      const shell = await new Promise<string>((resolve) => {
        execCheckBash.start({}, (err, stream) => {
          if (err) {
            resolve('sh')
            return
          }

          let output = ''
          stream.on('data', (chunk) => {
            output += chunk.toString()
          })

          stream.on('end', () => {
            resolve(output.trim() ? 'bash' : 'sh')
          })
        })
      })

      // Start the terminal process
      const execTerminal = await container.exec({
        Cmd: ['terminal', '-p', port.toString(), '-W', shell],
        AttachStdout: false,
        AttachStderr: false,
        Tty: true,
      })

      await execTerminal.start({
        Detach: true,
      })
    } catch (error) {
      this.logger.error('Error starting terminal process:', error)
      // Don't throw the error to prevent breaking the sandbox creation
    }
  }

  private async startDaytonaAgent(container: Docker.Container): Promise<void> {
    try {
      const execDaytona = await container.exec({
        Cmd: ['daytona', 'agent'],
        //  Cmd: ['python3', '-m', 'http.server', '2280'],
        AttachStdout: true,
        AttachStderr: true,
        Tty: true,
      })

      await execDaytona.start(
        {
          Detach: false,
        },
        (err, stream) => {
          if (err) {
            this.logger.error('Error in Daytona agent stream:', err)
            return
          }

          stream.on('data', (chunk) => {
            this.logger.log('Daytona agent output:', chunk.toString())
          })

          stream.on('error', (err) => {
            this.logger.error('Daytona agent stream error:', err)
          })
        },
      )

      return
    } catch (error) {
      this.logger.error('Error starting Daytona agent process:', error)
      // Don't throw the error to prevent breaking the sandbox creation
    }
  }

  async containerExists(containerId: string): Promise<boolean> {
    try {
      const container = this.docker.getContainer(containerId)
      await container.inspect()
      return true
    } catch (error) {
      return false
    }
  }

  async create(imageName: string, entrypoint?: string[]): Promise<string> {
    // Add this before creating the container
    const isValidArch = await this.validateImageArchitecture(imageName)
    if (!isValidArch) {
      throw new Error(`Image ${imageName} is not compatible with x64 architecture`)
    }

    // Create container with direct path binding
    const container = await this.docker.createContainer({
      //  name: sandbox.id,
      Image: imageName,
      // Remove Volumes configuration since we're using direct binding
      Env: ['DAYTONA_SANDBOX_ID=init-image', 'DAYTONA_SANDBOX_USER=root', `DAYTONA_SANDBOX_SNAPSHOT=${imageName}`],
      Entrypoint: entrypoint,
      platform: 'linux/amd64', // Force AMD64 architecture
      HostConfig: {
        Binds: [
          //  `${dirPath}:${osHome}/project`,  // Direct path binding
          ...(this.daytonaBinaryUrl ? [`${this.DAYTONA_BINARY_PATH}:/usr/local/bin/daytona`] : []),
          ...(this.terminalBinaryUrl ? [`${this.TERMINAL_BINARY_PATH}:/usr/local/bin/terminal`] : []),
        ],
        // StorageOpt: {
        //   size: `${sandbox.volume.quota}G`,
        // },
        //  Runtime: 'sysbox-runc',
        //  Privileged: true,
      },
    })
    await container.start()

    // Start both processes in parallel without waiting
    if (this.daytonaBinaryUrl) {
      this.startDaytonaAgent(container).catch((err) => this.logger.error('Failed to start Daytona agent:', err))
    }

    if (this.terminalBinaryUrl) {
      this.startTerminalProcess(container).catch((err) => this.logger.error('Failed to start terminal process:', err))
    }

    return container.id
  }

  private async deleteRepositoryWithPrefix(
    repository: string,
    prefix: string,
    registry: DockerRegistry,
  ): Promise<void> {
    const registryUrl = this.dockerRegistryService.getRegistryUrl(registry)
    const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')
    const repoPath = `${registry.project}/${prefix}${repository}`

    try {
      // Step 1: List all tags in the repository
      const tagsUrl = `${registryUrl}/v2/${repoPath}/tags/list`

      const tagsResponse = await axios({
        method: 'get',
        url: tagsUrl,
        headers: {
          Authorization: `Basic ${encodedCredentials}`,
        },
        validateStatus: (status) => status < 500,
        timeout: 30000,
      })

      if (tagsResponse.status === 404) {
        return
      }

      if (tagsResponse.status >= 300) {
        this.logger.error(`Error listing tags in repository ${repoPath}: ${tagsResponse.statusText}`)
        throw new Error(`Failed to list tags in repository ${repoPath}: ${tagsResponse.statusText}`)
      }

      const tags = tagsResponse.data.tags || []

      if (tags.length === 0) {
        this.logger.debug(`Repository ${repoPath} has no tags to delete`)
        return
      }

      // Step 2: Delete each tag
      for (const tag of tags) {
        try {
          // Get the digest for this tag
          const manifestUrl = `${registryUrl}/v2/${repoPath}/manifests/${tag}`

          const manifestResponse = await axios({
            method: 'head',
            url: manifestUrl,
            headers: {
              Authorization: `Basic ${encodedCredentials}`,
              Accept: 'application/vnd.docker.distribution.manifest.v2+json',
            },
            validateStatus: (status) => status < 500,
            timeout: 30000,
          })

          if (manifestResponse.status >= 300) {
            this.logger.warn(`Couldn't get manifest for tag ${tag}: ${manifestResponse.statusText}`)
            continue
          }

          const digest = manifestResponse.headers['docker-content-digest']
          if (!digest) {
            this.logger.warn(`Docker content digest not found for tag ${tag}`)
            continue
          }

          // Delete the manifest
          const deleteUrl = `${registryUrl}/v2/${repoPath}/manifests/${digest}`

          const deleteResponse = await axios({
            method: 'delete',
            url: deleteUrl,
            headers: {
              Authorization: `Basic ${encodedCredentials}`,
            },
            validateStatus: (status) => status < 500,
            timeout: 30000,
          })

          if (deleteResponse.status < 300) {
            this.logger.debug(`Deleted tag ${tag} from repository ${repoPath}`)
          } else {
            this.logger.warn(`Failed to delete tag ${tag}: ${deleteResponse.statusText}`)
          }
        } catch (error) {
          this.logger.warn(`Exception when deleting tag ${tag}: ${error.message}`)
          // Continue with other tags
        }
      }

      this.logger.debug(`Repository ${repoPath} cleanup completed`)
    } catch (error) {
      this.logger.error(`Exception when deleting repository ${repoPath}: ${error.message}`)
      throw error
    }
  }

  async deleteSandboxRepository(repository: string, registry: DockerRegistry): Promise<void> {
    try {
      // Delete both backup and snapshot repositories - necessary due to renaming
      await this.deleteRepositoryWithPrefix(repository, 'backup-', registry)
      await this.deleteRepositoryWithPrefix(repository, 'snapshot-', registry)
    } catch (error) {
      this.logger.error(`Failed to delete repositories for ${repository}: ${error.message}`)
      throw error
    }
  }

  async deleteBackupImageFromRegistry(imageName: string, registry: DockerRegistry): Promise<void> {
    // Extract tag
    const lastColonIndex = imageName.lastIndexOf(':')
    const fullPath = imageName.substring(0, lastColonIndex)
    const tag = imageName.substring(lastColonIndex + 1)

    const registryUrl = this.dockerRegistryService.getRegistryUrl(registry)

    // Remove registry prefix if present in the image name
    let projectAndRepo = fullPath
    if (fullPath.startsWith(registryUrl)) {
      projectAndRepo = fullPath.substring(registryUrl.length + 1) // +1 for the slash
    }

    // For Harbor format like: harbor.host/bbox-stage/backup-sandbox-75148d5a
    const parts = projectAndRepo.split('/')

    // Construct repository path (everything after the registry host)
    const repoPath = parts.slice(1).join('/')

    // First, get the digest for the tag using the manifests endpoint
    const manifestUrl = `${registryUrl}/v2/${repoPath}/manifests/${tag}`
    const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')

    try {
      // Get the digest from the headers
      const manifestResponse = await axios({
        method: 'head', // Using HEAD request to only fetch headers
        url: manifestUrl,
        headers: {
          Authorization: `Basic ${encodedCredentials}`,
          Accept: 'application/vnd.docker.distribution.manifest.v2+json',
        },
        validateStatus: (status) => status < 500,
        timeout: 30000,
      })

      if (manifestResponse.status >= 300) {
        this.logger.error(`Error getting manifest for image ${imageName}: ${manifestResponse.statusText}`)
        throw new Error(`Failed to get manifest for image ${imageName}: ${manifestResponse.statusText}`)
      }

      // Extract the digest from headers
      const digest = manifestResponse.headers['docker-content-digest']
      if (!digest) {
        throw new Error(`Docker content digest not found for image ${imageName}`)
      }

      // Now delete the image using the digest
      const deleteUrl = `${registryUrl}/v2/${repoPath}/manifests/${digest}`

      const deleteResponse = await axios({
        method: 'delete',
        url: deleteUrl,
        headers: {
          Authorization: `Basic ${encodedCredentials}`,
        },
        validateStatus: (status) => status < 500,
        timeout: 30000,
      })

      if (deleteResponse.status < 300) {
        this.logger.debug(`Image ${imageName} removed from the registry`)
        return
      }

      this.logger.error(`Error removing image ${imageName} from registry: ${deleteResponse.statusText}`)
      throw new Error(`Failed to remove image ${imageName} from registry: ${deleteResponse.statusText}`)
    } catch (error) {
      this.logger.error(`Exception when deleting image ${imageName}: ${error.message}`)
      throw error
    }
  }

  async remove(containerId: string): Promise<void> {
    try {
      const container = this.docker.getContainer(containerId)
      await container.remove({ force: true })
    } catch (error) {
      if (error.statusCode === 404) {
        return
      }
      this.logger.error('Error removing Docker container:', error)
      throw error // Rethrow to let sandbox service handle the error state
    }
  }

  async getContainerIPAddress(containerId: string): Promise<string> {
    const container = this.docker.getContainer(containerId)
    const data = await container.inspect()
    return data.NetworkSettings.IPAddress
  }

  async getImageEntrypoint(image: string): Promise<undefined | string | string[]> {
    const dockerImage = await this.docker.getImage(image).inspect()
    return dockerImage.Config.Entrypoint
  }

  async imageExists(image: string, includeLatest = false): Promise<boolean> {
    image = image.replace('docker.io/', '')
    if (image.endsWith(':latest') && !includeLatest) {
      return false
    }
    const images = await this.docker.listImages({})
    const imageExists = images.some((imageInfo) => imageInfo.RepoTags && imageInfo.RepoTags.includes(image))
    return imageExists
  }

  async isRunning(containerId: string): Promise<boolean> {
    if (!containerId) {
      return false
    }
    try {
      const container = this.docker.getContainer(containerId)
      const data = await container.inspect()
      return data.State.Running
    } catch (error) {
      if (error.statusCode === 404) {
        return false
      }
      this.logger.error('Error checking Docker container state:', error)
      return false // Return false instead of throwing
    }
  }

  async isDestroyed(containerId: string): Promise<boolean> {
    try {
      const container = this.docker.getContainer(containerId)
      await container.inspect()
      return false
    } catch (error) {
      return true
    }
  }

  async validateImageArchitecture(image: string): Promise<boolean> {
    try {
      const imageUnified = image.replace('docker.io/', '')

      const dockerImage = await this.docker.getImage(imageUnified).inspect()

      // Check the architecture from the image metadata
      const architecture = dockerImage.Architecture

      // Valid x64 architectures
      const x64Architectures = ['amd64', 'x86_64']

      // Check if the architecture matches x64
      const isX64 = x64Architectures.includes(architecture.toLowerCase())

      if (!isX64) {
        this.logger.warn(`Image ${image} architecture (${architecture}) is not x64 compatible`)
        return false
      }

      return true
    } catch (error) {
      this.logger.error(`Error validating architecture for image ${image}:`, error)
      throw new Error(`Failed to validate image architecture: ${error.message}`)
    }
  }

  /**
   * Checks if an image exists in the specified registry without pulling it
   */
  async checkImageExistsInRegistry(imageName: string, registry: DockerRegistry): Promise<boolean> {
    try {
      // extract tag
      const lastColonIndex = imageName.lastIndexOf(':')
      const fullPath = imageName.substring(0, lastColonIndex)
      const tag = imageName.substring(lastColonIndex + 1)

      const registryUrl = this.dockerRegistryService.getRegistryUrl(registry)

      // Remove registry prefix if present in the image name
      let projectAndRepo = fullPath
      if (fullPath.startsWith(registryUrl)) {
        projectAndRepo = fullPath.substring(registryUrl.length + 1) // +1 for the slash
      }

      // For Harbor format like: harbor.host/bbox-stage/backup-sandbox-75148d5a
      const parts = projectAndRepo.split('/')

      const apiUrl = `${registryUrl}/v2/${parts[1]}/${parts[2]}/manifests/${tag}`
      const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')

      const response = await axios({
        method: 'get',
        url: apiUrl,
        headers: {
          Authorization: `Basic ${encodedCredentials}`,
        },
        validateStatus: (status) => status < 500,
        timeout: 30000,
      })

      if (response.status === 200) {
        this.logger.debug(`Image ${imageName} exists in registry`)
        return true
      }

      this.logger.debug(`Image ${imageName} does not exist in registry (status: ${response.status})`)
      return false
    } catch (error) {
      this.logger.error(`Error checking if image ${imageName} exists in registry: ${error.message}`)
      return false
    }
  }

  private async retryWithExponentialBackoff<T>(
    operation: () => Promise<T>,
    maxAttempts = 3,
    initialDelayMs = 1000,
  ): Promise<T> {
    let attempt = 1
    let delay = initialDelayMs

    while (attempt <= maxAttempts) {
      try {
        return await operation()
      } catch (error) {
        if (attempt === maxAttempts) {
          throw error
        }

        if (error.fatal) {
          throw error.err
        }

        this.logger.warn(`Attempt ${attempt} failed, retrying in ${delay}ms...`, error)
        await new Promise((resolve) => setTimeout(resolve, delay))

        attempt++
        delay *= 2 // Exponential backoff
      }
    }

    throw new Error('Should not reach here')
  }

  async pullImage(image: string, registry?: { url: string; username: string; password: string }): Promise<void> {
    await this.retryWithExponentialBackoff(async () => {
      const options: any = {
        platform: 'linux/amd64',
      }

      if (registry) {
        options.authconfig = {
          username: registry.username,
          password: registry.password,
          serveraddress: registry.url,
          auth: '',
        }
      }

      try {
        const stream = await this.docker.pull(image, options)
        const err = await new Promise<Error | null>((resolve) => this.docker.modem.followProgress(stream, resolve))
        if (err) {
          throw err
        }
      } catch (err) {
        if (err.statusCode === 404) {
          let returnErr = err
          if (err.message?.includes('pull access denied') || err.message?.includes('no basic auth credentials')) {
            returnErr = new Error('Repository does not exist or may require container registry login credentials.')
          }
          throw {
            fatal: true,
            err: returnErr,
          }
        } else {
          throw err
        }
      }
    })

    // Validate architecture after pulling
    const isValidArch = await this.validateImageArchitecture(image)
    if (!isValidArch) {
      throw new Error(`Image ${image} is not compatible with x64 architecture`)
    }
  }

  async start(containerId: string): Promise<void> {
    try {
      const container = this.docker.getContainer(containerId)
      await container.start()

      // Start both processes in parallel without waiting
      if (this.daytonaBinaryUrl) {
        this.startDaytonaAgent(container).catch((err) => this.logger.error('Failed to start Daytona agent:', err))
      }

      if (this.terminalBinaryUrl) {
        this.startTerminalProcess(container).catch((err) => this.logger.error('Failed to start terminal process:', err))
      }
    } catch (error) {
      this.logger.error('Error starting Docker container:', error)
      throw error // Rethrow or handle as needed
    }
  }

  async stop(containerId: string): Promise<void> {
    try {
      const container = this.docker.getContainer(containerId)
      await container.stop()
    } catch (error) {
      this.logger.error('Error stopping Docker container:', error)
      throw error // Rethrow or handle as needed
    }
  }

  async removeImage(image: string): Promise<void> {
    try {
      await this.docker.getImage(image).remove()
    } catch (error) {
      this.logger.error('Error removing image:', error)
      throw error
    }
  }

  async getImageInfo(imageName: string): Promise<{ sizeGB: number; entrypoint?: string | string[] }> {
    try {
      const image = await this.docker.getImage(imageName).inspect()
      // Size is returned in bytes, convert to GB
      return {
        sizeGB: image.Size / (1024 * 1024 * 1024),
        entrypoint: image.Config.Entrypoint,
      }
    } catch (error) {
      this.logger.error(`Error getting size for image ${imageName}:`, error)
      throw new Error(`Failed to get image size: ${error.message}`)
    }
  }

  async pushImage(image: string, registry: { url: string; username: string; password: string }): Promise<void> {
    await this.retryWithExponentialBackoff(async () => {
      return new Promise((resolve, reject) => {
        const options: any = {
          authconfig: {
            username: registry.username,
            password: registry.password,
            serveraddress: registry.url,
            auth: '',
          },
        }

        this.docker.getImage(image).push(options, (err, stream) => {
          if (err) {
            this.logger.error('Error initiating Docker push:', err)
            reject(err)
            return
          }

          let errorEvent: Error | null = null
          let done = false

          this.docker.modem.followProgress(
            stream,
            (err: Error | null, output: any[]) => {
              if (done) {
                return
              }
              done = true

              if (err) {
                this.logger.error('Error following Docker push progress:', err)
                reject(err)
                return
              }
              if (errorEvent) {
                reject(errorEvent)
                return
              }
              resolve(output)
            },
            (event: any) => {
              // Optional progress callback
              if (event.error) {
                errorEvent = event.error
                this.logger.error('Push progress error:', event.error)
              }
            },
          )
        })
      })
    })
  }

  async tagImage(sourceImage: string, targetImage: string): Promise<void> {
    try {
      const lastColonIndex = targetImage.lastIndexOf(':')
      const repo = targetImage.substring(0, lastColonIndex)
      const tag = targetImage.substring(lastColonIndex + 1)

      if (!repo || !tag) {
        throw new Error('Invalid target image format')
      }

      const image = this.docker.getImage(sourceImage)
      await image.tag({
        repo,
        tag,
      })
    } catch (error) {
      this.logger.error(`Error tagging image ${sourceImage} as ${targetImage}:`, error)
      throw new Error(`Failed to tag image: ${error.message}`)
    }
  }

  async imagePrune(): Promise<void> {
    try {
      await this.docker.pruneImages({
        filters: {
          dangling: { true: true },
        },
      })
    } catch (error) {
      if (error.statusCode === 409) {
        //  if prune is already in progress, just return
        return
      } else {
        throw error
      }
    }
  }
}
