import { createTool } from '@mastra/core/tools'
import z from 'zod'
import { getDaytonaClient, getSandboxById, createFileUploadFormat, normalizeSandboxPath } from './utils'
import { Sandbox, CreateSandboxBaseParams, CodeRunParams, DaytonaNotFoundError, CodeLanguage } from '@daytonaio/sdk'

export const createSandbox = createTool({
  id: 'createSandbox',
  description: 'Create a sandbox',
  inputSchema: z.object({
    name: z.string().optional().describe('Custom sandbox name'),
    labels: z.record(z.string()).optional().describe('Custom sandbox labels'),
    language: z
      .nativeEnum(CodeLanguage)
      .default(CodeLanguage.PYTHON)
      .describe('Language used for code execution. If not provided, default python context is used'),
    envVars: z.record(z.string()).optional().describe(`
      Custom environment variables for the sandbox.
      Used when executing commands and code in the sandbox.
      Can be overridden with the \`envs\` argument when executing commands or code.
    `),
  }),
  outputSchema: z
    .object({
      sandboxId: z.string(),
    })
    .or(
      z.object({
        error: z.string(),
      }),
    ),
  execute: async ({ name, labels, language, envVars }) => {
    const daytona = getDaytonaClient()
    try {
      const sandboxParams: CreateSandboxBaseParams = {
        name,
        envVars,
        labels,
        language,
      }
      const sandbox: Sandbox = await daytona.create(sandboxParams)

      return {
        sandboxId: sandbox.id,
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})

export const runCode = createTool({
  id: 'runCode',
  description: 'Run code in a sandbox',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to run the code'),
    code: z.string().describe('The code to run in the sandbox'),
    argv: z.array(z.string()).optional().describe('Command line arguments to pass to the code.'),
    envs: z.record(z.string()).optional().describe('Custom environment variables for code execution.'),
    timeoutSeconds: z.number().optional().describe(`
          Maximum time in seconds to wait for execution to complete
      `),
  }),
  outputSchema: z
    .object({
      exitCode: z.number().describe('The exit code from the code execution'),
      stdout: z.string().optional().describe('The standard output from the code execution'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed execution'),
      }),
    ),
  execute: async ({ sandboxId, code, argv, envs, timeoutSeconds }, context) => {
    try {
      const sandbox = await getSandboxById(sandboxId)

      const codeRunParams = new CodeRunParams()
      codeRunParams.argv = argv ?? []
      codeRunParams.env = envs ?? {}

      const execution = await sandbox.process.codeRun(code, codeRunParams, timeoutSeconds)

      return {
        exitCode: execution.exitCode,
        stdout: execution.result,
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})

export const readFile = createTool({
  id: 'readFile',
  description: 'Read a file from the sandbox',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to read the file from'),
    path: z.string().describe('The path to the file to read'),
  }),
  outputSchema: z
    .object({
      content: z.string().describe('The content of the file'),
      path: z.string().describe('The path of the file that was read'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed file read'),
      }),
    ),
  execute: async ({ sandboxId, path }, context) => {
    try {
      const sandbox = await getSandboxById(sandboxId)
      const normalizedPath = normalizeSandboxPath(path)

      const fileBuffer = await sandbox.fs.downloadFile(normalizedPath)
      const fileContent = fileBuffer.toString('utf-8')

      return {
        content: fileContent,
        path: normalizedPath,
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})

export const writeFile = createTool({
  id: 'writeFile',
  description: 'Write a single file to the sandbox',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to write the file to'),
    path: z.string().describe('The path where the file should be written'),
    content: z.string().describe('The content to write to the file'),
  }),
  outputSchema: z
    .object({
      success: z.boolean().describe('Whether the file was written successfully'),
      path: z.string().describe('The path where the file was written'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed file write'),
      }),
    ),
  execute: async ({ sandboxId, path, content }, context) => {
    try {
      const sandbox = await getSandboxById(sandboxId)
      const normalizedPath = normalizeSandboxPath(path)

      const fileToUpload = [createFileUploadFormat(content, normalizedPath)]
      await sandbox.fs.uploadFiles(fileToUpload)

      return {
        success: true,
        path: normalizedPath,
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})

export const writeFiles = createTool({
  id: 'writeFiles',
  description: 'Write multiple files to the sandbox',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to write the files to'),
    files: z
      .array(
        z.object({
          path: z.string().describe('The path where the file should be written'),
          data: z.string().describe('The content to write to the file'),
        }),
      )
      .describe('Array of files to write, each with path and data'),
  }),
  outputSchema: z
    .object({
      success: z.boolean().describe('Whether all files were written successfully'),
      filesWritten: z.array(z.string()).describe('Array of file paths that were written'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed files write'),
      }),
    ),
  execute: async ({ sandboxId, files }, context) => {
    try {
      const sandbox = await getSandboxById(sandboxId)
      files = files.map((file) => ({
        ...file,
        path: normalizeSandboxPath(file.path),
      }))

      await sandbox.fs.uploadFiles(files.map((file) => createFileUploadFormat(file.data, file.path)))

      return {
        success: true,
        filesWritten: files.map((file) => file.path),
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})

export const listFiles = createTool({
  id: 'listFiles',
  description: 'List files and directories in a path within the sandbox',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to list files from'),
    path: z.string().default('/').describe('The directory path to list files from'),
  }),
  outputSchema: z
    .object({
      files: z
        .array(
          z.object({
            name: z.string().describe('The name of the file or directory'),
            path: z.string().describe('The full path of the file or directory'),
            isDirectory: z.boolean().describe('Whether this is a directory'),
          }),
        )
        .describe('Array of files and directories'),
      path: z.string().describe('The path that was listed'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed file listing'),
      }),
    ),
  execute: async ({ sandboxId, path }, context) => {
    try {
      const sandbox = await getSandboxById(sandboxId)
      const normalizedPath = normalizeSandboxPath(path)

      const fileList = await sandbox.fs.listFiles(normalizedPath)

      const basePath = normalizedPath.endsWith('/') ? normalizedPath.slice(0, -1) : normalizedPath

      return {
        files: fileList.map((file) => ({
          name: file.name,
          path: `${basePath}/${file.name}`,
          isDirectory: file.isDir,
        })),
        path: normalizedPath,
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})

export const deleteFile = createTool({
  id: 'deleteFile',
  description: 'Delete a file or directory from the sandbox',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to delete the file from'),
    path: z.string().describe('The path to the file or directory to delete'),
  }),
  outputSchema: z
    .object({
      success: z.boolean().describe('Whether the file was deleted successfully'),
      path: z.string().describe('The path that was deleted'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed file deletion'),
      }),
    ),
  execute: async ({ sandboxId, path }, context) => {
    try {
      const sandbox = await getSandboxById(sandboxId)
      const normalizedPath = normalizeSandboxPath(path)

      // Check if the path is a directory
      const fileInfo = await sandbox.fs.getFileDetails(normalizedPath)
      const isDirectory = fileInfo.isDir

      await sandbox.fs.deleteFile(normalizedPath, isDirectory)

      return {
        success: true,
        path: normalizedPath,
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})

export const createDirectory = createTool({
  id: 'createDirectory',
  description: 'Create a directory in the sandbox',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to create the directory in'),
    path: z.string().describe('The path where the directory should be created'),
  }),
  outputSchema: z
    .object({
      success: z.boolean().describe('Whether the directory was created successfully'),
      path: z.string().describe('The path where the directory was created'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed directory creation'),
      }),
    ),
  execute: async ({ sandboxId, path }, context) => {
    try {
      const sandbox = await getSandboxById(sandboxId)
      const normalizedPath = normalizeSandboxPath(path)

      await sandbox.fs.createFolder(normalizedPath, '755')

      return {
        success: true,
        path: normalizedPath,
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})

export const getFileInfo = createTool({
  id: 'getFileInfo',
  description: 'Get detailed information about a file or directory in the sandbox',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to get file information from'),
    path: z.string().describe('The path to the file or directory to get information about'),
  }),
  outputSchema: z
    .object({
      group: z.string().describe(`The group ID of the file or directory (e.g., '1001')`),
      isDir: z.boolean().describe('Whether this is a directory'),
      modTime: z.string().describe(`The last modified time in UTC format (e.g., '2025-04-18 22:47:34 +0000 UTC')`),
      mode: z.string().describe(`The file mode/permissions in symbolic format (e.g., '-rw-r--r--')`),
      name: z.string().describe('The name of the file or directory'),
      owner: z.string().describe(`The owner ID of the file or directory (e.g., '1001')`),
      permissions: z.string().describe(`The file permissions in octal format (e.g., '0644')`),
      size: z.number().describe('The size of the file or directory in bytes'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed file info request'),
      }),
    ),
  execute: async ({ sandboxId, path }, context) => {
    try {
      const sandbox = await getSandboxById(sandboxId)
      const normalizedPath = normalizeSandboxPath(path)

      const fileInfo = await sandbox.fs.getFileDetails(normalizedPath)

      return {
        group: fileInfo.group,
        isDir: fileInfo.isDir,
        modTime: fileInfo.modTime,
        mode: fileInfo.mode,
        name: fileInfo.name,
        owner: fileInfo.owner,
        permissions: fileInfo.permissions,
        size: fileInfo.size,
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})

export const checkFileExists = createTool({
  id: 'checkFileExists',
  description: 'Check if a file or directory exists in the sandbox',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to check file existence in'),
    path: z.string().describe('The path to check for existence'),
  }),
  outputSchema: z
    .object({
      exists: z.boolean().describe('Whether the file or directory exists'),
      path: z.string().describe('The path that was checked'),
      isDirectory: z.boolean().optional().describe('If the path exists, whether it is a directory'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed existence check'),
      }),
    ),
  execute: async ({ sandboxId, path }, context) => {
    try {
      const sandbox = await getSandboxById(sandboxId)
      const normalizedPath = normalizeSandboxPath(path)

      try {
        const fileInfo = await sandbox.fs.getFileDetails(normalizedPath)
        return {
          exists: true,
          path: normalizedPath,
          isDirectory: fileInfo.isDir,
        }
      } catch (e) {
        if (e instanceof DaytonaNotFoundError)
          return {
            exists: false,
            path: normalizedPath,
          }
        else throw e
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})

export const getFileSize = createTool({
  id: 'getFileSize',
  description: 'Get the size of a file or directory in the sandbox',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to get file size from'),
    path: z.string().describe('The path to the file or directory'),
    humanReadable: z
      .boolean()
      .default(false)
      .describe(`Whether to return size in human-readable format (e.g., '1.5 KB', '2.3 MB')`),
  }),
  outputSchema: z
    .object({
      size: z.number().describe('The size in bytes'),
      humanReadableSize: z.string().optional().describe('Human-readable size string if requested'),
      path: z.string().describe('The path that was checked'),
      isDirectory: z.boolean().describe('Whether this is a directory'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed size check'),
      }),
    ),
  execute: async ({ sandboxId, path, humanReadable }, context) => {
    try {
      const sandbox = await getSandboxById(sandboxId)
      const normalizedPath = normalizeSandboxPath(path)

      const fileInfo = await sandbox.fs.getFileDetails(normalizedPath)

      let humanReadableSize: string | undefined

      if (humanReadable) {
        const bytes = fileInfo.size
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
        if (bytes === 0) {
          humanReadableSize = '0 B'
        } else {
          const i = Math.floor(Math.log(bytes) / Math.log(1024))
          const size = (bytes / Math.pow(1024, i)).toFixed(1)
          humanReadableSize = `${size} ${sizes[i]}`
        }
      }

      return {
        size: fileInfo.size,
        humanReadableSize,
        path: normalizedPath,
        isDirectory: fileInfo.isDir,
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})

export const watchDirectory = createTool({
  id: 'watchDirectory',
  description:
    '⚠️ NOT SUPPORTED - This tool is currently not supported in the sandbox environment. Do not use this tool.',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to watch directory in'),
    path: z.string().describe('The directory path to watch for changes'),
    recursive: z.boolean().default(false).describe('Whether to watch subdirectories recursively'),
    watchDuration: z.number().describe('How long to watch for changes in milliseconds (default 30 seconds)'),
  }),
  outputSchema: z
    .object({
      watchStarted: z.boolean().describe('Whether the watch was started successfully'),
      path: z.string().describe('The path that was watched'),
      events: z
        .array(
          z.object({
            type: z.string().describe('The type of filesystem event'),
            name: z.string().describe('The name of the file that changed'),
            timestamp: z.string().describe('When the event occurred'),
          }),
        )
        .describe('Array of filesystem events that occurred during the watch period'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed directory watch'),
      }),
    ),
  execute: async () => {
    return {
      error: 'Directory watching is currently not supported in the sandbox environment.',
    }
  },
})

export const runCommand = createTool({
  id: 'runCommand',
  description: 'Run a shell command in the sandbox',
  inputSchema: z.object({
    sandboxId: z.string().describe('The sandboxId for the sandbox to run the command in'),
    command: z.string().describe('The shell command to execute'),
    envs: z.record(z.string()).optional().describe('Environment variables to set for the command'),
    workingDirectory: z
      .string()
      .optional()
      .describe('The working directory for command execution. If not specified, uses the sandbox working directory.'),
    timeoutSeconds: z
      .number()
      .optional()
      .describe('Maximum time in seconds to wait for the command to complete. 0 means wait indefinitely.'),
    captureOutput: z.boolean().default(true).describe('Whether to capture stdout and stderr output'),
  }),
  outputSchema: z
    .object({
      success: z.boolean().describe('Whether the command executed successfully'),
      exitCode: z.number().describe('The exit code of the command'),
      stdout: z.string().describe('The standard output from the command'),
      command: z.string().describe('The command that was executed'),
      executionTime: z.number().describe('How long the command took to execute in milliseconds'),
    })
    .or(
      z.object({
        error: z.string().describe('The error from a failed command execution'),
      }),
    ),
  execute: async ({ sandboxId, command, envs, workingDirectory, timeoutSeconds, captureOutput }, context) => {
    try {
      const sandbox = await getSandboxById(sandboxId)

      const startTime = Date.now()
      const response = await sandbox.process.executeCommand(command, workingDirectory, envs ?? {}, timeoutSeconds)

      const executionTime = Date.now() - startTime

      return {
        success: response.exitCode === 0,
        exitCode: response.exitCode,
        stdout: captureOutput ? response.result : '',
        command,
        executionTime,
      }
    } catch (e) {
      return {
        error: JSON.stringify(e),
      }
    }
  },
})
