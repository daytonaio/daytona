import { Daytona } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  // Auto-delete is disabled by default
  const sandbox1 = await daytona.create()
  console.log(sandbox1.autoDeleteInterval)

  // Auto-delete after the Sandbox has been stopped for 1 hour
  await sandbox1.setAutoDeleteInterval(60)
  console.log(sandbox1.autoDeleteInterval)

  // Delete immediately upon stopping
  await sandbox1.setAutoDeleteInterval(0)
  console.log(sandbox1.autoDeleteInterval)

  // Disable auto-delete
  await sandbox1.setAutoDeleteInterval(-1)
  console.log(sandbox1.autoDeleteInterval)

  // Auto-delete after the Sandbox has been stopped for 1 day
  const sandbox2 = await daytona.create({
    autoDeleteInterval: 1440,
  })
  console.log(sandbox2.autoDeleteInterval)
}

main().catch(console.error)
