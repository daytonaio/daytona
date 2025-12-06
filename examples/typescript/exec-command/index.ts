import { Daytona, Sandbox, Image, DaytonaTimeoutError, ExecutionError, OutputMessage } from '@daytonaio/sdk'

async function basicExec(sandbox: Sandbox) {
  //  run some typescript code directly
  const codeResult = await sandbox.process.codeRun('console.log("Hello World from code!")')
  if (codeResult.exitCode !== 0) {
    console.error('Error running code:', codeResult.exitCode)
  } else {
    console.log(codeResult.result)
  }

  //  run os command
  const cmdResult = await sandbox.process.executeCommand('echo "Hello World from CMD!"')
  if (cmdResult.exitCode !== 0) {
    console.error('Error running code:', cmdResult.exitCode)
  } else {
    console.log(cmdResult.result)
  }
}

async function sessionExec(sandbox: Sandbox) {
  //  exec session
  //  session allows for multiple commands to be executed in the same context
  await sandbox.process.createSession('exec-session-1')

  //  get the session details any time
  const session = await sandbox.process.getSession('exec-session-1')
  console.log('session: ', session)

  //  execute a first command in the session
  const command = await sandbox.process.executeSessionCommand('exec-session-1', {
    command: 'export FOO=BAR',
  })

  //  get the session details again to see the command has been executed
  const sessionUpdated = await sandbox.process.getSession('exec-session-1')
  console.log('sessionUpdated: ', sessionUpdated)

  //  get the command details
  const sessionCommand = await sandbox.process.getSessionCommand('exec-session-1', command.cmdId)
  console.log('sessionCommand: ', sessionCommand)

  //  execute a second command in the session and see that the environment variable is set
  const response = await sandbox.process.executeSessionCommand('exec-session-1', {
    command: 'echo $FOO',
  })
  console.log(`FOO=${response.stdout}`)

  //  we can also get the logs for the command any time after it is executed
  const logs = await sandbox.process.getSessionCommandLogs('exec-session-1', response.cmdId)
  console.log('[STDOUT]:', logs.stdout)
  console.log('[STDERR]:', logs.stderr)

  //  we can also delete the session
  await sandbox.process.deleteSession('exec-session-1')
}

async function sessionExecLogsAsync(sandbox: Sandbox) {
  console.log('Executing long running command in a session and streaming logs asynchronously...')

  const sessionId = 'exec-session-async-logs'
  await sandbox.process.createSession(sessionId)

  const command = await sandbox.process.executeSessionCommand(sessionId, {
    command:
      'counter=1; while (( counter <= 3 )); do echo "Count: $counter"; ((counter++)); sleep 2; done; non-existent-command',
    runAsync: true,
  })

  await sandbox.process.getSessionCommandLogs(
    sessionId,
    command.cmdId,
    (stdout) => console.log('[STDOUT]:', stdout),
    (stderr) => console.log('[STDERR]:', stderr),
  )
}

async function statefulCodeInterpreter(sandbox: Sandbox) {
  const logStdout = (msg: OutputMessage) => process.stdout.write(`[STDOUT] ${msg.output}`)
  const logStderr = (msg: OutputMessage) => process.stdout.write(`[STDERR] ${msg.output}`)
  const logError = (err: ExecutionError) => {
    process.stdout.write(`[ERROR] ${err.name}: ${err.value}\n`)
    if (err.traceback) {
      process.stdout.write(`${err.traceback}\n`)
    }
  }

  console.log('\n' + '='.repeat(60))
  console.log('Stateful Code Interpreter')
  console.log('='.repeat(60))
  const baseline = await sandbox.codeInterpreter.runCode(`counter = 1
print(f'Initialized counter = {counter}')`)
  process.stdout.write(`[STDOUT] ${baseline.stdout}`)

  await sandbox.codeInterpreter.runCode(
    `counter += 1
print(f'Counter after second call = {counter}')`,
    {
      onStdout: logStdout,
      onStderr: logStderr,
      onError: logError,
    },
  )

  console.log('\n' + '='.repeat(60))
  console.log('Context isolation')
  console.log('='.repeat(60))
  const ctx = await sandbox.codeInterpreter.createContext()
  try {
    await sandbox.codeInterpreter.runCode(
      `value = 'stored in isolated context'
print(f'Isolated context value: {value}')`,
      {
        context: ctx,
        onStdout: logStdout,
        onStderr: logStderr,
        onError: logError,
      },
    )

    console.log('--- Print value from same context ---')
    const ctxResult = await sandbox.codeInterpreter.runCode("print(f'Value still available: {value}')", {
      context: ctx,
    })
    process.stdout.write(`[STDOUT] ${ctxResult.stdout}`)

    console.log('--- Print value from different context ---')
    await sandbox.codeInterpreter.runCode('print(value)', {
      onStdout: logStdout,
      onStderr: logStderr,
      onError: logError,
    })
  } finally {
    await sandbox.codeInterpreter.deleteContext(ctx)
  }

  console.log('\n' + '='.repeat(60))
  console.log('Timeout handling')
  console.log('='.repeat(60))
  try {
    await sandbox.codeInterpreter.runCode(
      `import time
print('Starting long running task...')
time.sleep(5)
print('Finished!')`,
      {
        timeout: 1,
        onStdout: logStdout,
        onStderr: logStderr,
        onError: logError,
      },
    )
  } catch (error) {
    if (error instanceof DaytonaTimeoutError) {
      console.log(`Timed out as expected: ${error.message}`)
    } else {
      throw error
    }
  }
}

async function main() {
  const daytona = new Daytona()

  //  first, create a sandbox
  const sandbox = await daytona.create(
    {
      image: Image.base('ubuntu:22.04').runCommands(
        'apt-get update',
        'apt-get install -y --no-install-recommends python3 python3-pip python3-venv',
        'apt-get install -y --no-install-recommends nodejs npm coreutils',
        'curl -fsSL https://deb.nodesource.com/setup_20.x | bash -',
        'apt-get install -y nodejs',
        'npm install -g ts-node typescript',
      ),
      language: 'typescript',
      autoStopInterval: 60,
      autoArchiveInterval: 60,
      autoDeleteInterval: 120,
    },
    {
      timeout: 200,
      onSnapshotCreateLogs: console.log,
    },
  )

  try {
    await basicExec(sandbox)
    await sessionExec(sandbox)
    await sessionExecLogsAsync(sandbox)
    await statefulCodeInterpreter(sandbox)
  } catch (error) {
    console.error('Error executing commands:', error)
  } finally {
    //  cleanup
    await daytona.delete(sandbox)
  }
}

main()
