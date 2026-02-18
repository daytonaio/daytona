import { Daytona } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  const listWithDefaults = await daytona.list()
  console.log('listWithDefaults')
  console.log(listWithDefaults.total, listWithDefaults.page, listWithDefaults.totalPages)

  for (const sandbox of listWithDefaults.items) {
    console.log(`${sandbox.id}: ${sandbox.state}`)
  }

  const listWithCustomParams = await daytona.list(undefined, 1, 10)
  for (const sandbox of listWithCustomParams.items) {
    console.log(`${sandbox.id}: ${sandbox.state}`)
  }

  if (listWithCustomParams.totalPages > listWithCustomParams.page) {
    const nextPage = await daytona.list(undefined, listWithCustomParams.page + 1, 10)
    for (const sandbox of nextPage.items) {
      console.log(`${sandbox.id}: ${sandbox.state}`)
    }
  }
}

main().catch(console.error)
