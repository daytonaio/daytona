/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import CodeBlock from '@/components/CodeBlock'
import { CopyButton } from '@/components/CopyButton'
import TooltipButton from '@/components/TooltipButton'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  codeSnippetSupportedLanguages,
  FileSystemActions,
  GitOperationsActions,
  ProcessCodeExecutionActions,
} from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { usePlaygroundSandbox } from '@/hooks/usePlaygroundSandbox'
import { createErrorMessageOutput } from '@/lib/playground'
import { cn } from '@/lib/utils'
import { CodeLanguage, Sandbox } from '@daytonaio/sdk'
import { ChevronUpIcon, Loader2, PanelBottom, Play, XIcon } from 'lucide-react'
import { ReactNode, useCallback, useMemo, useRef, useState } from 'react'
import { Group, Panel, usePanelRef } from 'react-resizable-panels'
import ResponseCard from '../ResponseCard'
import { Window, WindowContent, WindowTitleBar } from '../Window'

type CodeSnippetsSectionStartNewLinesData = {
  isFileSystemOperationsFirstSectionSnippet: boolean
  isGitOperationsFirstSectionSnippet: boolean
}

// Type for the keys
type CodeSnippetsSectionFlagKey = keyof CodeSnippetsSectionStartNewLinesData

const SandboxCodeSnippetsResponse = ({ className }: { className?: string }) => {
  const [codeSnippetLanguage, setCodeSnippetLanguage] = useState<CodeLanguage>(CodeLanguage.PYTHON)
  const [codeSnippetOutput, setCodeSnippetOutput] = useState<string | ReactNode>('')
  const [isCodeSnippetRunning, setIsCodeSnippetRunning] = useState<boolean>(false)

  const { sandboxParametersState, actionRuntimeError, getSandboxParametersInfo } = usePlayground()
  const { updateSandbox, createSandboxFromParams } = usePlaygroundSandbox(true)

  const {
    useLanguageParam,
    useResources,
    useResourcesCPU,
    useResourcesMemory,
    useResourcesDisk,
    createSandboxParamsExist,
    useAutoStopInterval,
    useAutoArchiveInterval,
    useAutoDeleteInterval,
    useSandboxCreateParams,
    useCustomSandboxSnapshotName,
    createSandboxFromImage,
    createSandboxFromSnapshot,
  } = getSandboxParametersInfo()

  // useRef prevents new object reference creation on every render which would triger useEffect calls on every render
  const codeSnippetsSectionStartNewLinesData = useRef<CodeSnippetsSectionStartNewLinesData>({
    isFileSystemOperationsFirstSectionSnippet: false,
    isGitOperationsFirstSectionSnippet: false,
  })
  // Reset values to false on every render beacuse of possible sections layout change
  for (const property in codeSnippetsSectionStartNewLinesData.current)
    codeSnippetsSectionStartNewLinesData.current[property as CodeSnippetsSectionFlagKey] = false

  // Helper method do determine number of new line characters on the beginning of each code snippets sections to ensure consistent spacing
  const getCodeSnippetsSectionStartNewLines = useCallback((sectionPropertyName: CodeSnippetsSectionFlagKey) => {
    // For first section we need double new line character, for others just one
    let sectionStartNewLines = '\n'
    if (!codeSnippetsSectionStartNewLinesData.current[sectionPropertyName]) {
      // Signalize that first snippet section is encountered so that subsequent sections use single new line character
      codeSnippetsSectionStartNewLinesData.current[sectionPropertyName] = true
      sectionStartNewLines = '\n\n'
    }
    return sectionStartNewLines
  }, [])

  const useConfigObject = false // Currently not needed, we use jwtToken for client config

  const fileSystemListFilesLocationSet = !actionRuntimeError[FileSystemActions.LIST_FILES]

  // All parameters are required
  const fileSystemCreateFolderParamsSet = !actionRuntimeError[FileSystemActions.CREATE_FOLDER]

  const fileSystemDeleteFileRequiredParamsSet = !actionRuntimeError[FileSystemActions.DELETE_FILE]
  const useFileSystemDeleteFileRecursive =
    fileSystemDeleteFileRequiredParamsSet && sandboxParametersState['deleteFileParams'].recursive === true

  const shellCommandExists = !actionRuntimeError[ProcessCodeExecutionActions.SHELL_COMMANDS_RUN]

  const codeToRunExists = !actionRuntimeError[ProcessCodeExecutionActions.CODE_RUN]

  const gitCloneOperationRequiredParamsSet = !actionRuntimeError[GitOperationsActions.GIT_CLONE]
  const useGitCloneBranch = sandboxParametersState['gitCloneParams'].branchToClone
  const useGitCloneCommitId = sandboxParametersState['gitCloneParams'].commitToClone
  const useGitCloneUsername = sandboxParametersState['gitCloneParams'].authUsername
  const useGitClonePassword = sandboxParametersState['gitCloneParams'].authPassword

  const gitStatusOperationLocationSet = !actionRuntimeError[GitOperationsActions.GIT_STATUS]

  const gitBranchesOperationLocationSet = !actionRuntimeError[GitOperationsActions.GIT_BRANCHES_LIST]

  const getImportsCodeSnippet = useCallback(() => {
    const python =
      [
        'from daytona import Daytona',
        useConfigObject ? 'DaytonaConfig' : '',
        useSandboxCreateParams
          ? createSandboxFromSnapshot
            ? 'CreateSandboxFromSnapshotParams'
            : 'CreateSandboxFromImageParams'
          : '',
        useResources ? 'Resources' : '',
        createSandboxFromImage ? 'Image' : '',
      ]
        .filter(Boolean)
        .join(', ') + '\n'
    const typeScript =
      ['import { Daytona', useConfigObject ? 'DaytonaConfig' : '', createSandboxFromImage ? 'Image' : '']
        .filter(Boolean)
        .join(', ') + " } from '@daytonaio/sdk'\n"
    return { python, typeScript }
  }, [useConfigObject, useSandboxCreateParams, createSandboxFromSnapshot, createSandboxFromImage, useResources])

  const getDaytonaConfigCodeSnippet = useCallback(() => {
    let python = '',
      typeScript = ''
    if (useConfigObject) {
      python = ['\n# Define the configuration', 'config = DaytonaConfig()'].filter(Boolean).join('\n') + '\n'
      typeScript =
        ['\n// Define the configuration', 'const config: DaytonaConfig = { }'].filter(Boolean).join('\n') + '\n'
    }
    return { python, typeScript }
  }, [useConfigObject])

  const getDaytonaClientCodeSnippet = useCallback(() => {
    const python = ['# Initialize the Daytona client', `daytona = Daytona(${useConfigObject ? 'config' : ''})`]
      .filter(Boolean)
      .join('\n')
    const typeScript = [
      '\t// Initialize the Daytona client',
      `\tconst daytona = new Daytona(${useConfigObject ? 'config' : ''})`,
    ]
      .filter(Boolean)
      .join('\n')
    return { python, typeScript }
  }, [useConfigObject])

  const getResourcesCodeSnippet = useCallback(() => {
    let python = '',
      typeScript = ''
    if (useResources) {
      const pythonResourcesIndentation = '\t'
      const typeScriptResourcesIndentation = '\t\t\t\t'
      python = [
        '\n\n# Create a Sandbox with custom resources\nresources = Resources(',
        useResourcesCPU
          ? `${pythonResourcesIndentation}cpu=${sandboxParametersState['resources']['cpu']}, # ${sandboxParametersState['resources']['cpu']} CPU cores`
          : '',
        useResourcesMemory
          ? `${pythonResourcesIndentation}memory=${sandboxParametersState['resources']['memory']}, # ${sandboxParametersState['resources']['memory']}GB RAM`
          : '',
        useResourcesDisk
          ? `${pythonResourcesIndentation}disk=${sandboxParametersState['resources']['disk']}, # ${sandboxParametersState['resources']['disk']}GB disk space`
          : '',
        ')',
      ]
        .filter(Boolean)
        .join('\n')
      typeScript = [
        `${typeScriptResourcesIndentation.slice(0, -1)}resources: {`,
        useResourcesCPU
          ? `${typeScriptResourcesIndentation}cpu: ${sandboxParametersState['resources']['cpu']}, // ${sandboxParametersState['resources']['cpu']} CPU cores`
          : '',
        useResourcesMemory
          ? `${typeScriptResourcesIndentation}memory: ${sandboxParametersState['resources']['memory']}, // ${sandboxParametersState['resources']['memory']}GB RAM`
          : '',
        useResourcesDisk
          ? `${typeScriptResourcesIndentation}disk: ${sandboxParametersState['resources']['disk']}, // ${sandboxParametersState['resources']['disk']}GB disk space`
          : '',
        `${typeScriptResourcesIndentation.slice(0, -1)}}`,
      ]
        .filter(Boolean)
        .join('\n')
    }
    return { python, typeScript }
  }, [useResources, useResourcesCPU, useResourcesMemory, useResourcesDisk, sandboxParametersState])

  const getSandboxParamsSnippet = useCallback(() => {
    let python = '',
      typeScript = ''
    if (useSandboxCreateParams) {
      const pythonIndentation = '\t'
      const typeScriptIndentation = '\t\t\t'
      python = [
        `\n\nparams = ${createSandboxFromSnapshot ? 'CreateSandboxFromSnapshotParams' : 'CreateSandboxFromImageParams'}(`,
        useCustomSandboxSnapshotName ? `${pythonIndentation}snapshot="${sandboxParametersState['snapshotName']}",` : '',
        createSandboxFromImage ? `${pythonIndentation}image=Image.debian_slim("3.12"),` : '',
        useResources ? `${pythonIndentation}resources=resources,` : '',
        useLanguageParam ? `${pythonIndentation}language="${sandboxParametersState['language']}"` : '',
        ...(createSandboxParamsExist
          ? [
              useAutoStopInterval
                ? `${pythonIndentation}auto_stop_interval=${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']}, # ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] == 0 ? 'Disables the auto-stop feature' : `Sandbox will be stopped after ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']} minute${(sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] as number) > 1 ? 's' : ''}`}` // useAutoStopInterval guarantes that value isn't undefined so we put as number to silence TS compiler
                : '',
              useAutoArchiveInterval
                ? `${pythonIndentation}auto_archive_interval=${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']}, # Auto-archive after a Sandbox has been stopped for ${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval'] == 0 ? '30 days' : `${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']} minutes`}`
                : '',
              useAutoDeleteInterval
                ? `${pythonIndentation}auto_delete_interval=${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']}, # ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] == 0 ? 'Sandbox will be deleted immediately after stopping' : sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] == -1 ? 'Auto-delete functionality disabled' : `Auto-delete after a Sandbox has been stopped for ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']} minutes`}`
                : '',
            ]
          : []),
        ')',
      ]
        .filter(Boolean)
        .join('\n')
      typeScript = [
        `{`,
        useCustomSandboxSnapshotName
          ? `${typeScriptIndentation}snapshot: '${sandboxParametersState['snapshotName']}',`
          : '',
        createSandboxFromImage ? `${typeScriptIndentation}image: Image.debianSlim("3.13"),` : '',
        getResourcesCodeSnippet().typeScript,
        useLanguageParam ? `${typeScriptIndentation}language: '${sandboxParametersState['language']}',` : '',
        ...(createSandboxParamsExist
          ? [
              useAutoStopInterval
                ? `${typeScriptIndentation}autoStopInterval: ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']}, // ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] == 0 ? 'Disables the auto-stop feature' : `Sandbox will be stopped after ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']} minute${(sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] as number) > 1 ? 's' : ''}`}` // useAutoStopInterval guarantes that value isn't undefined so we put as number to silence TS compiler
                : '',
              useAutoArchiveInterval
                ? `${typeScriptIndentation}autoArchiveInterval: ${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']}, // Auto-archive after a Sandbox has been stopped for ${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval'] == 0 ? '30 days' : `${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']} minutes`}`
                : '',
              useAutoDeleteInterval
                ? `${typeScriptIndentation}autoDeleteInterval: ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']}, // ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] == 0 ? 'Sandbox will be deleted immediately after stopping' : sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] == -1 ? 'Auto-delete functionality disabled' : `Auto-delete after a Sandbox has been stopped for ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']} minutes`}`
                : '',
            ]
          : []),
        `${typeScriptIndentation.slice(0, -1)}}`,
      ]
        .filter(Boolean)
        .join('\n')
    }
    return { python, typeScript }
  }, [
    useSandboxCreateParams,
    useResources,
    useLanguageParam,
    createSandboxParamsExist,
    useAutoStopInterval,
    getResourcesCodeSnippet,
    useAutoArchiveInterval,
    useAutoDeleteInterval,
    useCustomSandboxSnapshotName,
    sandboxParametersState,
    createSandboxFromImage,
    createSandboxFromSnapshot,
  ])

  const getDaytonaCreateSnippet = useCallback(() => {
    const python = [
      '\n# Create the Sandbox instance',
      `sandbox = daytona.create(${useSandboxCreateParams ? 'params' : ''})`,
      'print(f"Sandbox created:{sandbox.id}")',
    ].join('\n')
    const typeScript = [
      '\t\t// Create the Sandbox instance',
      `\t\tconst sandbox = await daytona.create(${useSandboxCreateParams ? getSandboxParamsSnippet().typeScript : ''})`,
    ].join('\n')
    return { python, typeScript }
  }, [useSandboxCreateParams, getSandboxParamsSnippet])

  const getFileSystemOperationsSnippet = useCallback(() => {
    const python = [],
      typeScript = []
    const pythonIndentation = '\t'
    const typeScriptIndentation = '\t\t\t'
    if (fileSystemCreateFolderParamsSet) {
      // First section always has double new line characters -> we don't need getCodeSnippetsSectionStartNewLines return value, just signalize that first section is found
      getCodeSnippetsSectionStartNewLines('isFileSystemOperationsFirstSectionSnippet')
      python.push(
        '\n\n# Create folder with specific permissions',
        `sandbox.fs.create_folder("${sandboxParametersState['createFolderParams'].folderDestinationPath}", "${sandboxParametersState['createFolderParams'].permissions}")`,
      )
      typeScript.push(
        `\n\n${typeScriptIndentation.slice(0, -1)}// Create folder with specific permissions`,
        `${typeScriptIndentation.slice(0, -1)}await sandbox.fs.createFolder("${sandboxParametersState['createFolderParams'].folderDestinationPath}", "${sandboxParametersState['createFolderParams'].permissions}")`,
      )
    }
    if (fileSystemListFilesLocationSet) {
      const sectionStartNewLines = getCodeSnippetsSectionStartNewLines('isFileSystemOperationsFirstSectionSnippet')
      python.push(
        `${sectionStartNewLines}# List files in a directory`,
        `files = sandbox.fs.list_files("${sandboxParametersState['listFilesParams'].directoryPath}")`,
        'for file in files:',
        `${pythonIndentation}print(f"Name: {file.name}")`,
        `${pythonIndentation}print(f"Is directory: {file.is_dir}")`,
        `${pythonIndentation}print(f"Size: {file.size}")`,
        `${pythonIndentation}print(f"Modified: {file.mod_time}")`,
      )
      typeScript.push(
        `${sectionStartNewLines}${typeScriptIndentation.slice(0, -1)}// List files in a directory`,
        `${typeScriptIndentation.slice(0, -1)}const files = await sandbox.fs.listFiles("${sandboxParametersState['listFilesParams'].directoryPath}")`,
        `${typeScriptIndentation.slice(0, -1)}files.forEach(file => {`,
        `${typeScriptIndentation}console.log(\`Name: \${file.name}\`)`,
        `${typeScriptIndentation}console.log(\`Is directory: \${file.isDir}\`)`,
        `${typeScriptIndentation}console.log(\`Size: \${file.size}\`)`,
        `${typeScriptIndentation}console.log(\`Modified: \${file.modTime}\`)`,
        `${typeScriptIndentation.slice(0, -1)}})`,
      )
    }
    if (fileSystemDeleteFileRequiredParamsSet) {
      const sectionStartNewLines = getCodeSnippetsSectionStartNewLines('isFileSystemOperationsFirstSectionSnippet')
      python.push(
        `${sectionStartNewLines}# Delete ${useFileSystemDeleteFileRecursive ? 'directory' : 'file'}`,
        `sandbox.fs.delete_file("${sandboxParametersState['deleteFileParams'].filePath}"${useFileSystemDeleteFileRecursive ? ', True' : ''})`,
      )
      typeScript.push(
        `${sectionStartNewLines}${typeScriptIndentation.slice(0, -1)}// Delete ${useFileSystemDeleteFileRecursive ? 'directory' : 'file'}`,
        `${typeScriptIndentation.slice(0, -1)}await sandbox.fs.deleteFile("${sandboxParametersState['deleteFileParams'].filePath}"${useFileSystemDeleteFileRecursive ? ', true' : ''})`,
      )
    }
    return { python: python.join('\n'), typeScript: typeScript.join('\n') }
  }, [
    sandboxParametersState,
    getCodeSnippetsSectionStartNewLines,
    fileSystemListFilesLocationSet,
    fileSystemCreateFolderParamsSet,
    fileSystemDeleteFileRequiredParamsSet,
    useFileSystemDeleteFileRecursive,
  ])

  const getLanguageCodeToRunSnippet = useCallback(() => {
    let python = '',
      typeScript = ''
    if (codeToRunExists) {
      const pythonIndentation = '\t'
      const typeScriptIndentation = '\t\t'
      python = [
        '\n\n# Run code securely inside the Sandbox',
        'response = sandbox.process.code_run(',
        `'''${sandboxParametersState['codeRunParams'].languageCode}'''`,
        ')',
        'if response.exit_code != 0:',
        `${pythonIndentation}print(f"Error: {response.exit_code} {response.result}")`,
        'else:',
        `${pythonIndentation}print(response.result)`,
      ].join('\n')
      typeScript = [
        `\n\n${typeScriptIndentation}// Run code securely inside the Sandbox`,
        `${typeScriptIndentation}const response = await sandbox.process.codeRun(\``,
        `${sandboxParametersState['codeRunParams'].languageCode}`,
        `${typeScriptIndentation}\`)`,
        `${typeScriptIndentation}if (response.exitCode !== 0) {`,
        `${typeScriptIndentation + '\t'}console.error("Error running code:", response.exitCode, response.result)`,
        `${typeScriptIndentation}} else {`,
        `${typeScriptIndentation + '\t'}console.log(response.result)`,
        `${typeScriptIndentation}}`,
      ].join('\n')
    }
    return { python, typeScript }
  }, [codeToRunExists, sandboxParametersState])

  const getShellCodeToRunSnippet = useCallback(() => {
    let python = '',
      typeScript = ''
    if (shellCommandExists) {
      python = [
        '\n\n# Execute shell commands',
        `response = sandbox.process.exec("${sandboxParametersState['shellCommandRunParams'].shellCommand}")`,
        'print(response.result)',
      ].join('\n')
      const typeScriptIndentation = '\t\t'
      typeScript = [
        `\n\n${typeScriptIndentation}// Execute shell commands`,
        `${typeScriptIndentation}const response = await sandbox.process.executeCommand('${sandboxParametersState['shellCommandRunParams'].shellCommand}')`,
        `${typeScriptIndentation}console.log(response.result)`,
      ].join('\n')
    }
    return { python, typeScript }
  }, [shellCommandExists, sandboxParametersState])

  const getGitOperationsSnippet = useCallback(() => {
    const python = [],
      typeScript = []
    const pythonIndentation = '\t'
    const typeScriptIndentation = '\t\t\t'
    if (gitCloneOperationRequiredParamsSet) {
      // First section always has double new line characters -> we don't need getCodeSnippetsSectionStartNewLines return value, just signalize that first section is found
      getCodeSnippetsSectionStartNewLines('isGitOperationsFirstSectionSnippet')
      python.push(
        '\n\n# Clone git repository',
        'sandbox.git.clone(',
        `${pythonIndentation}url="${sandboxParametersState['gitCloneParams'].repositoryURL}",`,
        `${pythonIndentation}path="${sandboxParametersState['gitCloneParams'].cloneDestinationPath}",`,
        useGitCloneBranch
          ? `${pythonIndentation}branch="${sandboxParametersState['gitCloneParams'].branchToClone}",`
          : '',
        useGitCloneCommitId
          ? `${pythonIndentation}commit_id="${sandboxParametersState['gitCloneParams'].commitToClone}",`
          : '',
        useGitCloneUsername
          ? `${pythonIndentation}username="${sandboxParametersState['gitCloneParams'].authUsername}",`
          : '',
        useGitClonePassword
          ? `${pythonIndentation}password="${sandboxParametersState['gitCloneParams'].authPassword}"`
          : '',
        ')',
      )
      typeScript.push(
        `\n\n${typeScriptIndentation.slice(0, -1)}// Clone git repository`,
        `${typeScriptIndentation.slice(0, -1)}await sandbox.git.clone(`,
        `${typeScriptIndentation}"${sandboxParametersState['gitCloneParams'].repositoryURL}",`,
        `${typeScriptIndentation}"${sandboxParametersState['gitCloneParams'].cloneDestinationPath}",`,
        useGitCloneBranch ? `${typeScriptIndentation}"${sandboxParametersState['gitCloneParams'].branchToClone}",` : '',
        useGitCloneCommitId
          ? `${typeScriptIndentation}"${sandboxParametersState['gitCloneParams'].commitToClone}",`
          : '',
        useGitCloneUsername
          ? `${typeScriptIndentation}"${sandboxParametersState['gitCloneParams'].authUsername}",`
          : '',
        useGitClonePassword ? `${typeScriptIndentation}"${sandboxParametersState['gitCloneParams'].authPassword}"` : '',
        `${typeScriptIndentation.slice(0, -1)})`,
      )
    }
    if (gitStatusOperationLocationSet) {
      const sectionStartNewLines = getCodeSnippetsSectionStartNewLines('isGitOperationsFirstSectionSnippet')
      python.push(
        `${sectionStartNewLines}# Get repository status`,
        `status = sandbox.git.status("${sandboxParametersState['gitStatusParams'].repositoryPath}")`,
        'print(f"Current branch: {status.current_branch}")',
        'print(f"Commits ahead: {status.ahead}")',
        'print(f"Commits behind: {status.behind}")',
        'for file in status.file_status:',
        '\tprint(f"File: {file.name}")',
      )
      typeScript.push(
        `${sectionStartNewLines}${typeScriptIndentation.slice(0, -1)}// Get repository status`,
        `${typeScriptIndentation.slice(0, -1)}const status = await sandbox.git.status("${sandboxParametersState['gitStatusParams'].repositoryPath}")`,
        `${typeScriptIndentation.slice(0, -1)}console.log(\`Current branch: \${status.currentBranch}\`)`,
        `${typeScriptIndentation.slice(0, -1)}console.log(\`Commits ahead: \${status.ahead}\`)`,
        `${typeScriptIndentation.slice(0, -1)}console.log(\`Commits behind: \${status.behind}\`)`,
        `${typeScriptIndentation.slice(0, -1)}status.fileStatus.forEach(file => {`,
        `${typeScriptIndentation}console.log(\`File: \${file.name}\`)`,
        `${typeScriptIndentation.slice(0, -1)}})`,
      )
    }
    if (gitBranchesOperationLocationSet) {
      const sectionStartNewLines = getCodeSnippetsSectionStartNewLines('isGitOperationsFirstSectionSnippet')
      python.push(
        `${sectionStartNewLines}# List branches`,
        `response = sandbox.git.branches("${sandboxParametersState['gitBranchesParams'].repositoryPath}")`,
        'for branch in response.branches:',
        '\tprint(f"Branch: {branch}")',
      )
      typeScript.push(
        `${sectionStartNewLines}${typeScriptIndentation.slice(0, -1)}// List branches`,
        `${typeScriptIndentation.slice(0, -1)}const response = await sandbox.git.branches("${sandboxParametersState['gitBranchesParams'].repositoryPath}")`,
        `${typeScriptIndentation.slice(0, -1)}response.branches.forEach(branch => {`,
        `${typeScriptIndentation}console.log(\`Branch: \${branch}\`)`,
        `${typeScriptIndentation.slice(0, -1)}})`,
      )
    }
    return { python: python.filter(Boolean).join('\n'), typeScript: typeScript.filter(Boolean).join('\n') }
  }, [
    sandboxParametersState,
    getCodeSnippetsSectionStartNewLines,
    gitCloneOperationRequiredParamsSet,
    useGitCloneBranch,
    useGitCloneCommitId,
    useGitCloneUsername,
    useGitClonePassword,
    gitStatusOperationLocationSet,
    gitBranchesOperationLocationSet,
  ])

  const sandboxCodeSnippetsData = useMemo(() => {
    const { python: pythonImport, typeScript: typeScriptImport } = getImportsCodeSnippet()
    const { python: pythonDaytonaConfig, typeScript: typeScriptDaytonaConfig } = getDaytonaConfigCodeSnippet()
    const { python: pythonDaytonaClient, typeScript: typeScriptDaytonaClient } = getDaytonaClientCodeSnippet()
    const { python: pythonResources } = getResourcesCodeSnippet()
    const { python: pythonSandboxParams } = getSandboxParamsSnippet()
    const { python: pythonDaytonaCreate, typeScript: typeScriptDaytonaCreate } = getDaytonaCreateSnippet()
    const { python: pythonLanguageCodeToRun, typeScript: typeScriptLanguageCodeToRun } = getLanguageCodeToRunSnippet()
    const { python: pythonShellCodeToRun, typeScript: typeScriptShellCodeToRun } = getShellCodeToRunSnippet()
    const { python: pythonGitOperations, typeScript: typeScriptGitOperations } = getGitOperationsSnippet()
    const { python: pythonFileSystemOperations, typeScript: typeScriptFileSystemOperations } =
      getFileSystemOperationsSnippet()
    return {
      [CodeLanguage.PYTHON]: {
        code: `${pythonImport}${pythonDaytonaConfig}
${pythonDaytonaClient}${pythonResources}${pythonSandboxParams}
${pythonDaytonaCreate}${pythonLanguageCodeToRun}${pythonShellCodeToRun}${pythonFileSystemOperations}${pythonGitOperations}`,
      },
      [CodeLanguage.TYPESCRIPT]: {
        code: `${typeScriptImport}${typeScriptDaytonaConfig}
async function main() {
${typeScriptDaytonaClient}
\ttry {
${typeScriptDaytonaCreate}${typeScriptLanguageCodeToRun}${typeScriptShellCodeToRun}${typeScriptFileSystemOperations}${typeScriptGitOperations}
\t} catch (error) {
\t\tconsole.error("Sandbox flow error:", error)
\t}
}
main().catch(console.error)`,
      },
      [CodeLanguage.JAVASCRIPT]: { code: '' }, // Currently to prevent ts error when indexing
    }
  }, [
    getImportsCodeSnippet,
    getDaytonaConfigCodeSnippet,
    getDaytonaClientCodeSnippet,
    getResourcesCodeSnippet,
    getSandboxParamsSnippet,
    getDaytonaCreateSnippet,
    getLanguageCodeToRunSnippet,
    getShellCodeToRunSnippet,
    getGitOperationsSnippet,
    getFileSystemOperationsSnippet,
  ])

  const runCodeSnippet = async () => {
    setIsCodeSnippetRunning(true)
    let codeSnippetOutput = 'Creating sandbox...\n'
    setCodeSnippetOutput(codeSnippetOutput)
    let sandbox: Sandbox | undefined

    try {
      sandbox = await createSandboxFromParams()
      await updateSandbox(sandbox)
      codeSnippetOutput = `Sandbox successfully created: ${sandbox.id}\n`
      setCodeSnippetOutput(codeSnippetOutput)
      if (codeToRunExists) {
        setCodeSnippetOutput(codeSnippetOutput + '\nRunning code...')
        const codeRunResponse = await sandbox.process.codeRun(
          sandboxParametersState['codeRunParams'].languageCode as string,
        ) // codeToRunExists guarantes that value isn't undefined so we put as string to silence TS compiler
        codeSnippetOutput += `\nCode run result: ${codeRunResponse.result}`
        setCodeSnippetOutput(codeSnippetOutput)
      }
      if (shellCommandExists) {
        setCodeSnippetOutput(codeSnippetOutput + '\nRunning shell command...')
        const shellCommandResponse = await sandbox.process.executeCommand(
          sandboxParametersState['shellCommandRunParams'].shellCommand as string, // shellCommandExists guarantes that value isn't undefined so we put as string to silence TS compiler
        )
        codeSnippetOutput += `\nShell command result: ${shellCommandResponse.result}`
        setCodeSnippetOutput(codeSnippetOutput)
      }
      let codeRunShellCommandFinishedMessage = '\n'
      if (codeToRunExists && shellCommandExists) {
        codeRunShellCommandFinishedMessage += '🎉 Code and shell command executed successfully.'
      } else if (codeToRunExists) {
        codeRunShellCommandFinishedMessage += '🎉 Code executed successfully.'
      } else if (shellCommandExists) {
        codeRunShellCommandFinishedMessage += '🎉 Shell command executed successfully.'
      }
      codeSnippetOutput += codeRunShellCommandFinishedMessage + '\n'
      setCodeSnippetOutput(codeSnippetOutput)
      if (fileSystemCreateFolderParamsSet) {
        setCodeSnippetOutput(codeSnippetOutput + '\nCreating directory...')
        await sandbox.fs.createFolder(
          sandboxParametersState['createFolderParams'].folderDestinationPath,
          sandboxParametersState['createFolderParams'].permissions,
        )
        codeSnippetOutput += '\n🎉 Directory created successfully.\n'
        setCodeSnippetOutput(codeSnippetOutput)
      }
      if (fileSystemListFilesLocationSet) {
        setCodeSnippetOutput(codeSnippetOutput + '\nListing directory files...')
        const files = await sandbox.fs.listFiles(sandboxParametersState['listFilesParams'].directoryPath)
        codeSnippetOutput += '\nDirectory content:'
        codeSnippetOutput += '\n'
        files.forEach((file) => {
          codeSnippetOutput += `Name: ${file.name}\n`
          codeSnippetOutput += `Is directory: ${file.isDir}\n`
          codeSnippetOutput += `Size: ${file.size}\n`
          codeSnippetOutput += `Modified: ${file.modTime}\n`
        })
        setCodeSnippetOutput(codeSnippetOutput)
      }
      if (fileSystemDeleteFileRequiredParamsSet) {
        setCodeSnippetOutput(
          codeSnippetOutput + `\nDeleting ${useFileSystemDeleteFileRecursive ? 'directory' : 'file'}...`,
        )
        await sandbox.fs.deleteFile(
          sandboxParametersState['deleteFileParams'].filePath,
          useFileSystemDeleteFileRecursive || false,
        )
        codeSnippetOutput += `\n🎉 ${useFileSystemDeleteFileRecursive ? 'Directory' : 'File'} deleted successfully.\n`
        setCodeSnippetOutput(codeSnippetOutput)
      }
      if (gitCloneOperationRequiredParamsSet) {
        setCodeSnippetOutput(codeSnippetOutput + '\nCloning repo...')
        await sandbox.git.clone(
          sandboxParametersState['gitCloneParams'].repositoryURL,
          sandboxParametersState['gitCloneParams'].cloneDestinationPath,
          useGitCloneBranch ? sandboxParametersState['gitCloneParams'].branchToClone : undefined,
          useGitCloneCommitId ? sandboxParametersState['gitCloneParams'].commitToClone : undefined,
          useGitCloneUsername ? sandboxParametersState['gitCloneParams'].authUsername : undefined,
          useGitClonePassword ? sandboxParametersState['gitCloneParams'].authPassword : undefined,
        )
        codeSnippetOutput += '\n🎉 Repository cloned successfully.\n'
        setCodeSnippetOutput(codeSnippetOutput)
      }
      if (gitStatusOperationLocationSet) {
        setCodeSnippetOutput(codeSnippetOutput + '\nFetching repository status...')
        const status = await sandbox.git.status(sandboxParametersState['gitStatusParams'].repositoryPath)
        codeSnippetOutput += `\nCurrent branch: ${status.currentBranch}\n`
        codeSnippetOutput += `Commits ahead: ${status.ahead}\n`
        codeSnippetOutput += `Commits behind: ${status.behind}\n`
        status.fileStatus.forEach((file) => (codeSnippetOutput += `File: ${file.name}\n`))
        setCodeSnippetOutput(codeSnippetOutput)
      }
      if (gitBranchesOperationLocationSet) {
        setCodeSnippetOutput(codeSnippetOutput + '\nFetching repository branches...')
        const response = await sandbox.git.branches(sandboxParametersState['gitBranchesParams'].repositoryPath)
        codeSnippetOutput += '\n'
        response.branches.forEach((branch) => (codeSnippetOutput += `Branch: ${branch}\n`))
        setCodeSnippetOutput(codeSnippetOutput)
      }
      setCodeSnippetOutput(codeSnippetOutput + '\nSandbox session finished.')
    } catch (error) {
      console.error(error)
      setCodeSnippetOutput(
        <>
          <span>{codeSnippetOutput}</span>
          <br />
          {createErrorMessageOutput(error)}
        </>,
      )
    } finally {
      setIsCodeSnippetRunning(false)
    }
  }

  const resultPanelRef = usePanelRef()

  return (
    <Window className={className}>
      <WindowTitleBar>Sandbox Code</WindowTitleBar>
      <WindowContent className="relative">
        <Tabs
          value={codeSnippetLanguage}
          className="flex flex-col gap-4"
          onValueChange={(languageValue) => setCodeSnippetLanguage(languageValue as CodeLanguage)}
        >
          <div className="flex justify-between items-center">
            <TabsList>
              {codeSnippetSupportedLanguages.map((language) => (
                <TabsTrigger
                  key={language.value}
                  value={language.value}
                  className={cn('py-1 rounded-t-md', {
                    'bg-foreground/10': codeSnippetLanguage === language.value,
                  })}
                >
                  <div className="flex items-center text-sm">
                    <img src={language.icon} alt={language.icon} className="w-4 h-4" />
                    <span className="ml-2">{language.label}</span>
                  </div>
                </TabsTrigger>
              ))}
            </TabsList>
            <div className="flex items-center gap-2">
              <Button
                disabled={isCodeSnippetRunning}
                variant="outline"
                className="ml-auto"
                onClick={() => {
                  runCodeSnippet()
                  if (resultPanelRef.current?.isCollapsed()) {
                    resultPanelRef.current.resize(100)
                  }
                }}
              >
                {isCodeSnippetRunning ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="w-4 h-4" />} Run
              </Button>
              <TooltipButton
                tooltipText="Show result"
                className="!px-2"
                size="icon-sm"
                variant="outline"
                onClick={() => {
                  if (resultPanelRef.current?.isCollapsed()) {
                    resultPanelRef.current.resize('20%')
                  } else {
                    resultPanelRef.current?.collapse()
                  }
                }}
              >
                <PanelBottom />
              </TooltipButton>
            </div>
          </div>
          <Group orientation="vertical" className="min-h-[500px] border-border rounded-b-md">
            <Panel minSize={'20%'}>
              {codeSnippetSupportedLanguages.map((language) => (
                <TabsContent
                  key={language.value}
                  value={language.value}
                  className="rounded-md h-full overflow-auto mt-0"
                >
                  <CopyButton
                    className="absolute right-4 z-10 backdrop-blur-sm"
                    variant="ghost"
                    size="icon-sm"
                    value={sandboxCodeSnippetsData[language.value].code}
                  />
                  <ScrollArea fade="mask" horizontal className="h-full overflow-auto" fadeOffset={35}>
                    <CodeBlock
                      showCopy={false}
                      language={language.value}
                      code={sandboxCodeSnippetsData[language.value].code}
                      codeAreaClassName="text-sm [overflow:initial] min-w-fit"
                    />
                  </ScrollArea>
                </TabsContent>
              ))}
            </Panel>

            <Panel maxSize="80%" minSize="20%" panelRef={resultPanelRef} collapsedSize={0} collapsible defaultSize={33}>
              <div className="bg-background w-full border rounded-md overflow-auto h-full flex flex-col">
                <div className="flex justify-between border-b px-4 pr-2 py-1 text-xs items-center bg-muted/50">
                  <div className="text-muted-foreground font-mono">Result</div>
                  <div className="flex items-center gap-2">
                    <TooltipButton
                      onClick={() => resultPanelRef.current?.resize('80%')}
                      tooltipText="Maximize"
                      className="h-6 w-6"
                      size="sm"
                      variant="ghost"
                    >
                      <ChevronUpIcon className="w-4 h-4" />
                    </TooltipButton>
                    <TooltipButton
                      tooltipText="Close"
                      className="h-6 w-6"
                      size="sm"
                      variant="ghost"
                      onClick={() => resultPanelRef.current?.collapse()}
                    >
                      <XIcon />
                    </TooltipButton>
                  </div>
                </div>
                <div className="flex-1 overflow-y-auto">
                  <ResponseCard
                    responseContent={
                      codeSnippetOutput || (
                        <div className="text-muted-foreground font-mono">Code output will be shown here...</div>
                      )
                    }
                  />
                </div>
              </div>
            </Panel>
          </Group>
        </Tabs>
      </WindowContent>
    </Window>
  )
}

export default SandboxCodeSnippetsResponse
