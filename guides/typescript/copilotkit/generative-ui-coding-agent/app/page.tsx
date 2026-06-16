'use client'

import {
  CopilotChat,
  useConfigureSuggestions,
  useRenderTool,
} from '@copilotkit/react-core/v2'
import { z } from 'zod'
import { SandboxCard } from '@/components/SandboxCard'
import { TerminalCard } from '@/components/TerminalCard'
import { FileCard } from '@/components/FileCard'
import { FileListCard } from '@/components/FileListCard'
import { GrepCard } from '@/components/GrepCard'
import { ReplaceCard } from '@/components/ReplaceCard'
import { FileInfoCard } from '@/components/FileInfoCard'
import { PreviewCard } from '@/components/PreviewCard'

function parseResult<T>(result: unknown): T | undefined {
  if (typeof result !== 'string' || result.length === 0) return undefined
  try {
    return JSON.parse(result) as T
  } catch {
    return undefined
  }
}

const createSandboxParams = z.object({
  envVars: z.record(z.string()).optional(),
  labels: z.record(z.string()).optional(),
  autoStopInterval: z.number().optional(),
})
const runCommandParams = z.object({
  sandboxId: z.string(),
  command: z.string(),
  background: z.boolean().optional(),
})
const writeFileParams = z.object({
  sandboxId: z.string(),
  path: z.string(),
  content: z.string(),
})
const readFileParams = z.object({
  sandboxId: z.string(),
  path: z.string(),
})
const listFilesParams = z.object({
  sandboxId: z.string(),
  path: z.string(),
})
const findFilesParams = z.object({
  sandboxId: z.string(),
  path: z.string(),
  pattern: z.string(),
})
const searchFilesParams = z.object({
  sandboxId: z.string(),
  path: z.string(),
  pattern: z.string(),
})
const replaceInFilesParams = z.object({
  sandboxId: z.string(),
  files: z.array(z.string()),
  pattern: z.string(),
  newValue: z.string(),
})
const getFileDetailsParams = z.object({
  sandboxId: z.string(),
  path: z.string(),
})
const startWebServerParams = z.object({
  sandboxId: z.string(),
  command: z.string(),
  port: z.number(),
})
const getPreviewUrlParams = z.object({
  sandboxId: z.string(),
  port: z.number(),
})

type CreateSandboxResult = { sandboxId: string }
type RunCommandResult = {
  exitCode?: number
  stdout?: string
  command?: string
  background?: boolean
  sessionId?: string
  cmdId?: string
}
type WriteFileResult = { path: string; bytesWritten: number }
type ReadFileResult = { path: string; content: string; bytes: number }
type ListFilesResult = {
  path: string
  entries: Array<{ name: string; isDir: boolean; size: number; permissions: string }>
}
type FindFilesResult = {
  pattern: string
  matches: Array<{ file: string; line: number; content: string }>
}
type SearchFilesResult = { pattern: string; files: string[] }
type ReplaceInFilesResult = {
  pattern: string
  newValue: string
  results: Array<{ file?: string; success?: boolean; error?: string }>
}
type GetFileDetailsResult = {
  path: string
  name?: string
  isDir?: boolean
  size?: number
  mode?: string
  permissions?: string
  owner?: string
  group?: string
  modifiedAt?: string
}
type StartWebServerResult = {
  sandboxId: string
  port: number
  url: string
  sessionId?: string
  cmdId?: string
  command?: string
}
type GetPreviewUrlResult = { url: string; port: number }

export default function Page() {
  useRenderTool({
    name: 'createSandbox',
    parameters: createSandboxParams,
    render: ({ status, parameters, result }) => {
      const r = parseResult<CreateSandboxResult>(result)
      return (
        <SandboxCard
          status={status}
          sandboxId={r?.sandboxId}
          envVars={
            parameters?.envVars
              ? Object.fromEntries(
                  Object.keys(parameters.envVars).map((key) => [key, '[redacted]']),
                )
              : undefined
          }
          labels={parameters?.labels}
          autoStopInterval={parameters?.autoStopInterval}
        />
      )
    },
  })

  useRenderTool({
    name: 'runCommand',
    parameters: runCommandParams,
    render: ({ status, parameters, result }) => {
      const r = parseResult<RunCommandResult>(result)
      return (
        <TerminalCard
          status={status}
          command={parameters?.command}
          background={parameters?.background || r?.background}
          stdout={r?.stdout}
          exitCode={r?.exitCode}
          sessionId={r?.sessionId}
          cmdId={r?.cmdId}
        />
      )
    },
  })

  useRenderTool({
    name: 'writeFile',
    parameters: writeFileParams,
    render: ({ status, parameters, result }) => {
      const r = parseResult<WriteFileResult>(result)
      return (
        <FileCard
          status={status}
          verb="wrote"
          path={parameters?.path}
          content={parameters?.content}
          bytes={r?.bytesWritten}
        />
      )
    },
  })

  useRenderTool({
    name: 'readFile',
    parameters: readFileParams,
    render: ({ status, parameters, result }) => {
      const r = parseResult<ReadFileResult>(result)
      return (
        <FileCard
          status={status}
          verb="read"
          path={parameters?.path}
          content={r?.content}
          bytes={r?.bytes}
        />
      )
    },
  })

  useRenderTool({
    name: 'listFiles',
    parameters: listFilesParams,
    render: ({ status, parameters, result }) => {
      const r = parseResult<ListFilesResult>(result)
      return (
        <FileListCard
          status={status}
          verb="listed"
          path={parameters?.path}
          entries={r?.entries}
        />
      )
    },
  })

  useRenderTool({
    name: 'findFiles',
    parameters: findFilesParams,
    render: ({ status, parameters, result }) => {
      const r = parseResult<FindFilesResult>(result)
      return (
        <GrepCard
          status={status}
          path={parameters?.path}
          pattern={parameters?.pattern}
          matches={r?.matches}
        />
      )
    },
  })

  useRenderTool({
    name: 'searchFiles',
    parameters: searchFilesParams,
    render: ({ status, parameters, result }) => {
      const r = parseResult<SearchFilesResult>(result)
      return (
        <FileListCard
          status={status}
          verb="searched"
          path={parameters?.path}
          pattern={parameters?.pattern}
          files={r?.files}
        />
      )
    },
  })

  useRenderTool({
    name: 'replaceInFiles',
    parameters: replaceInFilesParams,
    render: ({ status, parameters, result }) => {
      const r = parseResult<ReplaceInFilesResult>(result)
      return (
        <ReplaceCard
          status={status}
          pattern={parameters?.pattern}
          newValue={parameters?.newValue}
          results={r?.results}
        />
      )
    },
  })

  useRenderTool({
    name: 'getFileDetails',
    parameters: getFileDetailsParams,
    render: ({ status, parameters, result }) => {
      const r = parseResult<GetFileDetailsResult>(result)
      return (
        <FileInfoCard
          status={status}
          info={r ? r : parameters?.path ? { path: parameters.path } : undefined}
        />
      )
    },
  })

  useRenderTool({
    name: 'startWebServer',
    parameters: startWebServerParams,
    render: ({ status, parameters, result }) => {
      const r = parseResult<StartWebServerResult>(result)
      return <PreviewCard status={status} url={r?.url} port={parameters?.port} />
    },
  })

  useRenderTool({
    name: 'getPreviewUrl',
    parameters: getPreviewUrlParams,
    render: ({ status, parameters, result }) => {
      const r = parseResult<GetPreviewUrlResult>(result)
      return <PreviewCard status={status} url={r?.url} port={parameters?.port} />
    },
  })

  useConfigureSuggestions({
    instructions:
      'Suggest 3 short, varied prompts a developer might ask a coding agent with shell access. Mix app-building requests with debugging, scripting, or data-analysis tasks (e.g. "Build a todo app", "Find the bug in this Python script", "Generate a CSV of prime numbers under 1000").',
    minSuggestions: 3,
    maxSuggestions: 3,
    available: 'always',
  })

  return (
    <main style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
      <header
        style={{
          padding: '14px 20px',
          borderBottom: '1px solid var(--border)',
          background: '#fff',
        }}
      >
        <h1 style={{ margin: 0, fontSize: 16, fontWeight: 600 }}>
          CopilotKit + Daytona Coding Agent
        </h1>
        <p style={{ margin: '4px 0 0', fontSize: 13, color: '#64748b' }}>
          The agent has access to a Daytona sandbox. Ask it to build, debug, run, or analyze code.
        </p>
      </header>
      <div style={{ flex: 1, minHeight: 0 }}>
        <CopilotChat
          labels={{
            welcomeMessageText: 'What should we work on today?',
          }}
        />
      </div>
    </main>
  )
}
