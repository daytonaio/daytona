import { Daytona } from '@daytonaio/sdk'
import { inspect } from 'util'

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

  const response = await existingSandbox.process.executeCommand('echo "Hello World from exec!"', '/home/daytona', 10)
  if (response.exitCode !== 0) {
    console.error(`Error: ${response.exitCode} ${response.result}`)
  } else {
    console.log(response.result)
  }

  const sandboxes = await daytona.list()
  console.log('Total sandboxes count:', sandboxes.length)
  // Use util.inspect to pretty print the sandbox info like Python's pprint
  console.log(inspect(await sandboxes[0].info(), { depth: null, colors: true }))

  console.log('Removing sandbox')
  await sandbox.delete()
  console.log('Sandbox removed')
}

main().catch(console.error)
