import { Daytona } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  // Default settings
  const sandbox1 = await daytona.create()
  console.log('networkBlockAll:', sandbox1.networkBlockAll)
  console.log('networkAllowList:', sandbox1.networkAllowList)

  // Block all network access
  const sandbox2 = await daytona.create({
    networkBlockAll: true,
  })
  console.log('networkBlockAll:', sandbox2.networkBlockAll)
  console.log('networkAllowList:', sandbox2.networkAllowList)

  // Explicitly allow list of network addresses
  const sandbox3 = await daytona.create({
    networkAllowList: '192.168.1.0/16,10.0.0.0/24',
  })
  console.log('networkBlockAll:', sandbox3.networkBlockAll)
  console.log('networkAllowList:', sandbox3.networkAllowList)

  await sandbox1.delete()
  await sandbox2.delete()
  await sandbox3.delete()
}

main().catch(console.error)
