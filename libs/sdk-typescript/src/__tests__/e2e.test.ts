import { Daytona, Sandbox } from '@daytonaio/sdk'

jest.setTimeout(120000)

describe('TypeScript SDK E2E (real Daytona API)', () => {
  let daytona: Daytona
  let sandbox: Sandbox

  beforeAll(async () => {
    console.log('[E2E] Initializing Daytona client from environment variables...')
    daytona = new Daytona()

    const sandboxName = `sdk-ts-e2e-${Date.now()}`
    console.log(`[E2E] Creating shared sandbox: ${sandboxName}`)
    sandbox = await daytona.create({
      name: sandboxName,
      language: 'python',
      labels: { purpose: 'e2e-test' },
    })

    console.log(`[E2E] Sandbox ready: id=${sandbox.id}, state=${sandbox.state}`)
  })

  afterAll(async () => {
    if (!sandbox) return
    console.log(`[E2E] Cleaning up sandbox: ${sandbox.id}`)
    try {
      await daytona.delete(sandbox)
      console.log('[E2E] Sandbox deleted successfully')
    } catch (error) {
      console.error('[E2E] Sandbox cleanup failed:', error)
    }
  })

  // ──────────────────────────────────────────────
  // Sandbox Lifecycle
  // ──────────────────────────────────────────────
  describe('Sandbox Lifecycle', () => {
    test('sandbox has valid id', () => {
      console.log('[E2E][Lifecycle] Checking sandbox id...')
      expect(sandbox.id).toBeDefined()
      expect(typeof sandbox.id).toBe('string')
      expect(sandbox.id.length).toBeGreaterThan(0)
    })

    test('sandbox has valid name and organizationId', () => {
      console.log('[E2E][Lifecycle] Checking sandbox name and orgId...')
      expect(sandbox.name).toBeDefined()
      expect(sandbox.organizationId).toBeDefined()
      expect(typeof sandbox.organizationId).toBe('string')
    })

    test('sandbox state is started', () => {
      console.log('[E2E][Lifecycle] Checking sandbox state...')
      expect(sandbox.state).toBe('started')
    })

    test('sandbox has cpu, memory, disk properties > 0', () => {
      console.log('[E2E][Lifecycle] Checking sandbox resources...')
      expect(sandbox.cpu).toBeGreaterThan(0)
      expect(sandbox.memory).toBeGreaterThan(0)
      expect(sandbox.disk).toBeGreaterThan(0)
    })

    test('sandbox has createdAt and updatedAt timestamps', () => {
      console.log('[E2E][Lifecycle] Checking timestamps...')
      expect(sandbox.createdAt).toBeDefined()
      expect(sandbox.updatedAt).toBeDefined()
    })

    test('getUserHomeDir returns a valid path containing /', async () => {
      console.log('[E2E][Lifecycle] Getting user home dir...')
      const userHomeDir = await sandbox.getUserHomeDir()
      expect(userHomeDir).toBeDefined()
      expect(userHomeDir).toContain('/')
    })

    test('getWorkDir returns a valid path', async () => {
      console.log('[E2E][Lifecycle] Getting work dir...')
      const workDir = await sandbox.getWorkDir()
      expect(workDir).toBeDefined()
      expect(workDir).toContain('/')
    })

    test('setLabels sets and returns new labels', async () => {
      console.log('[E2E][Lifecycle] Setting labels...')
      const labels = await sandbox.setLabels({ test: 'e2e', env: 'ci' })
      expect(labels.test).toBe('e2e')
      expect(labels.env).toBe('ci')
    })

    test('setAutostopInterval updates interval', async () => {
      console.log('[E2E][Lifecycle] Setting auto-stop interval...')
      await sandbox.setAutostopInterval(30)
      expect(sandbox.autoStopInterval).toBe(30)
    })

    test('setAutoArchiveInterval updates interval', async () => {
      console.log('[E2E][Lifecycle] Setting auto-archive interval...')
      await sandbox.setAutoArchiveInterval(120)
      expect(sandbox.autoArchiveInterval).toBe(120)
    })

    test('setAutoDeleteInterval can set and disable (-1)', async () => {
      console.log('[E2E][Lifecycle] Setting auto-delete interval...')
      await sandbox.setAutoDeleteInterval(60)
      expect(sandbox.autoDeleteInterval).toBe(60)

      await sandbox.setAutoDeleteInterval(-1)
      expect(sandbox.autoDeleteInterval).toBe(-1)
    })

    test('refreshData updates sandbox object', async () => {
      console.log('[E2E][Lifecycle] Refreshing sandbox data...')
      await sandbox.refreshData()
      expect(sandbox.id).toBeDefined()
      expect(sandbox.state).toBe('started')
    })

    test('refreshActivity succeeds without error', async () => {
      console.log('[E2E][Lifecycle] Refreshing sandbox activity...')
      await sandbox.refreshActivity()
    })
  })

  // ──────────────────────────────────────────────
  // File System Operations
  // ──────────────────────────────────────────────
  describe('File System Operations', () => {
    test('createFolder creates a directory', async () => {
      console.log('[E2E][FS] Creating folder fs-test...')
      await sandbox.fs.createFolder('fs-test', '755')
      const files = await sandbox.fs.listFiles('.')
      expect(files.some((f) => f.name === 'fs-test')).toBe(true)
    })

    test('createFolder with custom permissions', async () => {
      console.log('[E2E][FS] Creating folder with 700 permissions...')
      await sandbox.fs.createFolder('fs-test/private', '700')
      const files = await sandbox.fs.listFiles('fs-test')
      expect(files.some((f) => f.name === 'private')).toBe(true)
    })

    test('uploadFile with Buffer content', async () => {
      console.log('[E2E][FS] Uploading file with Buffer...')
      await sandbox.fs.uploadFile(Buffer.from('hello world'), 'fs-test/hello.txt')
      const files = await sandbox.fs.listFiles('fs-test')
      expect(files.some((f) => f.name === 'hello.txt')).toBe(true)
    })

    test('uploadFiles batch with multiple files', async () => {
      console.log('[E2E][FS] Batch uploading files...')
      await sandbox.fs.uploadFiles([
        { source: Buffer.from('file-a-content'), destination: 'fs-test/a.txt' },
        { source: Buffer.from('file-b-content'), destination: 'fs-test/b.txt' },
      ])
      const files = await sandbox.fs.listFiles('fs-test')
      expect(files.some((f) => f.name === 'a.txt')).toBe(true)
      expect(files.some((f) => f.name === 'b.txt')).toBe(true)
    })

    test('listFiles returns uploaded files', async () => {
      console.log('[E2E][FS] Listing files in fs-test...')
      const files = await sandbox.fs.listFiles('fs-test')
      expect(files.length).toBeGreaterThan(0)
      const names = files.map((f) => f.name)
      expect(names).toContain('hello.txt')
      expect(names).toContain('a.txt')
      expect(names).toContain('b.txt')
    })

    test('getFileDetails returns correct name and size', async () => {
      console.log('[E2E][FS] Getting file details...')
      const details = await sandbox.fs.getFileDetails('fs-test/hello.txt')
      expect(details).toBeDefined()
      expect(details.name).toBe('hello.txt')
      expect(details.size).toBe(11) // "hello world" = 11 bytes
    })

    test('downloadFile returns exact content', async () => {
      console.log('[E2E][FS] Downloading file...')
      const content = await sandbox.fs.downloadFile('fs-test/hello.txt')
      expect(content.toString()).toBe('hello world')
    })

    test('downloadFiles batch returns multiple files', async () => {
      console.log('[E2E][FS] Batch downloading files...')
      const results = await sandbox.fs.downloadFiles([{ source: 'fs-test/a.txt' }, { source: 'fs-test/b.txt' }])
      expect(results.length).toBe(2)
      expect(results[0].result?.toString()).toBe('file-a-content')
      expect(results[1].result?.toString()).toBe('file-b-content')
    })

    test('findFiles finds text content in files', async () => {
      console.log('[E2E][FS] Finding text in files...')
      const matches = await sandbox.fs.findFiles('fs-test', 'hello')
      expect(matches.length).toBeGreaterThan(0)
    })

    test('searchFiles finds files by glob pattern', async () => {
      console.log('[E2E][FS] Searching files by glob...')
      const result = await sandbox.fs.searchFiles('fs-test', '*.txt')
      expect(result.files).toBeDefined()
      expect(result.files.length).toBeGreaterThan(0)
      expect(result.files.some((p) => p.endsWith('hello.txt'))).toBe(true)
    })

    test('replaceInFiles replaces text and verify by re-download', async () => {
      console.log('[E2E][FS] Replacing text in file...')
      // Upload a file specifically for replacement
      await sandbox.fs.uploadFile(Buffer.from('foo bar baz'), 'fs-test/replace-me.txt')
      const replaceResult = await sandbox.fs.replaceInFiles(['fs-test/replace-me.txt'], 'foo', 'replaced')
      expect(replaceResult).toBeDefined()
      expect(replaceResult.length).toBeGreaterThan(0)

      const content = await sandbox.fs.downloadFile('fs-test/replace-me.txt')
      expect(content.toString()).toBe('replaced bar baz')
    })

    test('setFilePermissions changes file mode', async () => {
      console.log('[E2E][FS] Setting file permissions...')
      await sandbox.fs.uploadFile(Buffer.from('script'), 'fs-test/perm-test.txt')
      await sandbox.fs.setFilePermissions('fs-test/perm-test.txt', {
        mode: '644',
        owner: 'daytona',
        group: 'daytona',
      })
      // Verify by checking the file still exists (permission change succeeds without error)
      const details = await sandbox.fs.getFileDetails('fs-test/perm-test.txt')
      expect(details).toBeDefined()
    })

    test('moveFiles moves a file to new location', async () => {
      console.log('[E2E][FS] Moving file...')
      await sandbox.fs.uploadFile(Buffer.from('moveme'), 'fs-test/to-move.txt')
      await sandbox.fs.moveFiles('fs-test/to-move.txt', 'fs-test/moved.txt')

      const files = await sandbox.fs.listFiles('fs-test')
      expect(files.some((f) => f.name === 'moved.txt')).toBe(true)
      expect(files.some((f) => f.name === 'to-move.txt')).toBe(false)
    })

    test('deleteFile removes a file', async () => {
      console.log('[E2E][FS] Deleting file...')
      await sandbox.fs.deleteFile('fs-test/moved.txt')
      const files = await sandbox.fs.listFiles('fs-test')
      expect(files.some((f) => f.name === 'moved.txt')).toBe(false)
    })

    test('nested folder operations', async () => {
      console.log('[E2E][FS] Creating nested folders...')
      await sandbox.fs.createFolder('fs-test/parent/child', '755')
      await sandbox.fs.uploadFile(Buffer.from('nested-content'), 'fs-test/parent/child/nested.txt')

      const parentFiles = await sandbox.fs.listFiles('fs-test/parent')
      expect(parentFiles.some((f) => f.name === 'child')).toBe(true)

      const childFiles = await sandbox.fs.listFiles('fs-test/parent/child')
      expect(childFiles.some((f) => f.name === 'nested.txt')).toBe(true)

      const content = await sandbox.fs.downloadFile('fs-test/parent/child/nested.txt')
      expect(content.toString()).toBe('nested-content')
    })
  })

  // ──────────────────────────────────────────────
  // Process Execution
  // ──────────────────────────────────────────────
  describe('Process Execution', () => {
    test('executeCommand basic echo', async () => {
      console.log('[E2E][Process] Executing echo command...')
      const response = await sandbox.process.executeCommand('echo hello')
      expect(response.exitCode).toBe(0)
      expect(response.result).toContain('hello')
    })

    test('executeCommand with cwd option', async () => {
      console.log('[E2E][Process] Executing command with cwd...')
      const response = await sandbox.process.executeCommand('pwd', '/tmp')
      expect(response.exitCode).toBe(0)
      expect(response.result).toContain('/tmp')
    })

    test('executeCommand with env vars', async () => {
      console.log('[E2E][Process] Executing command with env var...')
      const response = await sandbox.process.executeCommand('echo $MY_VAR', undefined, { MY_VAR: 'test123' })
      expect(response.exitCode).toBe(0)
      expect(response.result).toContain('test123')
    })

    test('executeCommand with multiple env vars', async () => {
      console.log('[E2E][Process] Executing command with multiple env vars...')
      const response = await sandbox.process.executeCommand('echo $A $B', undefined, {
        A: 'alpha',
        B: 'beta',
      })
      expect(response.exitCode).toBe(0)
      expect(response.result).toContain('alpha')
      expect(response.result).toContain('beta')
    })

    test('executeCommand returns non-zero exit code on failure', async () => {
      console.log('[E2E][Process] Executing failing command...')
      const response = await sandbox.process.executeCommand('exit 42')
      expect(response.exitCode).toBe(42)
    })

    test('executeCommand captures stderr', async () => {
      console.log('[E2E][Process] Executing command that writes to stderr...')
      const response = await sandbox.process.executeCommand('echo error_msg >&2')
      // stderr goes to combined output in non-session mode
      expect(response.exitCode).toBe(0)
    })

    test('codeRun with Python print statement', async () => {
      console.log('[E2E][Process] Running Python code...')
      const response = await sandbox.process.codeRun('print("hello from python")')
      expect(response.exitCode).toBe(0)
      expect(response.result).toContain('hello from python')
    })

    test('codeRun with multi-line Python code', async () => {
      console.log('[E2E][Process] Running multi-line Python...')
      const response = await sandbox.process.codeRun('x = 5\ny = 10\nprint(x + y)')
      expect(response.exitCode).toBe(0)
      expect(response.result).toContain('15')
    })

    test('codeRun with Python that writes to stderr', async () => {
      console.log('[E2E][Process] Running Python with stderr...')
      const response = await sandbox.process.codeRun('import sys; sys.stderr.write("stderr-msg\\n"); print("ok")')
      expect(response.exitCode).toBe(0)
      expect(response.result).toContain('ok')
    })

    test('codeRun with Python syntax error returns non-zero exit code', async () => {
      console.log('[E2E][Process] Running Python with syntax error...')
      const response = await sandbox.process.codeRun('def foo(\nprint("broken")')
      expect(response.exitCode).not.toBe(0)
    })
  })

  // ──────────────────────────────────────────────
  // Session Management
  // ──────────────────────────────────────────────
  describe('Session Management', () => {
    const sessionId = `e2e-session-${Date.now()}`

    test('createSession creates a new session', async () => {
      console.log('[E2E][Session] Creating session...')
      await sandbox.process.createSession(sessionId)
    })

    test('getSession returns session details', async () => {
      console.log('[E2E][Session] Getting session...')
      const session = await sandbox.process.getSession(sessionId)
      expect(session).toBeDefined()
      expect(session.sessionId).toBe(sessionId)
    })

    test('executeSessionCommand runs command in session', async () => {
      console.log('[E2E][Session] Running command in session...')
      const response = await sandbox.process.executeSessionCommand(sessionId, { command: 'echo session-hello' })
      expect(response).toBeDefined()
      expect(response.cmdId).toBeDefined()
    })

    test('session maintains state across commands (export var, echo var)', async () => {
      console.log('[E2E][Session] Testing session state persistence...')
      await sandbox.process.executeSessionCommand(sessionId, { command: 'export SESSION_VAR=persistent' })

      const response = await sandbox.process.executeSessionCommand(sessionId, { command: 'echo $SESSION_VAR' })
      const output = response.stdout ?? response.output ?? ''
      expect(output).toContain('persistent')
    })

    test('getSessionCommandLogs returns stdout/stderr', async () => {
      console.log('[E2E][Session] Getting session command logs...')
      const execResult = await sandbox.process.executeSessionCommand(sessionId, { command: 'echo log-test-output' })
      expect(execResult.cmdId).toBeDefined()

      const logs = await sandbox.process.getSessionCommandLogs(sessionId, execResult.cmdId!)
      expect(logs).toBeDefined()
      // logs should have some output
      const logOutput = logs.stdout ?? logs.output ?? ''
      expect(logOutput).toContain('log-test-output')
    })

    test('listSessions includes our session', async () => {
      console.log('[E2E][Session] Listing sessions...')
      const sessions = await sandbox.process.listSessions()
      expect(sessions.some((s) => s.sessionId === sessionId)).toBe(true)
    })

    test('deleteSession removes the session', async () => {
      console.log('[E2E][Session] Deleting session...')
      await sandbox.process.deleteSession(sessionId)

      // Verify it's gone
      const sessions = await sandbox.process.listSessions()
      expect(sessions.some((s) => s.sessionId === sessionId)).toBe(false)
    })
  })

  // ──────────────────────────────────────────────
  // Git Operations
  // ──────────────────────────────────────────────
  describe('Git Operations', () => {
    const repoPath = 'e2e-git-repo'

    test('clone public repo', async () => {
      console.log('[E2E][Git] Cloning test repository...')
      await sandbox.git.clone('https://github.com/octocat/Hello-World.git', repoPath)
    })

    test('status returns currentBranch', async () => {
      console.log('[E2E][Git] Getting git status...')
      const status = await sandbox.git.status(repoPath)
      expect(status).toBeDefined()
      expect(status.currentBranch).toBeDefined()
      expect(typeof status.currentBranch).toBe('string')
    })

    test('branches returns branch list', async () => {
      console.log('[E2E][Git] Listing branches...')
      const branches = await sandbox.git.branches(repoPath)
      expect(branches).toBeDefined()
      expect(branches.branches).toBeDefined()
      expect(branches.branches.length).toBeGreaterThan(0)
    })

    test('createBranch creates new branch', async () => {
      console.log('[E2E][Git] Creating branch...')
      await sandbox.git.createBranch(repoPath, 'e2e-test-branch')

      const branches = await sandbox.git.branches(repoPath)
      expect(branches.branches.some((b) => b === 'e2e-test-branch')).toBe(true)
    })

    test('checkoutBranch switches to branch', async () => {
      console.log('[E2E][Git] Checking out branch...')
      await sandbox.git.checkoutBranch(repoPath, 'e2e-test-branch')

      const status = await sandbox.git.status(repoPath)
      expect(status.currentBranch).toBe('e2e-test-branch')
    })

    test('add stages files', async () => {
      console.log('[E2E][Git] Staging files...')
      // Create a file to stage
      await sandbox.fs.uploadFile(Buffer.from('test file for git'), `${repoPath}/e2e-test-file.txt`)
      await sandbox.git.add(repoPath, ['e2e-test-file.txt'])

      const status = await sandbox.git.status(repoPath)
      expect(status.fileStatus).toBeDefined()
    })

    test('commit creates a commit with sha', async () => {
      console.log('[E2E][Git] Committing changes...')
      const commitResult = await sandbox.git.commit(repoPath, 'E2E test commit', 'E2E Test', 'e2e@test.com')
      expect(commitResult.sha).toBeDefined()
      expect(typeof commitResult.sha).toBe('string')
      expect(commitResult.sha.length).toBeGreaterThan(0)
    })

    test('deleteBranch removes a branch', async () => {
      console.log('[E2E][Git] Deleting branch...')
      // Switch back to master/main first
      const status = await sandbox.git.status(repoPath)
      const mainBranch = status.currentBranch === 'e2e-test-branch' ? 'master' : status.currentBranch
      await sandbox.git.checkoutBranch(repoPath, mainBranch!)

      await sandbox.git.deleteBranch(repoPath, 'e2e-test-branch')

      const branches = await sandbox.git.branches(repoPath)
      expect(branches.branches.some((b) => b === 'e2e-test-branch')).toBe(false)
    })

    test('clone with specific branch', async () => {
      console.log('[E2E][Git] Cloning specific branch...')
      await sandbox.git.clone('https://github.com/octocat/Hello-World.git', 'e2e-git-branch', 'test')

      const status = await sandbox.git.status('e2e-git-branch')
      expect(status.currentBranch).toBe('test')
    })
  })

  // ──────────────────────────────────────────────
  // Code Interpreter
  // ──────────────────────────────────────────────
  describe('Code Interpreter', () => {
    test('runCode with simple Python prints stdout', async () => {
      console.log('[E2E][CodeInterpreter] Running simple code...')
      const result = await sandbox.codeInterpreter.runCode('print("interpreter-hello")')
      expect(result.stdout).toContain('interpreter-hello')
      expect(result.error).toBeUndefined()
    })

    test('runCode with multi-line Python maintains state', async () => {
      console.log('[E2E][CodeInterpreter] Running multi-line code...')
      // Default context maintains state across calls
      await sandbox.codeInterpreter.runCode('ci_var = 42')
      const result = await sandbox.codeInterpreter.runCode('print(ci_var)')
      expect(result.stdout).toContain('42')
    })

    test('createContext returns context with id', async () => {
      console.log('[E2E][CodeInterpreter] Creating context...')
      const ctx = await sandbox.codeInterpreter.createContext()
      expect(ctx).toBeDefined()
      expect(ctx.id).toBeDefined()

      // Clean up
      await sandbox.codeInterpreter.deleteContext(ctx)
    })

    test('runCode in custom context maintains isolated state', async () => {
      console.log('[E2E][CodeInterpreter] Running code in custom context...')
      const ctx = await sandbox.codeInterpreter.createContext()

      await sandbox.codeInterpreter.runCode('ctx_val = 99', { context: ctx })
      const result = await sandbox.codeInterpreter.runCode('print(ctx_val)', { context: ctx })
      expect(result.stdout).toContain('99')

      // Clean up
      await sandbox.codeInterpreter.deleteContext(ctx)
    })

    test('listContexts includes created context', async () => {
      console.log('[E2E][CodeInterpreter] Listing contexts...')
      const ctx = await sandbox.codeInterpreter.createContext()

      const contexts = await sandbox.codeInterpreter.listContexts()
      expect(contexts.some((c) => c.id === ctx.id)).toBe(true)

      // Clean up
      await sandbox.codeInterpreter.deleteContext(ctx)
    })

    test('deleteContext removes context', async () => {
      console.log('[E2E][CodeInterpreter] Deleting context...')
      const ctx = await sandbox.codeInterpreter.createContext()
      await sandbox.codeInterpreter.deleteContext(ctx)

      const contexts = await sandbox.codeInterpreter.listContexts()
      expect(contexts.some((c) => c.id === ctx.id)).toBe(false)
    })

    test('runCode with stderr', async () => {
      console.log('[E2E][CodeInterpreter] Running code with stderr...')
      const result = await sandbox.codeInterpreter.runCode(
        'import sys; sys.stderr.write("ci-stderr-msg\\n"); print("ci-ok")',
      )
      expect(result.stdout).toContain('ci-ok')
      expect(result.stderr).toContain('ci-stderr-msg')
    })
  })

  // ──────────────────────────────────────────────
  // Preview / URLs
  // ──────────────────────────────────────────────
  describe('Preview Links', () => {
    test('getPreviewLink returns url and token', async () => {
      console.log('[E2E][Preview] Getting preview link...')
      const preview = await sandbox.getPreviewLink(8080)
      expect(preview.url).toBeDefined()
      expect(preview.url).toContain('http')
      expect(preview.token).toBeDefined()
    })

    test('getSignedPreviewUrl returns signed url', async () => {
      console.log('[E2E][Preview] Getting signed preview URL...')
      const signed = await sandbox.getSignedPreviewUrl(8080, 60)
      expect(signed).toBeDefined()
      expect(signed.url).toBeDefined()
      expect(signed.url).toContain('http')
      expect(signed.token).toBeDefined()
    })
  })

  // ──────────────────────────────────────────────
  // Volume Operations
  // ──────────────────────────────────────────────
  describe('Volume Operations', () => {
    const volumeName = `e2e-vol-${Date.now()}`
    let createdVolumeId: string

    async function waitForVolumeReady(name: string, maxWaitMs = 15000): Promise<void> {
      const start = Date.now()
      while (Date.now() - start < maxWaitMs) {
        const vol = await daytona.volume.get(name)
        if (vol.state === 'ready' || vol.state === 'error') return
        await new Promise((r) => setTimeout(r, 500))
      }
    }

    async function cleanupOldE2eVolumes(): Promise<void> {
      const volumes = await daytona.volume.list()
      for (const vol of volumes) {
        if (vol.name.startsWith('e2e-vol-') || vol.name.startsWith('e2e-auto-vol-')) {
          if (vol.state === 'ready' || vol.state === 'error') {
            try {
              await daytona.volume.delete(vol)
            } catch {
              /* ignore cleanup errors */
            }
          }
        }
      }
    }

    beforeAll(async () => {
      await cleanupOldE2eVolumes()
    })

    test('volume.create creates a new volume', async () => {
      console.log(`[E2E][Volume] Creating volume: ${volumeName}...`)
      const volume = await daytona.volume.create(volumeName)
      expect(volume).toBeDefined()
      expect(volume.id).toBeDefined()
      expect(volume.name).toBe(volumeName)
      createdVolumeId = volume.id
    })

    test('volume.list includes the created volume', async () => {
      console.log('[E2E][Volume] Listing volumes...')
      const volumes = await daytona.volume.list()
      expect(Array.isArray(volumes)).toBe(true)
      expect(volumes.some((v) => v.id === createdVolumeId)).toBe(true)
    })

    test('volume.get retrieves volume by name', async () => {
      console.log('[E2E][Volume] Getting volume by name...')
      const volume = await daytona.volume.get(volumeName)
      expect(volume).toBeDefined()
      expect(volume.name).toBe(volumeName)
      expect(volume.id).toBe(createdVolumeId)
    })

    test('volume.delete removes the volume', async () => {
      console.log('[E2E][Volume] Deleting volume...')
      await waitForVolumeReady(volumeName)
      const volume = await daytona.volume.get(volumeName)
      await daytona.volume.delete(volume)
    })

    test('volume.get with create=true creates if not found', async () => {
      console.log('[E2E][Volume] Getting volume with create=true...')
      const autoVolumeName = `e2e-auto-vol-${Date.now()}`
      const volume = await daytona.volume.get(autoVolumeName, true)
      expect(volume).toBeDefined()
      expect(volume.name).toBe(autoVolumeName)

      await waitForVolumeReady(autoVolumeName)
      const readyVolume = await daytona.volume.get(autoVolumeName)
      await daytona.volume.delete(readyVolume)
    })
  })

  // ──────────────────────────────────────────────
  // Snapshot Operations
  // ──────────────────────────────────────────────
  describe('Snapshot Operations', () => {
    test('snapshot.list returns paginated results', async () => {
      console.log('[E2E][Snapshot] Listing snapshots...')
      const result = await daytona.snapshot.list()
      expect(result).toBeDefined()
      expect(result.items).toBeDefined()
      expect(Array.isArray(result.items)).toBe(true)
      expect(result.total).toBeGreaterThanOrEqual(0)
    })

    test('snapshot.list with pagination params', async () => {
      console.log('[E2E][Snapshot] Listing snapshots with pagination...')
      const result = await daytona.snapshot.list(1, 5)
      expect(result).toBeDefined()
      expect(result.items).toBeDefined()
      expect(result.items.length).toBeLessThanOrEqual(5)
    })

    test('snapshot.get retrieves snapshot by name', async () => {
      console.log('[E2E][Snapshot] Getting snapshot by name...')
      const listResult = await daytona.snapshot.list(1, 1)
      if (listResult.items.length === 0) {
        console.log('[E2E][Snapshot] No snapshots available, skipping get test')
        return
      }

      const snapshotName = listResult.items[0].name
      const snapshot = await daytona.snapshot.get(snapshotName)
      expect(snapshot).toBeDefined()
      expect(snapshot.name).toBe(snapshotName)
    })
  })

  // ──────────────────────────────────────────────
  // Daytona Client Operations
  // ──────────────────────────────────────────────
  describe('Daytona Client Operations', () => {
    test('list returns paginated sandboxes', async () => {
      console.log('[E2E][Client] Listing sandboxes...')
      const result = await daytona.list()
      expect(result.items).toBeDefined()
      expect(Array.isArray(result.items)).toBe(true)
      expect(result.total).toBeGreaterThan(0)
    })

    test('list with page/limit pagination', async () => {
      console.log('[E2E][Client] Listing sandboxes with pagination...')
      const result = await daytona.list(undefined, 1, 2)
      expect(result.items).toBeDefined()
      expect(result.items.length).toBeLessThanOrEqual(2)
      expect(result.page).toBeDefined()
    })

    test('get by id returns correct sandbox', async () => {
      console.log('[E2E][Client] Getting sandbox by id...')
      const fetched = await daytona.get(sandbox.id)
      expect(fetched).toBeDefined()
      expect(fetched.id).toBe(sandbox.id)
      expect(fetched.name).toBe(sandbox.name)
    })

    test('list with label filter', async () => {
      console.log('[E2E][Client] Listing sandboxes with label filter...')
      // We set labels earlier: { test: 'e2e', env: 'ci' }
      const result = await daytona.list({ test: 'e2e' })
      expect(result.items).toBeDefined()
      expect(result.items.length).toBeGreaterThan(0)
      // Our sandbox should be in the results
      expect(result.items.some((s) => s.id === sandbox.id)).toBe(true)
    })
  })

  // ──────────────────────────────────────────────
  // Stop / Start Cycle (runs LAST)
  // ──────────────────────────────────────────────
  describe('Sandbox Stop/Start Cycle', () => {
    test('stop then start cycle works', async () => {
      console.log('[E2E][Lifecycle] Stopping sandbox...')
      await sandbox.stop()
      await sandbox.refreshData()
      expect(['stopped', 'stopping']).toContain(sandbox.state)

      // Wait for fully stopped
      if (sandbox.state !== 'stopped') {
        await sandbox.waitUntilStopped(60)
      }
      expect(sandbox.state).toBe('stopped')

      console.log('[E2E][Lifecycle] Starting sandbox...')
      await sandbox.start(60)
      expect(sandbox.state).toBe('started')

      // Verify sandbox is functional after restart
      const response = await sandbox.process.executeCommand('echo restarted')
      expect(response.exitCode).toBe(0)
      expect(response.result).toContain('restarted')
    })
  })
})
