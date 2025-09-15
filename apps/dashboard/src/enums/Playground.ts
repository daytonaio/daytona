/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import PythonIcon from '@/assets/python.svg'
import TypescriptIcon from '@/assets/typescript.svg'
import { CodeLanguage } from '@daytonaio/sdk-typescript/src'

export enum PlaygroundCategories {
  SANDBOX = 'sandbox',
  TERMINAL = 'terminal',
  VNC = 'vnc',
}

export const playgroundCategoriesData = [
  { value: PlaygroundCategories.SANDBOX, label: 'Sandbox' },
  { value: PlaygroundCategories.TERMINAL, label: 'Terminal' },
  { value: PlaygroundCategories.VNC, label: 'VNC' },
]

export enum SandboxParametersSections {
  SANDBOX_MANAGMENT = 'sandbox_managment',
  AGENT_TOOLBOX = 'agent_toolbox',
  FILE_SYSTEM = 'file_system',
  GIT_OPERATIONS = 'git_operations',
  PROCESS_CODE_EXECUTION = 'process_code_execution',
}

export const sandboxParametersSectionsData = [
  { value: 'sandbox_managment', label: 'Managment' },
  { value: 'agent_toolbox', label: 'Agent Toolbox' },
  { value: 'file_system', label: 'File System' },
  { value: 'git_operations', label: 'Git Operations' },
  { value: 'process_code_execution', label: 'Process & Code Execution' },
]

export const codeSnippetSupportedLanguages = [
  { value: CodeLanguage.PYTHON, label: 'Python', icon: PythonIcon },
  { value: CodeLanguage.TYPESCRIPT, label: 'TypeScript', icon: TypescriptIcon },
]
