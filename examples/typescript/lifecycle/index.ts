import { Daytona, SandboxListSortDirection, SandboxListSortField, SandboxState } from '@daytona/sdk'

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

  for await (const sb of daytona.list({
    limit: 10,
    labels: { env: 'dev' },
    states: [SandboxState.STARTED],
    sort: SandboxListSortField.CREATED_AT,
    order: SandboxListSortDirection.DESC,
  })) {
    console.log(sb.id)
  }

  console.log('Deleting sandbox')
  await sandbox.delete()
  console.log('Sandbox deleted')
}

main().catch(console.error)
