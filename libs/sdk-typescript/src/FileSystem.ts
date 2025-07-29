/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { FileInfo, Match, ReplaceRequest, ReplaceResult, SearchFilesResponse, ToolboxApi } from '@daytonaio/api-client'
import { prefixRelativePath } from './utils/Path'
import { dynamicImport } from './utils/Import'
import { Buffer } from 'buffer'
import { RUNTIME, Runtime } from './utils/Runtime'

/**
 * Parameters for setting file permissions in the Sandbox.
 *
 * @interface
 * @property {string} [mode] - File mode/permissions in octal format (e.g. "644")
 * @property {string} [owner] - User owner of the file
 * @property {string} [group] - Group owner of the file
 *
 * @example
 * const permissions: FilePermissionsParams = {
 *   mode: '644',
 *   owner: 'daytona',
 *   group: 'users'
 * };
 */
export type FilePermissionsParams = {
  /** Group owner of the file */
  group?: string
  /** File mode/permissions in octal format (e.g. "644") */
  mode?: string
  /** User owner of the file */
  owner?: string
}

/**
 * Represents a file to be uploaded to the Sandbox.
 *
 * @interface
 * @property {string | Buffer} source - File to upload. If a Buffer, it is interpreted as the file content which is loaded into memory.
 * Make sure it fits into memory, otherwise use the local file path which content will be streamed to the Sandbox.
 * @property {string} destination - Absolute destination path in the Sandbox. Relative paths are resolved based on the user's
 * root directory.
 */
export interface FileUpload {
  source: string | Buffer
  destination: string
}

/**
 * Provides file system operations within a Sandbox.
 *
 * @class
 */
export class FileSystem {
  constructor(
    private readonly sandboxId: string,
    private readonly toolboxApi: ToolboxApi,
    private readonly getRootDir: () => Promise<string>,
  ) {}

  /**
   * Create a new directory in the Sandbox with specified permissions.
   *
   * @param {string} path - Path where the directory should be created. Relative paths are resolved based on the user's
   * root directory.
   * @param {string} mode - Directory permissions in octal format (e.g. "755")
   * @returns {Promise<void>}
   *
   * @example
   * // Create a directory with standard permissions
   * await fs.createFolder('app/data', '755');
   */
  public async createFolder(path: string, mode: string): Promise<void> {
    const response = await this.toolboxApi.createFolder(
      this.sandboxId,
      prefixRelativePath(await this.getRootDir(), path),
      mode,
    )
    return response.data
  }

  /**
   * Deletes a file or directory from the Sandbox.
   *
   * @param {string} path - Path to the file or directory to delete. Relative paths are resolved based on the user's
   * root directory.
   * @returns {Promise<void>}
   *
   * @example
   * // Delete a file
   * await fs.deleteFile('app/temp.log');
   */
  public async deleteFile(path: string): Promise<void> {
    const response = await this.toolboxApi.deleteFile(this.sandboxId, prefixRelativePath(await this.getRootDir(), path))
    return response.data
  }

  /**
   * Downloads a file from the Sandbox. This method loads the entire file into memory, so it is not recommended
   * for downloading large files.
   *
   * @param {string} remotePath - Path to the file to download. Relative paths are resolved based on the user's
   * root directory.
   * @param {number} [timeout] - Timeout for the download operation in seconds. 0 means no timeout.
   * Default is 30 minutes.
   * @returns {Promise<Buffer>} The file contents as a Buffer.
   *
   * @example
   * // Download and process a file
   * const fileBuffer = await fs.downloadFile('tmp/data.json');
   * console.log('File content:', fileBuffer.toString());
   */
  public async downloadFile(remotePath: string, timeout?: number): Promise<Buffer>
  /**
   * Downloads a file from the Sandbox and saves it to a local file. This method uses streaming to download the file,
   * so it is recommended for downloading larger files.
   *
   * @param {string} remotePath - Path to the file to download in the Sandbox. Relative paths are resolved based on the user's
   * root directory.
   * @param {string} localPath - Path to save the downloaded file.
   * @param {number} [timeout] - Timeout for the download operation in seconds. 0 means no timeout.
   * Default is 30 minutes.
   * @returns {Promise<void>}
   *
   * @example
   * // Download and save a file
   * await fs.downloadFile('tmp/data.json', 'local_file.json');
   */
  public async downloadFile(remotePath: string, localPath: string, timeout?: number): Promise<void>
  public async downloadFile(src: string, dst?: string | number, timeout: number = 30 * 60): Promise<Buffer | void> {
    const remotePath = prefixRelativePath(await this.getRootDir(), src)

    if (typeof dst !== 'string') {
      timeout = dst as number
      const { data } = await this.toolboxApi.downloadFile(this.sandboxId, remotePath, undefined, {
        responseType: 'arraybuffer',
        timeout: timeout * 1000,
      })

      if (Buffer.isBuffer(data)) {
        return data
      }

      if (data instanceof ArrayBuffer) {
        return Buffer.from(data)
      }

      return Buffer.from(await data.arrayBuffer())
    }

    const fs = await dynamicImport('fs', 'Downloading file to local file is not supported: ')

    const response = await this.toolboxApi.downloadFile(this.sandboxId, remotePath, undefined, {
      responseType: 'stream',
      timeout: timeout * 1000,
    })
    const writer = fs.createWriteStream(dst)
    ;(response.data as any).pipe(writer)
    await new Promise<void>((resolve, reject) => {
      writer.on('finish', () => resolve())
      writer.on('error', (err) => reject(err))
    })
  }

  /**
   * Searches for text patterns within files in the Sandbox.
   *
   * @param {string} path - Directory to search in. Relative paths are resolved based on the user's
   * root directory.
   * @param {string} pattern - Search pattern
   * @returns {Promise<Array<Match>>} Array of matches with file and line information
   *
   * @example
   * // Find all TODO comments in TypeScript files
   * const matches = await fs.findFiles('app/src', 'TODO:');
   * matches.forEach(match => {
   *   console.log(`${match.file}:${match.line}: ${match.content}`);
   * });
   */
  public async findFiles(path: string, pattern: string): Promise<Array<Match>> {
    const response = await this.toolboxApi.findInFiles(
      this.sandboxId,
      prefixRelativePath(await this.getRootDir(), path),
      pattern,
    )
    return response.data
  }

  /**
   * Retrieves detailed information about a file or directory.
   *
   * @param {string} path - Path to the file or directory. Relative paths are resolved based on the user's
   * root directory.
   * @returns {Promise<FileInfo>} Detailed file information including size, permissions, modification time
   *
   * @example
   * // Get file details
   * const info = await fs.getFileDetails('app/config.json');
   * console.log(`Size: ${info.size}, Modified: ${info.modTime}`);
   */
  public async getFileDetails(path: string): Promise<FileInfo> {
    const response = await this.toolboxApi.getFileInfo(
      this.sandboxId,
      prefixRelativePath(await this.getRootDir(), path),
    )
    return response.data
  }

  /**
   * Lists contents of a directory in the Sandbox.
   *
   * @param {string} path - Directory path to list. Relative paths are resolved based on the user's
   * root directory.
   * @returns {Promise<FileInfo[]>} Array of file and directory information
   *
   * @example
   * // List directory contents
   * const files = await fs.listFiles('app/src');
   * files.forEach(file => {
   *   console.log(`${file.name} (${file.size} bytes)`);
   * });
   */
  public async listFiles(path: string): Promise<FileInfo[]> {
    const response = await this.toolboxApi.listFiles(
      this.sandboxId,
      undefined,
      prefixRelativePath(await this.getRootDir(), path),
    )
    return response.data
  }

  /**
   * Moves or renames a file or directory.
   *
   * @param {string} source - Source path. Relative paths are resolved based on the user's
   * root directory.
   * @param {string} destination - Destination path. Relative paths are resolved based on the user's
   * root directory.
   * @returns {Promise<void>}
   *
   * @example
   * // Move a file to a new location
   * await fs.moveFiles('app/temp/data.json', 'app/data/data.json');
   */
  public async moveFiles(source: string, destination: string): Promise<void> {
    const response = await this.toolboxApi.moveFile(
      this.sandboxId,
      prefixRelativePath(await this.getRootDir(), source),
      prefixRelativePath(await this.getRootDir(), destination),
    )
    return response.data
  }

  /**
   * Replaces text content in multiple files.
   *
   * @param {string[]} files - Array of file paths to process. Relative paths are resolved based on the user's
   * @param {string} pattern - Pattern to replace
   * @param {string} newValue - Replacement text
   * @returns {Promise<Array<ReplaceResult>>} Results of the replace operation for each file
   *
   * @example
   * // Update version number across multiple files
   * const results = await fs.replaceInFiles(
   *   ['app/package.json', 'app/version.ts'],
   *   '"version": "1.0.0"',
   *   '"version": "1.1.0"'
   * );
   */
  public async replaceInFiles(files: string[], pattern: string, newValue: string): Promise<Array<ReplaceResult>> {
    for (let i = 0; i < files.length; i++) {
      files[i] = prefixRelativePath(await this.getRootDir(), files[i])
    }

    const replaceRequest: ReplaceRequest = {
      files,
      newValue,
      pattern,
    }

    const response = await this.toolboxApi.replaceInFiles(this.sandboxId, replaceRequest)
    return response.data
  }

  /**
   * Searches for files and directories by name pattern in the Sandbox.
   *
   * @param {string} path - Directory to search in. Relative paths are resolved based on the user's
   * @param {string} pattern - File name pattern (supports globs)
   * @returns {Promise<SearchFilesResponse>} Search results with matching files
   *
   * @example
   * // Find all TypeScript files
   * const result = await fs.searchFiles('app', '*.ts');
   * result.files.forEach(file => console.log(file));
   */
  public async searchFiles(path: string, pattern: string): Promise<SearchFilesResponse> {
    const response = await this.toolboxApi.searchFiles(
      this.sandboxId,
      prefixRelativePath(await this.getRootDir(), path),
      pattern,
    )
    return response.data
  }

  /**
   * Sets permissions and ownership for a file or directory.
   *
   * @param {string} path - Path to the file or directory. Relative paths are resolved based on the user's
   * root directory.
   * @param {FilePermissionsParams} permissions - Permission settings
   * @returns {Promise<void>}
   *
   * @example
   * // Set file permissions and ownership
   * await fs.setFilePermissions('app/script.sh', {
   *   owner: 'daytona',
   *   group: 'users',
   *   mode: '755'  // Execute permission for shell script
   * });
   */
  public async setFilePermissions(path: string, permissions: FilePermissionsParams): Promise<void> {
    const response = await this.toolboxApi.setFilePermissions(
      this.sandboxId,
      prefixRelativePath(await this.getRootDir(), path),
      undefined,
      permissions.owner!,
      permissions.group!,
      permissions.mode!,
    )
    return response.data
  }

  /**
   * Uploads a file to the Sandbox. This method loads the entire file into memory, so it is not recommended
   * for uploading large files.
   *
   * @param {Buffer} file - Buffer of the file to upload.
   * @param {string} remotePath - Destination path in the Sandbox. Relative paths are resolved based on the user's
   * root directory.
   * @param {number} [timeout] - Timeout for the upload operation in seconds. 0 means no timeout.
   * Default is 30 minutes.
   * @returns {Promise<void>}
   *
   * @example
   * // Upload a configuration file
   * await fs.uploadFile(Buffer.from('{"setting": "value"}'), 'tmp/config.json');
   */
  public async uploadFile(file: Buffer, remotePath: string, timeout?: number): Promise<void>
  /**
   * Uploads a file from the local file system to the Sandbox. This method uses streaming to upload the file,
   * so it is recommended for uploading larger files.
   *
   * @param {string} localPath - Path to the local file to upload.
   * @param {string} remotePath - Destination path in the Sandbox. Relative paths are resolved based on the user's
   * root directory.
   * @param {number} [timeout] - Timeout for the upload operation in seconds. 0 means no timeout.
   * Default is 30 minutes.
   * @returns {Promise<void>}
   *
   * @example
   * // Upload a local file
   * await fs.uploadFile('local_file.txt', 'tmp/file.txt');
   */
  public async uploadFile(localPath: string, remotePath: string, timeout?: number): Promise<void>
  public async uploadFile(src: string | Buffer, dst: string, timeout: number = 30 * 60): Promise<void> {
    await this.uploadFiles([{ source: src, destination: dst }], timeout)
  }

  /**
   * Uploads multiple files to the Sandbox. If files already exist at the destination paths,
   * they will be overwritten.
   *
   * @param {FileUpload[]} files - Array of files to upload.
   * @param {number} [timeout] - Timeout for the upload operation in seconds. 0 means no timeout.
   * Default is 30 minutes.
   * @returns {Promise<void>}
   *
   * @example
   * // Upload multiple text files
   * const files = [
   *   {
   *     source: Buffer.from('Content of file 1'),
   *     destination: '/tmp/file1.txt'
   *   },
   *   {
   *     source: 'app/data/file2.txt',
   *     destination: '/tmp/file2.txt'
   *   },
   *   {
   *     source: Buffer.from('{"key": "value"}'),
   *     destination: '/tmp/config.json'
   *   }
   * ];
   * await fs.uploadFiles(files);
   */
  public async uploadFiles(files: FileUpload[], timeout: number = 30 * 60): Promise<void> {
    // Use native FormData in Deno
    const FormDataClass =
      RUNTIME === Runtime.DENO
        ? FormData
        : (await dynamicImport('form-data', 'Uploading files is not supported: ')).default
    const form = new FormDataClass()
    const rootDir = await this.getRootDir()

    for (const [i, { source, destination }] of files.entries()) {
      const dst = prefixRelativePath(rootDir, destination)
      form.append(`files[${i}].path`, dst)
      const payload = await this.makeFilePayload(source)
      // the third arg sets filename in Content-Disposition
      form.append(`files[${i}].file`, payload as any, dst)
    }

    await this.toolboxApi.uploadFiles(this.sandboxId, undefined, {
      data: form,
      maxRedirects: 0,
      timeout: timeout * 1000,
    })
  }

  private async makeFilePayload(source: Uint8Array | string) {
    // 1) file‐path
    if (typeof source === 'string') {
      const fs = await dynamicImport('fs', 'Uploading file from local file system is not supported: ')
      return fs.createReadStream(source)
    }

    // 2) browser → Blob
    if (RUNTIME === Runtime.BROWSER) {
      return new Blob([source], { type: 'application/octet-stream' })
    }

    // 3) Node (or other server runtimes) → stream.Readable
    const stream = await dynamicImport('stream', 'Uploading file is not supported: ')
    return stream.Readable.from(source)
  }
}
