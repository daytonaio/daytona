import { daytona_typescript, hello_world_typescript } from './typescript'
import { daytona_python, hello_world_python } from './python'
import { hello_world_bash } from './bash'

export enum CodeLanguage {
  TypeScript = 'typescript',
  Python = 'python',
  Bash = 'bash',
}

export const SAMPLES: Record<CodeLanguage, Record<'Default' | string, string>> = {
  [CodeLanguage.TypeScript]: {
    Default: daytona_typescript,
    'Hello World': hello_world_typescript,
  },
  [CodeLanguage.Python]: {
    Default: daytona_python,
    'Hello World': hello_world_python,
  },
  [CodeLanguage.Bash]: {
    Default: hello_world_bash,
  },
}

export * from './python'
export * from './typescript'
export * from './bash'
