// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

export const createApiResponse = <T>(data: T): { data: T } => ({ data })

export const flushPromises = async (): Promise<void> => {
  await Promise.resolve()
}

export const toMuxedOutput = (stdout: string, stderr: string): string => {
  const stdoutPrefix = new Uint8Array([0x01, 0x01, 0x01])
  const stderrPrefix = new Uint8Array([0x02, 0x02, 0x02])
  const encoder = new TextEncoder()

  const out = encoder.encode(stdout)
  const err = encoder.encode(stderr)

  const bytes = new Uint8Array(stdoutPrefix.length + out.length + stderrPrefix.length + err.length)
  bytes.set(stdoutPrefix, 0)
  bytes.set(out, stdoutPrefix.length)
  bytes.set(stderrPrefix, stdoutPrefix.length + out.length)
  bytes.set(err, stdoutPrefix.length + out.length + stderrPrefix.length)

  return new TextDecoder().decode(bytes)
}
