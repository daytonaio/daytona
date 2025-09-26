import { Daytona, Sandbox } from '@daytonaio/sdk'

async function interactivePtySession(sandbox: Sandbox) {
  console.log('=== First PTY Session: Interactive Command with Exit ===')

  const ptySessionId = 'interactive-pty-session'

  // Create PTY session with data handler
  const ptyHandle = await sandbox.process.createPty({
    id: ptySessionId,
    cols: 120,
    rows: 30,
    onData: (data) => {
      // Decode UTF-8 bytes to text and write directly to preserve terminal formatting
      const text = new TextDecoder().decode(data)
      process.stdout.write(text)
    },
  })

  // Send interactive command
  console.log('\nSending interactive read command...')
  await ptyHandle.sendInput('read -p "Enter your name: " name && echo "Hello, $name!"\n')

  // Wait and respond
  await new Promise((resolve) => setTimeout(resolve, 1000))
  console.log("\nResponding with 'Bob'...")
  await ptyHandle.sendInput('Bob\n')

  // Resize the PTY session
  const ptySessionInfo = await sandbox.process.resizePtySession(ptySessionId, 80, 25)
  console.log(`\nPTY session resized to ${ptySessionInfo.cols}x${ptySessionInfo.rows}`)

  // Send another command
  await new Promise((resolve) => setTimeout(resolve, 1000))
  console.log('\nSending directory listing command...')
  await ptyHandle.sendInput('ls -la\n')

  // Send exit command
  await new Promise((resolve) => setTimeout(resolve, 1000))
  console.log('\nSending exit command...')
  await ptyHandle.sendInput('exit\n')

  // Wait for PTY to exit
  const result = await ptyHandle.wait()
  console.log(`\nPTY session exited with code: ${result.exitCode}`)
  if (result.error) {
    console.log(`Error: ${result.error}`)
  }
}

async function killPtySession(sandbox: Sandbox) {
  console.log('\n=== Second PTY Session: Kill PTY Session ===')

  const ptySessionId = 'kill-pty-session'

  // Create PTY session with data handler
  const ptyHandle = await sandbox.process.createPty({
    id: ptySessionId,
    cols: 120,
    rows: 30,
    onData: (data) => {
      // Decode UTF-8 bytes to text and write directly to preserve terminal formatting
      const text = new TextDecoder().decode(data)
      process.stdout.write(text)
    },
  })

  // Send a long-running command
  console.log('\nSending long-running command (infinite loop)...')
  await ptyHandle.sendInput('while true; do echo "Running... $(date)"; sleep 1; done\n')

  // Let it run for a few seconds
  await new Promise((resolve) => setTimeout(resolve, 3000))

  // Kill the PTY session
  console.log('\nKilling PTY session...')
  await ptyHandle.kill()

  // Wait for PTY to terminate
  const result = await ptyHandle.wait()
  console.log(`\nPTY session terminated. Exit code: ${result.exitCode}`)
  if (result.error) {
    console.log(`Error: ${result.error}`)
  }
}

async function main() {
  const daytona = new Daytona()
  const sandbox = await daytona.create()

  try {
    // Interactive PTY session with exit
    await interactivePtySession(sandbox)
    // PTY session killed with .kill()
    await killPtySession(sandbox)
  } catch (error) {
    console.error('Error executing PTY commands:', error)
  } finally {
    console.log(`\nDeleting sandbox: ${sandbox.id}`)
    await daytona.delete(sandbox)
  }
}

main().catch(console.error)
