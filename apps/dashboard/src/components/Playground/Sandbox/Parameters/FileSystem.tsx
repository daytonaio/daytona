/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { FileSystemActions, PlaygroundActionFormDataBasic } from '@/enums/Playground'
import PlaygroundActionForm from '../../ActionForm'

const SandboxFileSystem: React.FC = () => {
  const fileSystemActionsFormData: PlaygroundActionFormDataBasic<FileSystemActions>[] = [
    {
      methodName: FileSystemActions.LIST_FILES,
      label: 'listFiles()',
      description: 'Lists files and directories in a given path and returns their information',
    },
    {
      methodName: FileSystemActions.CREATE_FOLDER,
      label: 'createFolder()',
      description: 'Creates a new directory in the Sandbox at the specified path with the given permissions',
    },
    {
      methodName: FileSystemActions.UPLOAD_FILE,
      label: 'uploadFile()',
      description: 'Uploads a file to the specified path in the Sandbox',
    },
    {
      methodName: FileSystemActions.UPLOAD_FILES,
      label: 'uploadFiles()',
      description: 'Uploads multiple files to the Sandbox',
    },
    {
      methodName: FileSystemActions.DOWNLOAD_FILE,
      label: 'downloadFile()',
      description: 'Downloads a file from the Sandbox',
    },
    {
      methodName: FileSystemActions.DOWNLOAD_FILES,
      label: 'downloadFiles()',
      description: 'Downloads multiple files from the Sandbox',
    },
    {
      methodName: FileSystemActions.DELETE_FILE,
      label: 'deleteFile()',
      description: 'Deletes a file from the Sandbox',
    },
  ]

  return (
    <div className="space-y-6">
      {fileSystemActionsFormData.map((fileSystemActionFormData) => (
        <div key={fileSystemActionFormData.methodName} className="space-y-4">
          <PlaygroundActionForm<FileSystemActions> actionFormItem={fileSystemActionFormData} hideRunActionButton />
        </div>
      ))}
    </div>
  )
}

export default SandboxFileSystem
