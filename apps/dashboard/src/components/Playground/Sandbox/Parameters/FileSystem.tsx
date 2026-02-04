/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  CreateFolderParams,
  DeleteFileParams,
  FileSystemActionFormData,
  FileSystemActions,
  ListFilesParams,
  ParameterFormData,
  ParameterFormItem,
} from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { useState } from 'react'
import PlaygroundActionForm from '../../ActionForm'
import FormCheckboxInput from '../../Inputs/CheckboxInput'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormTextInput from '../../Inputs/TextInput'

const SandboxFileSystem: React.FC = () => {
  const { sandboxParametersState, playgroundActionParamValueSetter } = usePlayground()
  const [createFolderParams, setCreateFolderParams] = useState<CreateFolderParams>(
    sandboxParametersState['createFolderParams'],
  )
  const [listFilesParams, setListFilesParams] = useState<ListFilesParams>(sandboxParametersState['listFilesParams'])
  const [deleteFileParams, setDeleteFileParams] = useState<DeleteFileParams>(sandboxParametersState['deleteFileParams'])

  const listFilesDirectoryFormData: ParameterFormItem & { key: 'directoryPath' } = {
    label: 'Directory location',
    key: 'directoryPath',
    placeholder: 'Directory path to list',
    required: true,
  }

  const createFolderParamsFormData: ParameterFormData<CreateFolderParams> = [
    {
      label: 'Folder location',
      key: 'folderDestinationPath',
      placeholder: 'Path where the directory should be created',
      required: true,
    },
    {
      label: 'Permissions',
      key: 'permissions',
      placeholder: 'Directory permissions in octal format (e.g. "755")',
      required: true,
    },
  ]

  const deleteFileLocationFormData: ParameterFormItem & { key: 'filePath' } = {
    label: 'File location',
    key: 'filePath',
    placeholder: 'Path to the file or directory to delete',
    required: true,
  }
  const deleteFileRecursiveFormData: ParameterFormItem & { key: 'recursive' } = {
    label: 'Delete directory',
    key: 'recursive',
    placeholder: 'If the file is a directory, this must be true to delete it.',
  }

  const fileSystemActionsFormData: FileSystemActionFormData<ListFilesParams | CreateFolderParams | DeleteFileParams>[] =
    [
      {
        methodName: FileSystemActions.CREATE_FOLDER,
        label: 'createFolder()',
        description: 'Creates a new directory in the Sandbox at the specified path with the given permissions',
        parametersFormItems: createFolderParamsFormData,
        parametersState: createFolderParams,
      },
      {
        methodName: FileSystemActions.LIST_FILES,
        label: 'listFiles()',
        description: 'Lists files and directories in a given path and returns their information',
        parametersFormItems: [listFilesDirectoryFormData],
        parametersState: listFilesParams,
      },
      {
        methodName: FileSystemActions.DELETE_FILE,
        label: 'deleteFile()',
        description: 'Deletes a file from the Sandbox',
        parametersFormItems: [deleteFileLocationFormData, deleteFileRecursiveFormData],
        parametersState: deleteFileParams,
      },
    ]

  return (
    <div className="space-y-6">
      {fileSystemActionsFormData.map((fileSystemAction) => (
        <div key={fileSystemAction.methodName} className="space-y-4">
          <PlaygroundActionForm<FileSystemActions> actionFormItem={fileSystemAction} hideRunActionButton />
          <div className="space-y-2">
            {fileSystemAction.methodName === FileSystemActions.LIST_FILES && (
              <InlineInputFormControl formItem={listFilesDirectoryFormData}>
                <FormTextInput
                  formItem={listFilesDirectoryFormData}
                  textValue={listFilesParams[listFilesDirectoryFormData.key]}
                  onChangeHandler={(value) =>
                    playgroundActionParamValueSetter(
                      fileSystemAction,
                      listFilesDirectoryFormData,
                      setListFilesParams,
                      'listFilesParams',
                      value,
                    )
                  }
                />
              </InlineInputFormControl>
            )}
            {fileSystemAction.methodName === FileSystemActions.CREATE_FOLDER && (
              <>
                {createFolderParamsFormData.map((createFolderParamFormItem) => (
                  <InlineInputFormControl key={createFolderParamFormItem.key} formItem={createFolderParamFormItem}>
                    <FormTextInput
                      formItem={createFolderParamFormItem}
                      textValue={createFolderParams[createFolderParamFormItem.key]}
                      onChangeHandler={(value) =>
                        playgroundActionParamValueSetter(
                          fileSystemAction,
                          createFolderParamFormItem,
                          setCreateFolderParams,
                          'createFolderParams',
                          value,
                        )
                      }
                    />
                  </InlineInputFormControl>
                ))}
              </>
            )}
            {fileSystemAction.methodName === FileSystemActions.DELETE_FILE && (
              <>
                <InlineInputFormControl formItem={deleteFileLocationFormData}>
                  <FormTextInput
                    formItem={deleteFileLocationFormData}
                    textValue={deleteFileParams[deleteFileLocationFormData.key]}
                    onChangeHandler={(value) =>
                      playgroundActionParamValueSetter(
                        fileSystemAction,
                        deleteFileLocationFormData,
                        setDeleteFileParams,
                        'deleteFileParams',
                        value,
                      )
                    }
                  />
                </InlineInputFormControl>
                <InlineInputFormControl formItem={deleteFileRecursiveFormData}>
                  <FormCheckboxInput
                    formItem={deleteFileRecursiveFormData}
                    checkedValue={deleteFileParams[deleteFileRecursiveFormData.key]}
                    onChangeHandler={(checked) =>
                      playgroundActionParamValueSetter(
                        fileSystemAction,
                        deleteFileRecursiveFormData,
                        setDeleteFileParams,
                        'deleteFileParams',
                        checked,
                      )
                    }
                  />
                </InlineInputFormControl>
              </>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

export default SandboxFileSystem
