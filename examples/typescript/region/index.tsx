import { Daytona, Image } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona({
    target: 'us',
  })

  const snapshot1 = `us-${Date.now()}`
  console.log(`Creating snapshot ${snapshot1}`)
  try {
    await daytona.snapshot.create({
      name: snapshot1,
      image: Image.debianSlim('3.12'),
      regionId: 'us',
    })
  } catch (error: any) {
    console.error(error?.message)
  }
  console.log('--------------------------------')

  const snapshot2 = `eu-${Date.now()}`
  console.log(`Creating snapshot ${snapshot2}`)
  try {
    await daytona.snapshot.create({
      name: snapshot2,
      image: Image.debianSlim('3.13'),
      regionId: 'eu',
    })
  } catch (error: any) {
    console.error('error', error?.message)
  }
  console.log('--------------------------------')

  console.log(`Creating sandbox from snapshot ${snapshot1}`)
  try {
    const sandbox = await daytona.create({
      snapshot: snapshot1,
    })
    await daytona.delete(sandbox)
  } catch (error: any) {
    console.error(error?.message)
  }
  console.log('--------------------------------')

  console.log(`Creating sandbox from snapshot ${snapshot2}`)
  try {
    const sandbox = await daytona.create({
      snapshot: snapshot2,
    })
    await daytona.delete(sandbox)
  } catch (error: any) {
    console.error('error', error?.message)
  }
}

main().catch(console.error)
