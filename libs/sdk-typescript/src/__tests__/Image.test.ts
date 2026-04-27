// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

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

  it('uses the latest supported python version by default', async () => {
    const { Image } = await import('../Image')

    const slim = Image.debianSlim()
    expect(slim.dockerfile).toContain('FROM python:3.13.3-slim-bookworm')
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

  it('formats pip install options and sorts package names', async () => {
    const { Image } = await import('../Image')

    const image = Image.base('python:3.12').pipInstall(['pandas', 'numpy'], {
      findLinks: ['https://links.example.com'],
      indexUrl: 'https://index.example.com/simple',
      extraIndexUrls: ['https://extra.example.com/simple'],
      pre: true,
      extraOptions: '  --no-cache-dir  ',
    })

    expect(image.dockerfile).toContain('pip install numpy pandas')
    expect(image.dockerfile).toContain('--find-links')
    expect(image.dockerfile).toContain('--index-url')
    expect(image.dockerfile).toContain('--extra-index-url')
    expect(image.dockerfile).toContain('--pre')
    expect(image.dockerfile).toContain('--no-cache-dir')
  })

  it('does not modify the dockerfile when pipInstall receives no packages', async () => {
    const { Image } = await import('../Image')

    const image = Image.base('python:3.12')
    const before = image.dockerfile
    image.pipInstall([])

    expect(image.dockerfile).toBe(before)
  })

  it('validates entrypoint/cmd/env input', async () => {
    const { Image } = await import('../Image')
    expect(() => Image.base('x').entrypoint([1] as unknown as string[])).toThrow(
      'entrypoint_commands must be a list of strings',
    )
    expect(() => Image.base('x').cmd([1] as unknown as string[])).toThrow('Image CMD must be a list of strings')
    expect(() => Image.base('x').env({ A: 1 as unknown as string })).toThrow('Image ENV variables must be strings')
  })

  it('validates unsupported python versions', async () => {
    const { Image } = await import('../Image')

    expect(() => Image.debianSlim('3.8' as never)).toThrow('Unsupported Python version: 3.8')
  })

  it('validates nested non-string pipInstall arguments', async () => {
    const { Image } = await import('../Image')
    const runtime = Image.base('python:3.12') as unknown as { pipInstall: (packages: unknown) => unknown }
    const badPackages: unknown = ['numpy', [123]]

    expect(() => runtime.pipInstall(badPackages)).toThrow('pipInstall: packages must only contain strings')
  })

  it('validates dockerfile entrypoint command arrays', async () => {
    const { Image } = await import('../Image')

    expect(() => Image.base('x').entrypoint('bash' as never)).toThrow('entrypoint_commands must be a list of strings')
  })

  it('supports fromDockerfile and local context additions', async () => {
    const { Image } = await import('../Image')

    const fsModule = {
      existsSync: jest.fn(() => true),
      readFileSync: jest.fn(() => 'FROM debian:12\nCOPY ./src /app/src\n'),
      statSync: jest.fn(() => ({ isDirectory: () => true, isFile: () => true })),
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

  it('throws when fromDockerfile path does not exist', async () => {
    const { Image } = await import('../Image')

    const fsModule = {
      existsSync: jest.fn(() => false),
      statSync: jest.fn(),
      readFileSync: jest.fn(),
    }

    mockDynamicRequire.mockImplementation((moduleName: string) => {
      if (moduleName === 'fs') return fsModule
      if (moduleName === 'expand-tilde') return (value: string) => value
      return {}
    })

    expect(() => Image.fromDockerfile('/repo/Dockerfile')).toThrow('Dockerfile /repo/Dockerfile does not exist')
  })

  it('throws when local file or directory inputs are invalid', async () => {
    const { Image } = await import('../Image')

    const fsModule = {
      existsSync: jest.fn((path: string) => path !== '/missing'),
      statSync: jest.fn((path: string) => ({
        isDirectory: () => path === '/repo/file.txt',
        isFile: () => path === '/repo/dir',
      })),
    }

    mockDynamicRequire.mockImplementation((moduleName: string) => {
      if (moduleName === 'fs') return fsModule
      if (moduleName === 'expand-tilde') return (value: string) => value
      return {}
    })

    expect(() => Image.base('x').addLocalFile('/missing', '/tmp/')).toThrow('Local file /missing does not exist')
    expect(() => Image.base('x').addLocalFile('/repo/file.txt', '/tmp/')).toThrow(
      'Local path /repo/file.txt exists but is not a file',
    )
    expect(() => Image.base('x').addLocalDir('/repo/dir', '/tmp/dir')).toThrow(
      'Local path /repo/dir exists but is not a directory',
    )
  })

  it('validates dockerfileCommands context directories', async () => {
    const { Image } = await import('../Image')

    const fsModule = {
      existsSync: jest.fn((path: string) => path !== '/missing'),
      statSync: jest.fn((path: string) => ({
        isDirectory: () => path === '/repo',
      })),
    }

    mockDynamicRequire.mockImplementation((moduleName: string) => {
      if (moduleName === 'fs') return fsModule
      if (moduleName === 'expand-tilde') return (value: string) => value
      if (moduleName === 'fast-glob') return { sync: jest.fn(() => []) }
      return {}
    })

    expect(() => Image.base('x').dockerfileCommands(['COPY ./src /app/src'], '/missing')).toThrow(
      'Context directory /missing does not exist',
    )
    expect(() => Image.base('x').dockerfileCommands(['COPY ./src /app/src'], '/file')).toThrow(
      'Context path /file exists but is not a directory',
    )
  })

  it('supports requirements files and pyproject dependencies', async () => {
    const { Image } = await import('../Image')

    const fsModule = {
      existsSync: jest.fn(() => true),
      statSync: jest.fn(() => ({ isFile: () => true })),
      readFileSync: jest.fn(() =>
        [
          '[project]',
          'dependencies = ["requests", "numpy"]',
          '[project.optional-dependencies]',
          'dev = ["pytest"]',
        ].join('\n'),
      ),
    }

    mockDynamicRequire.mockImplementation((moduleName: string) => {
      if (moduleName === 'fs') return fsModule
      if (moduleName === 'expand-tilde') return (value: string) => value
      if (moduleName === '@iarna/toml')
        return {
          parse: jest.fn(() => ({
            project: { dependencies: ['requests', 'numpy'], 'optional-dependencies': { dev: ['pytest'] } },
          })),
        }
      return {}
    })

    const image = Image.base('python:3.12')
      .pipInstallFromRequirements('requirements.txt')
      .pipInstallFromPyproject('pyproject.toml', { optionalDependencies: ['dev'] })

    expect(image.dockerfile).toContain('COPY requirements.txt /.requirements.txt')
    expect(image.dockerfile).toContain('pip install -r /.requirements.txt')
    expect(image.dockerfile).toContain('pip install numpy pytest requests')
  })

  it('validates missing or malformed pyproject files', async () => {
    const { Image } = await import('../Image')

    const fsModule = {
      existsSync: jest.fn(() => true),
      statSync: jest.fn(() => ({ isFile: () => true })),
      readFileSync: jest.fn(() => 'invalid = ['),
    }

    mockDynamicRequire.mockImplementation((moduleName: string) => {
      if (moduleName === 'fs') return fsModule
      if (moduleName === 'expand-tilde') return (value: string) => value
      if (moduleName === '@iarna/toml')
        return {
          parse: jest.fn(() => {
            throw new Error('Unexpected end of input')
          }),
        }
      return {}
    })

    expect(() => Image.base('python:3.12').pipInstallFromPyproject('pyproject.toml')).toThrow(
      'Invalid pyproject.toml file pyproject.toml: Unexpected end of input',
    )
  })

  it('validates pyproject files without dependencies metadata', async () => {
    const { Image } = await import('../Image')

    const fsModule = {
      existsSync: jest.fn(() => true),
      statSync: jest.fn(() => ({ isFile: () => true })),
      readFileSync: jest.fn(() => '[project]'),
    }

    mockDynamicRequire.mockImplementation((moduleName: string) => {
      if (moduleName === 'fs') return fsModule
      if (moduleName === 'expand-tilde') return (value: string) => value
      if (moduleName === '@iarna/toml') return { parse: jest.fn(() => ({ project: {} })) }
      return {}
    })

    expect(() => Image.base('python:3.12').pipInstallFromPyproject('pyproject.toml')).toThrow(
      'No [project.dependencies] section in pyproject.toml file.',
    )
  })

  it('validates malformed optional-dependencies in pyproject files', async () => {
    const { Image } = await import('../Image')

    const fsModule = {
      existsSync: jest.fn(() => true),
      statSync: jest.fn(() => ({ isFile: () => true })),
      readFileSync: jest.fn(() => '[project]'),
    }

    mockDynamicRequire.mockImplementation((moduleName: string) => {
      if (moduleName === 'fs') return fsModule
      if (moduleName === 'expand-tilde') return (value: string) => value
      if (moduleName === '@iarna/toml') {
        return {
          parse: jest.fn(() => ({ project: { dependencies: ['requests'], 'optional-dependencies': ['bad'] } })),
        }
      }
      return {}
    })

    expect(() =>
      Image.base('python:3.12').pipInstallFromPyproject('pyproject.toml', { optionalDependencies: ['dev'] }),
    ).toThrow('optional-dependencies must be a mapping in pyproject.toml')
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

  it('parseCopyCommand handles json array copy commands', async () => {
    const { Image } = await import('../Image')
    const imageRuntime = Image as unknown as Record<string, (...args: unknown[]) => unknown>

    const parsed = imageRuntime.parseCopyCommand('COPY ["./a.txt", "./b.txt", "/app/"]') as {
      sources: string[]
      dest: string
    }

    expect(parsed).toEqual({ sources: ['./a.txt,', './b.txt,'], dest: '/app/' })
  })

  it('extractCopySources ignores heredoc and stage copy commands', async () => {
    const { Image } = await import('../Image')
    const fastGlob = { sync: jest.fn(() => ['/repo/a.txt']) }
    mockDynamicRequire.mockImplementation((moduleName: string) => {
      if (moduleName === 'fast-glob') return fastGlob
      return {}
    })

    const imageRuntime = Image as unknown as Record<string, (...args: unknown[]) => unknown>
    const sources = imageRuntime.extractCopySources(
      ['COPY --from=builder /tmp/a.txt /app/a.txt', 'COPY <<EOF /app/a.txt', 'COPY ./b.txt /app/b.txt'].join('\n'),
      '/repo',
    ) as Array<[string, string]>

    expect(sources).toEqual([['/repo/a.txt', './b.txt']])
  })
})
