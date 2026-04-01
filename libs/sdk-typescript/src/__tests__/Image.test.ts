const mockDynamicRequire = jest.fn()

jest.mock('../utils/Import', () => ({
  dynamicRequire: (...args: unknown[]) => mockDynamicRequire(...args),
}))

describe('Image', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('builds images using static factories', async () => {
    const { Image } = await import('../Image')

    const base = Image.base('debian:12')
    expect(base.dockerfile).toContain('FROM debian:12')

    const slim = Image.debianSlim('3.12')
    expect(slim.dockerfile).toContain('FROM python:3.12.10-slim-bookworm')
  })

  it('supports dockerfile mutation methods', async () => {
    const { Image } = await import('../Image')

    const image = Image.base('python:3.12')
      .pipInstall(['numpy', 'pandas'])
      .runCommands('echo hi', ['bash', '-lc', 'pwd'])
      .env({ NODE_ENV: 'test' })
      .workdir('/workspace')
      .entrypoint(['/bin/bash'])
      .cmd(['-lc', 'echo ready'])

    expect(image.dockerfile).toContain('pip install')
    expect(image.dockerfile).toContain('RUN echo hi')
    expect(image.dockerfile).toContain('ENV NODE_ENV')
    expect(image.dockerfile).toContain('WORKDIR')
    expect(image.dockerfile).toContain('ENTRYPOINT')
    expect(image.dockerfile).toContain('CMD')
  })

  it('validates entrypoint/cmd/env input', async () => {
    const { Image } = await import('../Image')
    expect(() => Image.base('x').entrypoint([1] as unknown as string[])).toThrow(
      'entrypoint_commands must be a list of strings',
    )
    expect(() => Image.base('x').cmd([1] as unknown as string[])).toThrow('Image CMD must be a list of strings')
    expect(() => Image.base('x').env({ A: 1 as unknown as string })).toThrow('Image ENV variables must be strings')
  })

  it('supports fromDockerfile and local context additions', async () => {
    const { Image } = await import('../Image')

    const fsModule = {
      existsSync: jest.fn(() => true),
      readFileSync: jest.fn(() => 'FROM debian:12\nCOPY ./src /app/src\n'),
      statSync: jest.fn(() => ({ isDirectory: () => true })),
    }
    const expandTilde = (value: string) => value
    const fastGlob = { sync: jest.fn(() => ['/repo/src']) }

    mockDynamicRequire.mockImplementation((moduleName: string) => {
      if (moduleName === 'fs') return fsModule
      if (moduleName === 'expand-tilde') return expandTilde
      if (moduleName === 'fast-glob') return fastGlob
      return {}
    })

    const image = Image.fromDockerfile('/repo/Dockerfile')
      .addLocalFile('/repo/.env', '/workspace/')
      .addLocalDir('/repo/src', '/workspace/src')
      .dockerfileCommands(['COPY ./assets /app/assets'], '/repo')

    expect(image.contextList.length).toBeGreaterThan(0)
    expect(image.dockerfile).toContain('COPY')
  })

  it('extractCopySources and parseCopyCommand handle copy variants', async () => {
    const { Image } = await import('../Image')
    const fastGlob = { sync: jest.fn(() => ['/repo/a.txt']) }
    mockDynamicRequire.mockImplementation((moduleName: string) => {
      if (moduleName === 'fast-glob') return fastGlob
      return {}
    })

    const imageRuntime = Image as unknown as Record<string, (...args: unknown[]) => unknown>

    const parsed = imageRuntime.parseCopyCommand('COPY --chown=user:group ./a.txt /app/a.txt') as {
      sources: string[]
      dest: string
    }
    expect(parsed.sources).toEqual(['./a.txt'])
    expect(parsed.dest).toBe('/app/a.txt')

    const sources = imageRuntime.extractCopySources('COPY ./a.txt /app/a.txt', '/repo') as Array<[string, string]>
    expect(sources[0]).toEqual(['/repo/a.txt', './a.txt'])
  })
})
