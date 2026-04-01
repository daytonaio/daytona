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
    })

    console.log(`[E2E] Sandbox ready: id=${sandbox.id}, state=${sandbox.state}`)
  })

  afterAll(async () => {
    if (!sandbox) {
      return
    }

    console.log(`[E2E] Cleaning up sandbox: ${sandbox.id}`)
    try {
      await daytona.delete(sandbox)
      console.log('[E2E] Sandbox deleted successfully')
    } catch (error) {
      console.error('[E2E] Sandbox cleanup failed:', error)
    }
  })

  describe('Sandbox Lifecycle', () => {
    test('sandbox metadata and lifecycle utility methods work', async () => {
      console.log('[E2E][Lifecycle] Verifying sandbox state and metadata...')
      expect(sandbox.state).toBe('started')
      expect(sandbox.id).toBeDefined()
      expect(sandbox.name).toBeDefined()
      expect(sandbox.organizationId).toBeDefined()

      const userHomeDir = await sandbox.getUserHomeDir()
      console.log('[E2E][Lifecycle] userHomeDir:', userHomeDir)
      expect(userHomeDir).toBeDefined()
      expect(userHomeDir).toContain('/')

      const workDir = await sandbox.getWorkDir()
      console.log('[E2E][Lifecycle] workDir:', workDir)
      expect(workDir).toBeDefined()
      expect(workDir).toContain('/')

      console.log('[E2E][Lifecycle] Setting labels...')
      const labels = await sandbox.setLabels({ test: 'e2e' })
      expect(labels.test).toBe('e2e')

      console.log('[E2E][Lifecycle] Setting auto-stop interval...')
      await sandbox.setAutostopInterval(30)
      expect(sandbox.autoStopInterval).toBe(30)

      console.log('[E2E][Lifecycle] Refreshing sandbox data...')
      await sandbox.refreshData()
      expect(sandbox.id).toBeDefined()
      expect(sandbox.state).toBe('started')

      console.log('[E2E][Lifecycle] Refreshing sandbox activity...')
      await sandbox.refreshActivity()
      expect(true).toBe(true)
    })
  })

  describe('File System Operations', () => {
    test('filesystem operations succeed end-to-end', async () => {
      console.log('[E2E][FS] Creating folder test-dir...')
      await sandbox.fs.createFolder('test-dir', '755')

      console.log('[E2E][FS] Uploading file test-dir/hello.txt...')
      await sandbox.fs.uploadFile(Buffer.from('hello'), 'test-dir/hello.txt')

      console.log('[E2E][FS] Listing files in test-dir...')
      const files = await sandbox.fs.listFiles('test-dir')
      expect(files.length).toBeGreaterThan(0)
      expect(files.some((file) => file.name === 'hello.txt')).toBe(true)

      console.log('[E2E][FS] Getting file details...')
      const fileDetails = await sandbox.fs.getFileDetails('test-dir/hello.txt')
      expect(fileDetails).toBeDefined()
      expect(fileDetails.size).toBe(5)
      expect(fileDetails.name).toBe('hello.txt')

      console.log('[E2E][FS] Downloading file and checking contents...')
      const downloaded = await sandbox.fs.downloadFile('test-dir/hello.txt')
      expect(downloaded.toString()).toBe('hello')

      console.log('[E2E][FS] Finding text in files...')
      const findMatches = await sandbox.fs.findFiles('test-dir', 'hello')
      expect(findMatches.length).toBeGreaterThan(0)

      console.log('[E2E][FS] Searching files by glob...')
      const searchResult = await sandbox.fs.searchFiles('test-dir', '*.txt')
      expect(searchResult.files).toBeDefined()
      expect(searchResult.files.length).toBeGreaterThan(0)
      expect(searchResult.files.some((path) => path.endsWith('hello.txt'))).toBe(true)

      console.log('[E2E][FS] Replacing text in file...')
      const replaceResult = await sandbox.fs.replaceInFiles(['test-dir/hello.txt'], 'hello', 'world')
      expect(replaceResult).toBeDefined()
      expect(replaceResult.length).toBeGreaterThan(0)

      console.log('[E2E][FS] Downloading file again to verify replacement...')
      const replaced = await sandbox.fs.downloadFile('test-dir/hello.txt')
      expect(replaced.toString()).toBe('world')

      console.log('[E2E][FS] Moving file to test-dir/moved.txt...')
      await sandbox.fs.moveFiles('test-dir/hello.txt', 'test-dir/moved.txt')

      console.log('[E2E][FS] Deleting moved file...')
      await sandbox.fs.deleteFile('test-dir/moved.txt')

      const filesAfterDelete = await sandbox.fs.listFiles('test-dir')
      expect(filesAfterDelete.some((file) => file.name === 'moved.txt')).toBe(false)
    })
  })

  describe('Process Execution', () => {
    test('process execution and codeRun work', async () => {
      console.log('[E2E][Process] Executing simple echo command...')
      const echoResponse = await sandbox.process.executeCommand('echo hello')
      expect(echoResponse.exitCode).toBe(0)
      expect(echoResponse.result).toContain('hello')

      console.log('[E2E][Process] Executing command with cwd...')
      const cwdResponse = await sandbox.process.executeCommand('ls /', '/tmp')
      expect(cwdResponse.exitCode).toBe(0)
      expect(cwdResponse.result).toContain('bin')

      console.log('[E2E][Process] Executing command with env var...')
      const envResponse = await sandbox.process.executeCommand('echo $MY_VAR', undefined, { MY_VAR: 'test123' })
      expect(envResponse.exitCode).toBe(0)
      expect(envResponse.result).toContain('test123')

      console.log('[E2E][Process] Running python code via codeRun...')
      const codeRunResponse = await sandbox.process.codeRun('print("hello from python")')
      expect(codeRunResponse.exitCode).toBe(0)
      expect(codeRunResponse.result).toContain('hello from python')

      console.log('[E2E][Process] Executing failing command...')
      const failResponse = await sandbox.process.executeCommand('exit 1')
      expect(failResponse.exitCode).not.toBe(0)
    })
  })

  describe('Session Management', () => {
    test('session lifecycle and session stateful env work', async () => {
      const sessionId = 'test-session'

      console.log('[E2E][Session] Creating session...')
      await sandbox.process.createSession(sessionId)

      console.log('[E2E][Session] Getting session details...')
      const session = await sandbox.process.getSession(sessionId)
      expect(session).toBeDefined()
      expect(session.sessionId).toBe(sessionId)

      console.log('[E2E][Session] Exporting FOO=bar in session...')
      const exportResponse = await sandbox.process.executeSessionCommand(sessionId, { command: 'export FOO=bar' })
      expect(exportResponse).toBeDefined()

      console.log('[E2E][Session] Reading FOO from same session...')
      const echoResponse = await sandbox.process.executeSessionCommand(sessionId, { command: 'echo $FOO' })
      expect(echoResponse.stdout ?? echoResponse.output ?? '').toContain('bar')

      console.log('[E2E][Session] Listing sessions...')
      const sessions = await sandbox.process.listSessions()
      expect(sessions.some((s) => s.sessionId === sessionId)).toBe(true)

      console.log('[E2E][Session] Deleting session...')
      await sandbox.process.deleteSession(sessionId)
    })
  })

  describe('Git Operations', () => {
    test('clone/status/branches work in sandbox', async () => {
      console.log('[E2E][Git] Cloning test repository...')
      await sandbox.git.clone('https://github.com/octocat/Hello-World.git', 'hello-world')

      console.log('[E2E][Git] Getting git status...')
      const status = await sandbox.git.status('hello-world')
      expect(status).toBeDefined()
      expect(status.currentBranch).toBeDefined()

      console.log('[E2E][Git] Listing branches...')
      const branches = await sandbox.git.branches('hello-world')
      expect(branches).toBeDefined()
      expect(branches.branches).toBeDefined()
      expect(branches.branches.length).toBeGreaterThan(0)
    })
  })

  describe('Preview Link', () => {
    test('getPreviewLink returns URL and token data', async () => {
      console.log('[E2E][Preview] Getting preview link for port 8080...')
      const preview = await sandbox.getPreviewLink(8080)

      expect(preview.url).toBeDefined()
      expect(preview.url).toContain('http')
      expect(preview.token).toBeDefined()
      expect(preview.url.includes(preview.token!) || preview.token!.length > 0).toBe(true)
    })
  })

  describe('Sandbox List/Get via Daytona client', () => {
    test('daytona.list and daytona.get work', async () => {
      console.log('[E2E][Client] Listing sandboxes...')
      const listResult = await daytona.list()
      expect(listResult.items).toBeDefined()
      expect(Array.isArray(listResult.items)).toBe(true)
      expect(listResult.total).toBeGreaterThan(0)

      console.log('[E2E][Client] Getting sandbox by id...')
      const fetchedSandbox = await daytona.get(sandbox.id)
      expect(fetchedSandbox).toBeDefined()
      expect(fetchedSandbox.id).toBe(sandbox.id)
    })
  })
})
