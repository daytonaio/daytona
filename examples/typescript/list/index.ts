import { Daytona, SandboxState } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  const statesFilter = [SandboxState.STARTED, SandboxState.STOPPED]

  const page1 = await daytona.listV2(undefined, 2, statesFilter)
  for (const sandbox of page1.items) {
    console.log(`${sandbox.id}: ${sandbox.state}`)
  }

  if (page1.nextCursor) {
    const page2 = await daytona.listV2(page1.nextCursor, 2, statesFilter)
    for (const sandbox of page2.items) {
      console.log(`${sandbox.id}: ${sandbox.state}`)
    }
  }
}

main().catch(console.error)
