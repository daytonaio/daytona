import { Daytona } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  // Default settings
  const sandbox1 = await daytona.create()
  console.log('networkAllowAll:', sandbox1.networkAllowAll)
  console.log('networkAllowList:', sandbox1.networkAllowList)

  // Explicitly allow all network access
  const sandbox2 = await daytona.create({
    networkAllowAll: true,
  })
  console.log('networkAllowAll:', sandbox2.networkAllowAll)
  console.log('networkAllowList:', sandbox2.networkAllowList)

  // Explicitly allow list of network addresses
  const sandbox3 = await daytona.create({
    networkAllowAll: false,
    networkAllowList: '192.168.1.0/24,10.0.0.0/24',
  })
  console.log('networkAllowAll:', sandbox3.networkAllowAll)
  console.log('networkAllowList:', sandbox3.networkAllowList)

  await sandbox1.delete()
  await sandbox2.delete()
  await sandbox3.delete()
}

main().catch(console.error)
