import { Daytona } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  const result = await daytona.snapshot.list(2, 10)
  console.log(`Found ${result.total} snapshots`)
  result.items.forEach((snapshot) => console.log(`${snapshot.name} (${snapshot.imageName})`))
}

main().catch(console.error)
