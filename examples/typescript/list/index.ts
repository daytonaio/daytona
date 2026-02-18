import { Daytona, SandboxState } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  const limit = 2
  const states = [SandboxState.STARTED, SandboxState.STOPPED]

  const page1 = await daytona.list({
    limit,
    states,
  })
  console.log('Listing page 1')
  for (const sandbox of page1.items) {
    console.log(`${sandbox.id}: ${sandbox.state}`)
  }

  if (page1.nextCursor) {
    const page2 = await daytona.list({
      cursor: page1.nextCursor,
      limit,
      states,
    })
    console.log('Listing page 2')
    for (const sandbox of page2.items) {
      console.log(`${sandbox.id}: ${sandbox.state}`)
    }
  }
}

main().catch(console.error)
