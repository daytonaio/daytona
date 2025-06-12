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

  const sandboxes = await daytona.list()
  console.log('Total sandboxes count:', sandboxes.length)

  console.log(`Printing sandboxes[0] -> id: ${sandboxes[0].id} state: ${sandboxes[0].state}`)

  console.log('Deleting sandbox')
  await sandbox.delete()
  console.log('Sandbox deleted')
}

main().catch(console.error)
