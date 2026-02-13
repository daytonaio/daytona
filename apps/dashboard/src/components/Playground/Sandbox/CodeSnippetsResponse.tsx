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
import PythonIcon from '@/assets/python.svg'
import TypescriptIcon from '@/assets/typescript.svg'
import { FileSystemActions, GitOperationsActions, ProcessCodeExecutionActions } from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { usePlaygroundSandbox } from '@/hooks/usePlaygroundSandbox'
import { createErrorMessageOutput } from '@/lib/playground'
import { cn } from '@/lib/utils'
import { CodeLanguage, Sandbox } from '@daytonaio/sdk'
import { ChevronUpIcon, Loader2, PanelBottom, Play, XIcon } from 'lucide-react'
import { ReactNode, useMemo, useState } from 'react'
import { Group, Panel, usePanelRef } from 'react-resizable-panels'
import ResponseCard from '../ResponseCard'
import { Window, WindowContent, WindowTitleBar } from '../Window'
import { codeSnippetGenerators, CodeSnippetParams } from './CodeSnippets'

const codeSnippetSupportedLanguages = [
  { value: CodeLanguage.PYTHON, label: 'Python', icon: PythonIcon },
  { value: CodeLanguage.TYPESCRIPT, label: 'TypeScript', icon: TypescriptIcon },
]

const SandboxCodeSnippetsResponse = ({ className }: { className?: string }) => {
  const [codeSnippetLanguage, setCodeSnippetLanguage] = useState<CodeLanguage>(CodeLanguage.PYTHON)
  const [codeSnippetOutput, setCodeSnippetOutput] = useState<string | ReactNode>('')
  const [isCodeSnippetRunning, setIsCodeSnippetRunning] = useState<boolean>(false)

  const { sandboxParametersState, actionRuntimeError, getSandboxParametersInfo } = usePlayground()
  const {
    sandbox: { create: createSandbox },
  } = usePlaygroundSandbox()

  const useConfigObject = false // Currently not needed, we use jwtToken for client config

  const fileSystemListFilesLocationSet = !actionRuntimeError[FileSystemActions.LIST_FILES]
  const fileSystemCreateFolderParamsSet = !actionRuntimeError[FileSystemActions.CREATE_FOLDER]
  const fileSystemDeleteFileRequiredParamsSet = !actionRuntimeError[FileSystemActions.DELETE_FILE]
  const useFileSystemDeleteFileRecursive =
    fileSystemDeleteFileRequiredParamsSet && sandboxParametersState['deleteFileParams'].recursive === true
  const shellCommandExists = !actionRuntimeError[ProcessCodeExecutionActions.SHELL_COMMANDS_RUN]
  const codeToRunExists = !actionRuntimeError[ProcessCodeExecutionActions.CODE_RUN]
  const gitCloneOperationRequiredParamsSet = !actionRuntimeError[GitOperationsActions.GIT_CLONE]
  const useGitCloneBranch = !!sandboxParametersState['gitCloneParams'].branchToClone
  const useGitCloneCommitId = !!sandboxParametersState['gitCloneParams'].commitToClone
  const useGitCloneUsername = !!sandboxParametersState['gitCloneParams'].authUsername
  const useGitClonePassword = !!sandboxParametersState['gitCloneParams'].authPassword
  const gitStatusOperationLocationSet = !actionRuntimeError[GitOperationsActions.GIT_STATUS]
  const gitBranchesOperationLocationSet = !actionRuntimeError[GitOperationsActions.GIT_BRANCHES_LIST]

  const codeSnippetParams = useMemo<CodeSnippetParams>(
    () => ({
      state: sandboxParametersState,
      config: getSandboxParametersInfo(),
      actions: {
        useConfigObject,
        fileSystemListFilesLocationSet,
        fileSystemCreateFolderParamsSet,
        fileSystemDeleteFileRequiredParamsSet,
        useFileSystemDeleteFileRecursive,
        shellCommandExists,
        codeToRunExists,
        gitCloneOperationRequiredParamsSet,
        useGitCloneBranch,
        useGitCloneCommitId,
        useGitCloneUsername,
        useGitClonePassword,
        gitStatusOperationLocationSet,
        gitBranchesOperationLocationSet,
      },
    }),
    [
      sandboxParametersState,
      getSandboxParametersInfo,
      useConfigObject,
      fileSystemListFilesLocationSet,
      fileSystemCreateFolderParamsSet,
      fileSystemDeleteFileRequiredParamsSet,
      useFileSystemDeleteFileRecursive,
      shellCommandExists,
      codeToRunExists,
      gitCloneOperationRequiredParamsSet,
      useGitCloneBranch,
      useGitCloneCommitId,
      useGitCloneUsername,
      useGitClonePassword,
      gitStatusOperationLocationSet,
      gitBranchesOperationLocationSet,
    ],
  )

  const sandboxCodeSnippetsData = useMemo(
    () => ({
      [CodeLanguage.PYTHON]: { code: codeSnippetGenerators[CodeLanguage.PYTHON].buildFullSnippet(codeSnippetParams) },
      [CodeLanguage.TYPESCRIPT]: {
        code: codeSnippetGenerators[CodeLanguage.TYPESCRIPT].buildFullSnippet(codeSnippetParams),
      },
    }),
    [codeSnippetParams],
  )

  const runCodeSnippet = async () => {
    setIsCodeSnippetRunning(true)
    let codeSnippetOutput = 'Creating sandbox...\n'
    setCodeSnippetOutput(codeSnippetOutput)
    let sandbox: Sandbox | undefined

    try {
      sandbox = await createSandbox()
      codeSnippetOutput = `Sandbox successfully created: ${sandbox.id}\n`
      setCodeSnippetOutput(codeSnippetOutput)
      if (codeToRunExists) {
        setCodeSnippetOutput(codeSnippetOutput + '\nRunning code...')
        const codeRunResponse = await sandbox.process.codeRun(
          sandboxParametersState['codeRunParams'].languageCode as string,
        ) // codeToRunExists guarantees that value isn't undefined so we put as string to silence TS compiler
        codeSnippetOutput += `\nCode run result: ${codeRunResponse.result}`
        setCodeSnippetOutput(codeSnippetOutput)
      }
      if (shellCommandExists) {
        setCodeSnippetOutput(codeSnippetOutput + '\nRunning shell command...')
        const shellCommandResponse = await sandbox.process.executeCommand(
          sandboxParametersState['shellCommandRunParams'].shellCommand as string, // shellCommandExists guarantees that value isn't undefined so we put as string to silence TS compiler
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
                    <img src={language.icon} alt={`${language.label} icon`} className="w-4 h-4" />
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
