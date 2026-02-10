/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxParams, SandboxParametersInfo } from '@/contexts/PlaygroundContext'

export interface CodeSnippetActionFlags {
  useConfigObject: boolean
  fileSystemListFilesLocationSet: boolean
  fileSystemCreateFolderParamsSet: boolean
  fileSystemDeleteFileRequiredParamsSet: boolean
  useFileSystemDeleteFileRecursive: boolean
  shellCommandExists: boolean
  codeToRunExists: boolean
  gitCloneOperationRequiredParamsSet: boolean
  useGitCloneBranch: boolean
  useGitCloneCommitId: boolean
  useGitCloneUsername: boolean
  useGitClonePassword: boolean
  gitStatusOperationLocationSet: boolean
  gitBranchesOperationLocationSet: boolean
}

export interface CodeSnippetParams {
  state: SandboxParams
  config: SandboxParametersInfo
  actions: CodeSnippetActionFlags
}

export interface CodeSnippetGenerator {
  getImports(p: CodeSnippetParams): string
  getConfig(p: CodeSnippetParams): string
  getClientInit(p: CodeSnippetParams): string
  getResources(p: CodeSnippetParams): string
  getSandboxParams(p: CodeSnippetParams): string
  getSandboxCreate(p: CodeSnippetParams): string
  getCodeRun(p: CodeSnippetParams): string
  getShellRun(p: CodeSnippetParams): string
  getFileSystemOps(p: CodeSnippetParams): string
  getGitOps(p: CodeSnippetParams): string
  buildFullSnippet(p: CodeSnippetParams): string
}
