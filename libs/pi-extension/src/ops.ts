/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Daytona-backed implementations of Pi's pluggable tool operations.
 *
 * The `bash`, `read`, `write`, `edit`, and `ls` tools accept an `*Operations`
 * object. We provide versions that run against a remote Daytona sandbox instead
 * of the local machine. The tool factories are given the sandbox working
 * directory as their `cwd`, so the absolute paths Pi resolves and hands to these
 * ops are already sandbox-rooted.
 *
 * `grep` and `find` are NOT here: Pi runs ripgrep/fd locally and their
 * operations don't delegate the actual search (grep's ops only read context
 * lines; find's glob can't be expressed via Daytona's basename-only
 * searchFiles). Those run inside the sandbox via `grep-tool.ts` / `find-tool.ts`.
 */

import type { Sandbox } from '@daytona/sdk'
import type {
  BashOperations,
  EditOperations,
  LsOperations,
  ReadOperations,
  WriteOperations,
} from '@earendil-works/pi-coding-agent'
import { execCommand, withRecovery } from './sandbox.ts'
import { shellQuote } from './util.ts'

const IMAGE_MIME_TYPES = new Set(['image/jpeg', 'image/png', 'image/gif', 'image/webp'])

/** Run a command in the sandbox and return its combined stdout and exit code. */
async function run(sandbox: Sandbox, command: string): Promise<{ stdout: string; exitCode: number }> {
  const res = await execCommand(sandbox, command)
  const stdout = res.result ?? res.artifacts?.stdout ?? ''
  return { stdout, exitCode: res.exitCode ?? 0 }
}

/**
 * Wrap a command so backgrounded processes can't hang the call.
 *
 * Daytona's `executeCommand` resolves only when the command's stdout/stderr
 * reach EOF. A backgrounded process (`server &`) inherits those pipes and holds
 * them open indefinitely, so the call never returns. We run the command in a
 * subshell whose combined output is redirected to a temp file, then replay the
 * file and re-raise the real exit code. Background descendants then write to the
 * file (not the result pipe), so the call returns as soon as the FOREGROUND
 * finishes — matching Pi's local "background and return" behavior. A subshell
 * (not a brace group) keeps any `exit` inside the user command from skipping the
 * replay. The newline before `)` correctly terminates a trailing `&`.
 */
function backgroundSafe(command: string): string {
  return [
    '__pi_out=$(mktemp 2>/dev/null || echo "/tmp/pi-daytona-$$.out")',
    `( ${command}`,
    ') >"$__pi_out" 2>&1',
    '__pi_rc=$?',
    'cat "$__pi_out"',
    'rm -f "$__pi_out"',
    'exit $__pi_rc',
  ].join('\n')
}

/**
 * Build bash operations backed by the sandbox.
 *
 * `cwdOverride` forces the working directory regardless of what the caller
 * passes. The agent's bash tool resolves sandbox-rooted paths and omits it;
 * the user `!` handler passes the sandbox cwd (see the user_bash handler in
 * tools.ts for why).
 */
export function createBashOps(sandbox: Sandbox, cwdOverride?: string): BashOperations {
  return {
    // Daytona's executeCommand is non-streaming, so we emit the whole output
    // once when it resolves. We wrap the command (see backgroundSafe) so a
    // backgrounded process like `python3 -m http.server 8080 &` returns
    // immediately instead of hanging on the inherited output pipe.
    exec: async (command, cwd, { onData, signal, timeout }) => {
      // We can only honor a pre-aborted signal: Daytona's executeCommand is a
      // single blocking call with no cancellation, so once it's running there's
      // no way to abort mid-flight (racing a timer would just stop awaiting
      // while the sandbox command keeps running). `timeout` bounds the worst case.
      if (signal?.aborted) throw new Error('aborted')
      // We deliberately do not forward the host `env` into the sandbox: the
      // container has its own environment, and leaking host vars is unsafe.
      const res = await execCommand(sandbox, backgroundSafe(command), cwdOverride ?? cwd, timeout)
      const output = res.result ?? res.artifacts?.stdout ?? ''
      if (output) onData(Buffer.from(output))
      return { exitCode: res.exitCode ?? null }
    },
  }
}

export function createReadOps(sandbox: Sandbox): ReadOperations {
  return {
    readFile: (path) => withRecovery(sandbox, () => sandbox.fs.downloadFile(path)),
    access: async (path) => {
      const { exitCode } = await run(sandbox, `test -r ${shellQuote(path)}`)
      if (exitCode !== 0) throw new Error(`File not readable: ${path}`)
    },
    detectImageMimeType: async (path) => {
      try {
        const { stdout } = await run(sandbox, `file --mime-type -b ${shellQuote(path)}`)
        const mime = stdout.trim()
        return IMAGE_MIME_TYPES.has(mime) ? mime : null
      } catch {
        return null
      }
    },
  }
}

export function createWriteOps(sandbox: Sandbox): WriteOperations {
  return {
    writeFile: (path, content) =>
      withRecovery(sandbox, () => sandbox.fs.uploadFile(Buffer.from(content, 'utf8'), path)),
    // `mkdir -p` is idempotent; fs.createFolder errors if the folder exists.
    mkdir: async (dir) => {
      const { exitCode } = await run(sandbox, `mkdir -p ${shellQuote(dir)}`)
      if (exitCode !== 0) throw new Error(`Failed to create directory: ${dir}`)
    },
  }
}

export function createEditOps(sandbox: Sandbox): EditOperations {
  // Pi's edit tool reads the file, applies exact oldText->newText edits
  // in-process, then writes it back — preserving its uniqueness checks.
  // This is the download -> modify -> upload strategy.
  return {
    readFile: (path) => withRecovery(sandbox, () => sandbox.fs.downloadFile(path)),
    writeFile: (path, content) =>
      withRecovery(sandbox, () => sandbox.fs.uploadFile(Buffer.from(content, 'utf8'), path)),
    access: async (path) => {
      const { exitCode } = await run(sandbox, `test -r ${shellQuote(path)} && test -w ${shellQuote(path)}`)
      if (exitCode !== 0) throw new Error(`File not readable/writable: ${path}`)
    },
  }
}

export function createLsOps(sandbox: Sandbox): LsOperations {
  return {
    exists: async (path) => {
      const { exitCode } = await run(sandbox, `test -e ${shellQuote(path)}`)
      return exitCode === 0
    },
    stat: async (path) => {
      const q = shellQuote(path)
      // `test -e || exit 1` makes a missing path a real non-zero exit; the
      // `|| echo other` only masks the `test -d` result, which is fine since
      // existence is already decided.
      const { stdout, exitCode } = await run(sandbox, `test -e ${q} || exit 1; test -d ${q} && echo dir || echo other`)
      if (exitCode !== 0) throw new Error(`Path not found: ${path}`)
      const isDir = stdout.trim() === 'dir'
      return { isDirectory: () => isDir }
    },
    readdir: async (path) => {
      const { stdout, exitCode } = await run(sandbox, `ls -1A ${shellQuote(path)}`)
      if (exitCode !== 0) throw new Error(`Failed to read directory: ${path}`)
      return stdout.split('\n').filter((line) => line.length > 0)
    },
  }
}
