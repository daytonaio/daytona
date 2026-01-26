import { Daytona } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  console.log('Creating sandbox')
  const sandbox = await daytona.create()
  console.log('Sandbox created')

  await sandbox.setLabels({
    public: 'true',
  })

  console.log('Stopping sandbox')
  await sandbox.stop()
  console.log('Sandbox stopped')

  console.log('Starting sandbox')
  await sandbox.start()
  console.log('Sandbox started')

  console.log('Getting existing sandbox')
  const existingSandbox = await daytona.get(sandbox.id)
  console.log('Got existing sandbox')

  const response = await existingSandbox.process.executeCommand(
    'echo "Hello World from exec!"',
    undefined,
    undefined,
    10,
  )
  if (response.exitCode !== 0) {
    console.error(`Error: ${response.exitCode} ${response.result}`)
  } else {
    console.log(response.result)
  }

  const result = await daytona.list()
  console.log('Total sandboxes count:', result.total)

  console.log(`Printing first sandbox -> id: ${result.items[0].id} state: ${result.items[0].state}`)

  // Hot resize: increase CPU and memory on a running sandbox
  console.log('Resizing sandbox (hot resize)...')
  await sandbox.resize({ cpu: 2, memory: 2 }, true)
  console.log(`Hot resize complete: CPU=${sandbox.cpu}, Memory=${sandbox.memory}GB, Disk=${sandbox.disk}GB`)

  // Cold resize: stop sandbox first, then resize (can also change disk)
  console.log('Stopping sandbox for cold resize...')
  await sandbox.stop()
  console.log('Resizing sandbox (cold resize)...')
  await sandbox.resize({ cpu: 4, memory: 4, disk: 20 }, false)
  console.log(`Cold resize complete: CPU=${sandbox.cpu}, Memory=${sandbox.memory}GB, Disk=${sandbox.disk}GB`)
  await sandbox.start()
  console.log('Sandbox restarted with new resources')

  console.log('Deleting sandbox')
  await sandbox.delete()
  console.log('Sandbox deleted')
}

main().catch(console.error)
