/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CodeLanguage } from '@daytonaio/sdk'
import { PythonSnippetGenerator } from './python'
import { CodeSnippetGenerator } from './types'
import { TypeScriptSnippetGenerator } from './typescript'

export const codeSnippetGenerators: Record<Exclude<CodeLanguage, CodeLanguage.JAVASCRIPT>, CodeSnippetGenerator> = {
  [CodeLanguage.PYTHON]: PythonSnippetGenerator,
  [CodeLanguage.TYPESCRIPT]: TypeScriptSnippetGenerator,
}

export type { CodeSnippetActionFlags, CodeSnippetGenerator, CodeSnippetParams } from './types'
