import { Daytona } from '@daytona/sdk'

async function main() {
  const daytona = new Daytona()

  for await (const sandbox of daytona.list({
    limit: 10,
    labels: { env: 'dev' },
    states: ['started'],
    sort: 'createdAt',
    order: 'desc',
  })) {
    console.log(sandbox.id)
  }
}

main().catch(console.error)
