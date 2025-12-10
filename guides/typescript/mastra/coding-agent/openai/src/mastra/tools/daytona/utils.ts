import { Daytona, Sandbox } from '@daytonaio/sdk'
import { FileUpload } from '@daytonaio/sdk/src/FileSystem'

let daytonaInstance: Daytona | null = null

export const getDaytonaClient = () => {
  if (!daytonaInstance) {
    daytonaInstance = new Daytona()
  }
  return daytonaInstance
}

export const getSandboxById = async (sandboxId: string): Promise<Sandbox> => {
  const daytona = getDaytonaClient()
  const sandbox = await daytona.get(sandboxId)
  return sandbox
}

export const createFileUploadFormat = (content: string, path: string): FileUpload => {
  return {
    source: Buffer.from(content, 'utf-8'),
    destination: path,
  }
}

// Default working directory for Daytona sandboxes
const DEFAULT_WORKING_DIR = '/home/daytona'

export const normalizeSandboxPath = (path: string): string => {
  // If path already starts with the working directory, return as-is
  if (path.startsWith(DEFAULT_WORKING_DIR)) {
    return path
  }

  // If path starts with ./, remove the dot and treat as relative
  if (path.startsWith('./')) {
    return `${DEFAULT_WORKING_DIR}${path.slice(1)}`
  }

  // If path starts with /, treat it as relative to working directory
  if (path.startsWith('/')) {
    return `${DEFAULT_WORKING_DIR}${path}`
  }

  // For relative paths, prepend working directory
  return `${DEFAULT_WORKING_DIR}/${path}`
}
