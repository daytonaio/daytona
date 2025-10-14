import { Daytona } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  console.log('🚀 Starting Disk Management Example')
  console.log('=====================================')

  try {
    // List all existing disks
    console.log('\n📋 Listing all disks...')
    const existingDisks = await daytona.disk.list()
    console.log(`Found ${existingDisks.length} existing disks:`)
    existingDisks.forEach((disk) => {
      console.log(`  - ${disk.name} (${disk.id}) - ${disk.size}GB - State: ${disk.state}`)
    })

    // Create a new disk
    console.log('\n💾 Creating a new disk...')
    const diskName = `example-disk-${Date.now()}`
    const diskSize = 20 // 20GB
    const disk = await daytona.disk.create(diskName, diskSize)
    console.log(`✅ Created disk: ${disk.name} (${disk.id}) - ${disk.size}GB - State: ${disk.state}`)

    // Get the disk by ID
    console.log('\n🔍 Getting disk details...')
    const retrievedDisk = await daytona.disk.get(disk.id)
    console.log(`✅ Retrieved disk: ${retrievedDisk.name} - ${retrievedDisk.size}GB - State: ${retrievedDisk.state}`)

    // List disks again to see the new one
    console.log('\n📋 Listing disks after creation...')
    const updatedDisks = await daytona.disk.list()
    console.log(`Found ${updatedDisks.length} disks:`)
    updatedDisks.forEach((d) => {
      console.log(`  - ${d.name} (${d.id}) - ${d.size}GB - State: ${d.state}`)
    })

    // Wait a moment before deletion
    console.log('\n⏳ Waiting 2 seconds before cleanup...')
    await new Promise((resolve) => setTimeout(resolve, 2000))

    // Delete the disk
    console.log('\n🗑️  Deleting the disk...')
    await daytona.disk.delete(disk)
    console.log(`✅ Deleted disk: ${disk.name}`)

    // Final list to confirm deletion
    console.log('\n📋 Final disk list...')
    const finalDisks = await daytona.disk.list()
    console.log(`Found ${finalDisks.length} disks after cleanup`)

    console.log('\n🎉 Disk management example completed successfully!')
  } catch (error) {
    console.error('❌ Error during disk management:', error)
    process.exit(1)
  }
}

main()
