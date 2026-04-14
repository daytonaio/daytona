import { Daytona } from '@daytona/sdk'

async function main() {
  const daytona = new Daytona()

  console.log('Creating sandbox')
  const sandbox = await daytona.create()
  console.log(`Sandbox created: ${sandbox.id}`)

  // Write a file so we can verify it persists across fork and snapshot
  await sandbox.process.executeCommand('echo "Hello from the original sandbox" > /home/daytona/test.txt')

  // Fork the sandbox — creates a copy-on-write clone with identical filesystem
  console.log('Forking sandbox')
  const forked = await daytona.fork(sandbox)
  console.log(`Forked sandbox: ${forked.id}`)

  // Verify the forked sandbox has the same file
  const result = await forked.process.executeCommand('cat /home/daytona/test.txt')
  console.log(`File content in fork: ${result.result.trim()}`)

  // Write something new in the fork — does not affect the original
  await forked.process.executeCommand('echo "Modified in fork" >> /home/daytona/test.txt')

  // Create a snapshot from the forked sandbox
  console.log('Creating snapshot from forked sandbox')
  await forked.createSnapshot('my-fork-snapshot')
  console.log('Snapshot created')

  // Create a new sandbox from the snapshot
  console.log('Creating sandbox from snapshot')
  const fromSnapshot = await daytona.create({ snapshot: 'my-fork-snapshot' })
  console.log(`Sandbox from snapshot: ${fromSnapshot.id}`)

  const snapshotResult = await fromSnapshot.process.executeCommand('cat /home/daytona/test.txt')
  console.log(`File content from snapshot: ${snapshotResult.result.trim()}`)

  // Cleanup
  console.log('Cleaning up')
  await fromSnapshot.delete()
  await forked.delete()
  await sandbox.delete()
  console.log('Done')
}

main().catch(console.error)
