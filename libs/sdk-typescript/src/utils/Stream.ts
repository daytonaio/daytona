/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Process a streaming response from a URL. Stream will terminate if the server-side stream
 * ends or if the shouldTerminate function returns True.
 *
 * @param getStream - A function that returns a promise of an AxiosResponse with .data being the stream
 * @param onChunk - A function to process each chunk of the response
 * @param shouldTerminate - A function to check if the response should be terminated
 * @param chunkTimeout - The timeout for each chunk
 * @param requireConsecutiveTermination - Whether to require two consecutive termination signals
 * to terminate the stream.
 */
export async function processStreamingResponse(
  getStream: () => Promise<any>, // can return AxiosResponse with .data being the stream
  onChunk: (chunk: string) => void,
  shouldTerminate: () => Promise<boolean>,
  chunkTimeout = 2000,
  requireConsecutiveTermination = true,
): Promise<void> {
  const response = await getStream()
  const stream = response.data

  let nextChunkPromise: Promise<Buffer | null> | null = null
  let exitCheckStreak = 0
  let terminated = false

  const readNext = (): Promise<Buffer | null> => {
    return new Promise((resolve) => {
      const onData = (data: Buffer) => {
        cleanup()
        resolve(data)
      }
      const cleanup = () => {
        stream.off('data', onData)
      }
      stream.once('data', onData)
    })
  }

  const terminationPromise = new Promise<void>((resolve, reject) => {
    stream.on('end', () => {
      terminated = true
      resolve()
    })
    stream.on('close', () => {
      terminated = true
      resolve()
    })
    stream.on('error', (err: Error) => {
      terminated = true
      reject(err)
    })
  })

  const processLoop = async () => {
    while (!terminated) {
      if (!nextChunkPromise) {
        nextChunkPromise = readNext()
      }

      const timeoutPromise = new Promise<null>((resolve) => setTimeout(() => resolve(null), chunkTimeout))
      const result = await Promise.race([nextChunkPromise, timeoutPromise])

      if (result instanceof Buffer) {
        onChunk(result.toString('utf8'))
        nextChunkPromise = null
        exitCheckStreak = 0
      } else {
        const shouldEnd = await shouldTerminate()
        if (shouldEnd) {
          exitCheckStreak += 1
          if (!requireConsecutiveTermination || exitCheckStreak > 1) {
            break
          }
        } else {
          exitCheckStreak = 0
        }
      }
    }
    stream.destroy()
    stream.removeAllListeners()
  }

  await Promise.race([processLoop(), terminationPromise])
}
