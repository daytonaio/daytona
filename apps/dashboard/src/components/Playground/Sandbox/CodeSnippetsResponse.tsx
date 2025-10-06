/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Image,
  CodeLanguage,
  CreateSandboxFromSnapshotParams,
  CreateSandboxFromImageParams,
  Sandbox,
} from '@daytonaio/sdk'
import { codeSnippetSupportedLanguages, GitOperationsActions, ProcessCodeExecutionActions } from '@/enums/Playground'
import CodeBlock from '@/components/CodeBlock'
import { Button } from '@/components/ui/button'
import { usePlayground } from '@/hooks/usePlayground'
import { Loader2, Play } from 'lucide-react'
import { createErrorMessageOutput } from '@/lib/playground'
import { ReactNode, useCallback, useMemo, useState } from 'react'
import ResponseCard from '../ResponseCard'

const SandboxCodeSnippetsResponse: React.FC = () => {
  const [codeSnippetLanguage, setCodeSnippetLanguage] = useState<CodeLanguage>(CodeLanguage.PYTHON)
  const [codeSnippetOutput, setCodeSnippetOutput] = useState<string | ReactNode>('')
  const [isCodeSnippetRunning, setIsCodeSnippetRunning] = useState<boolean>(false)

  const { sandboxParametersState, DaytonaClient, actionRuntimeError } = usePlayground()

  const objectHasAnyValue = (obj: object) => Object.values(obj).some((v) => v !== '' && v !== undefined)

  const useConfigObject = false // Currently not needed, we use jwtToken for client config

  const useLanguageParam = sandboxParametersState['language']

  const shellCommandExists = !actionRuntimeError[ProcessCodeExecutionActions.SHELL_COMMANDS_RUN]

  const codeToRunExists = !actionRuntimeError[ProcessCodeExecutionActions.CODE_RUN]

  const gitCloneOperationRequiredParamsSet = !actionRuntimeError[GitOperationsActions.GIT_CLONE]
  const useGitCloneBranch = sandboxParametersState['gitCloneParams'].branchToClone
  const useGitCloneCommitId = sandboxParametersState['gitCloneParams'].commitToClone
  const useGitCloneUsername = sandboxParametersState['gitCloneParams'].authUsername
  const useGitClonePassword = sandboxParametersState['gitCloneParams'].authPassword

  const gitStatusOperationLocationSet = !actionRuntimeError[GitOperationsActions.GIT_STATUS]

  const gitBranchesOperationLocationSet = !actionRuntimeError[GitOperationsActions.GIT_BRANCHES_LIST]

  const useResources = objectHasAnyValue(sandboxParametersState['resources'])
  const useResourcesCPU = useResources && sandboxParametersState['resources']['cpu'] !== undefined
  const useResourcesMemory = useResources && sandboxParametersState['resources']['memory'] !== undefined
  const useResourcesDisk = useResources && sandboxParametersState['resources']['disk'] !== undefined

  const createSandboxParamsExist = objectHasAnyValue(sandboxParametersState['createSandboxBaseParams'])
  const useAutoStopInterval =
    createSandboxParamsExist && sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] !== undefined
  const useAutoArchiveInterval =
    createSandboxParamsExist && sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval'] !== undefined
  const useAutoDeleteInterval =
    createSandboxParamsExist && sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] !== undefined

  const useSandboxCreateParams = useLanguageParam || useResources || createSandboxParamsExist

  const getImportsCodeSnippet = useCallback(() => {
    const python =
      [
        'from daytona import Daytona',
        useConfigObject ? 'DaytonaConfig' : '',
        useSandboxCreateParams ? 'CreateSandboxFromImageParams' : '',
        useResources ? 'Resources' : '',
        useSandboxCreateParams ? 'Image' : '',
      ]
        .filter(Boolean)
        .join(', ') + '\n'
    const typeScript =
      ['import { Daytona', useConfigObject ? 'DaytonaConfig' : '', useSandboxCreateParams ? 'Image' : '']
        .filter(Boolean)
        .join(', ') + " } from '@daytonaio/sdk'\n"
    return { python, typeScript }
  }, [useConfigObject, useSandboxCreateParams, useResources])

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
        '\n\nparams = CreateSandboxFromImageParams(',
        `${pythonIndentation}image=Image.debian_slim("3.12"),`,
        useResources ? `${pythonIndentation}resources=resources,` : '',
        useLanguageParam ? `${pythonIndentation}language="${sandboxParametersState['language']}"` : '',
        ...(createSandboxParamsExist
          ? [
              useAutoStopInterval
                ? `${pythonIndentation}auto_stop_interval=${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']}, # ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] == 0 ? 'Disables the auto-stop feature' : `Sandbox will be stopped after ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']} minute${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] > 1 ? 's' : ''}`}`
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
        `{\n${typeScriptIndentation}image: Image.debianSlim("3.13"),`,
        getResourcesCodeSnippet().typeScript,
        useLanguageParam ? `${typeScriptIndentation}language: '${sandboxParametersState['language']}',` : '',
        ...(createSandboxParamsExist
          ? [
              useAutoStopInterval
                ? `${typeScriptIndentation}autoStopInterval: ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']}, // ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] == 0 ? 'Disables the auto-stop feature' : `Sandbox will be stopped after ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']} minute${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] > 1 ? 's' : ''}`}`
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
    sandboxParametersState,
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
      python.push(
        '\n# Get repository status',
        `status = sandbox.git.status("${sandboxParametersState['gitStatusParams'].repositoryPath}")`,
        'print(f"Current branch: {status.current_branch}")',
        'print(f"Commits ahead: {status.ahead}")',
        'print(f"Commits behind: {status.behind}")',
        'for file in status.file_status:',
        '\tprint(f"File: {file.name}")',
      )
      typeScript.push(
        `\n${typeScriptIndentation.slice(0, -1)}// Get repository status`,
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
      python.push(
        '\n# List branches',
        `response = sandbox.git.branches("${sandboxParametersState['gitBranchesParams'].repositoryPath}")`,
        'for branch in response.branches:',
        '\tprint(f"Branch: {branch}")',
      )
      typeScript.push(
        `\n${typeScriptIndentation.slice(0, -1)}// List branches`,
        `${typeScriptIndentation.slice(0, -1)}const response = await sandbox.git.branches("${sandboxParametersState['gitBranchesParams'].repositoryPath}")`,
        `${typeScriptIndentation.slice(0, -1)}response.branches.forEach(branch => {`,
        `${typeScriptIndentation}console.log(\`Branch: \${branch}\`)`,
        `${typeScriptIndentation.slice(0, -1)}})`,
      )
    }
    return { python: python.filter(Boolean).join('\n'), typeScript: typeScript.filter(Boolean).join('\n') }
  }, [
    sandboxParametersState,
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
    return {
      [CodeLanguage.PYTHON]: {
        code: `${pythonImport}${pythonDaytonaConfig}
${pythonDaytonaClient}${pythonResources}${pythonSandboxParams}
${pythonDaytonaCreate}${pythonLanguageCodeToRun}${pythonShellCodeToRun}${pythonGitOperations}`,
      },
      [CodeLanguage.TYPESCRIPT]: {
        code: `${typeScriptImport}${typeScriptDaytonaConfig}
async function main() {
${typeScriptDaytonaClient}
\ttry {
${typeScriptDaytonaCreate}${typeScriptLanguageCodeToRun}${typeScriptShellCodeToRun}${typeScriptGitOperations}
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
  ])

  const runCodeSnippet = async () => {
    setIsCodeSnippetRunning(true)
    let codeSnippetOutput = 'Creating sandbox...\n'
    setCodeSnippetOutput(codeSnippetOutput)
    let sandbox: Sandbox

    try {
      let createSandboxFromImageParams: CreateSandboxFromImageParams
      const createSandboxFromSnapshotParams: CreateSandboxFromSnapshotParams = { snapshot: 'ubuntu:24.04' }
      const createSandboxFromImage = useSandboxCreateParams
      if (createSandboxFromImage) {
        // Set CreateSandboxFromImageParams specific params
        createSandboxFromImageParams = { image: Image.debianSlim('3.13') }
        if (useResources) {
          createSandboxFromImageParams.resources = {}
          if (useResourcesCPU) createSandboxFromImageParams.resources.cpu = sandboxParametersState['resources']['cpu']
          if (useResourcesMemory)
            createSandboxFromImageParams.resources.memory = sandboxParametersState['resources']['memory']
          if (useResourcesDisk)
            createSandboxFromImageParams.resources.disk = sandboxParametersState['resources']['disk']
        }
      }
      const createSandboxParams: CreateSandboxFromImageParams | CreateSandboxFromSnapshotParams = createSandboxFromImage
        ? createSandboxFromImageParams
        : createSandboxFromSnapshotParams
      // Set CreateSandboxBaseParams params which are common for both params types
      if (useLanguageParam) createSandboxParams.language = sandboxParametersState['language']
      if (useAutoStopInterval)
        createSandboxParams.autoStopInterval = sandboxParametersState['createSandboxBaseParams']['autoStopInterval']
      if (useAutoArchiveInterval)
        createSandboxParams.autoArchiveInterval =
          sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']
      if (useAutoDeleteInterval)
        createSandboxParams.autoDeleteInterval = sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']
      createSandboxParams.labels = { 'daytona-playground': 'true' }
      if (useLanguageParam)
        createSandboxParams.labels['daytona-playground-language'] = sandboxParametersState['language']
      sandbox = await DaytonaClient.create(createSandboxParams)
      codeSnippetOutput = `Sandbox successfully created: ${sandbox.id}\n`
      setCodeSnippetOutput(codeSnippetOutput)
      if (codeToRunExists) {
        setCodeSnippetOutput(codeSnippetOutput + '\nRunning code...')
        const codeRunResponse = await sandbox.process.codeRun(sandboxParametersState['codeRunParams'].languageCode)
        codeSnippetOutput += `\nCode run result: ${codeRunResponse.result}`
        setCodeSnippetOutput(codeSnippetOutput)
      }
      if (shellCommandExists) {
        setCodeSnippetOutput(codeSnippetOutput + '\nRunning shell command...')
        const shellCommandResponse = await sandbox.process.executeCommand(
          sandboxParametersState['shellCommandRunParams'].shellCommand,
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
      if (sandbox) {
        try {
          await sandbox.delete()
        } catch (cleanupError) {
          console.error('Failed to delete sandbox during cleanup:', cleanupError)
        }
      }
      setIsCodeSnippetRunning(false)
    }
  }

  return (
    <>
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Code</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs
            value={codeSnippetLanguage}
            className="flex flex-col"
            onValueChange={(languageValue) => setCodeSnippetLanguage(languageValue as CodeLanguage)}
          >
            <TabsList className="px-4 w-full flex-shrink-0">
              {codeSnippetSupportedLanguages.map((language) => (
                <TabsTrigger
                  key={language.value}
                  value={language.value}
                  className={codeSnippetLanguage === language.value ? 'bg-foreground/10' : ''}
                >
                  <div className="flex items-center text-sm">
                    <img src={language.icon} alt={language.icon} className="w-4 h-4" />
                    <span className="ml-2">{language.label}</span>
                  </div>
                </TabsTrigger>
              ))}
              <Button disabled={isCodeSnippetRunning} variant="outline" className="ml-auto" onClick={runCodeSnippet}>
                {isCodeSnippetRunning ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="w-4 h-4" />} Run
              </Button>
            </TabsList>
            {codeSnippetSupportedLanguages.map((language) => (
              <TabsContent key={language.value} value={language.value}>
                <CodeBlock
                  language={language.value}
                  code={sandboxCodeSnippetsData[language.value].code}
                  codeAreaClassName="overflow-y-scroll h-[350px]"
                />
              </TabsContent>
            ))}
          </Tabs>
        </CardContent>
      </Card>
      <ResponseCard titleText="Result" responseText={codeSnippetOutput} />
    </>
  )
}

export default SandboxCodeSnippetsResponse
