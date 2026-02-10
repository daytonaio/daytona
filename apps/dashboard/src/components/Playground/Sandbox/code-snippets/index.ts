/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CodeLanguage } from '@daytonaio/sdk'
import { CodeSnippetGenerator } from './types'
import { PythonSnippetGenerator } from './python'
import { TypeScriptSnippetGenerator } from './typescript'

export const codeSnippetGenerators: Record<CodeLanguage, CodeSnippetGenerator> = {
  [CodeLanguage.PYTHON]: PythonSnippetGenerator,
  [CodeLanguage.TYPESCRIPT]: TypeScriptSnippetGenerator,
}

export type { CodeSnippetGenerator, CodeSnippetActionFlags, CodeSnippetParams } from './types'
