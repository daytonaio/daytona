import { Daytona } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  const result = await daytona.list({ 'my-label': 'my-value' }, 2, 10)
  for (const sandbox of result.items) {
    console.log(`${sandbox.id}: ${sandbox.state}`)
  }
}

main().catch(console.error)
