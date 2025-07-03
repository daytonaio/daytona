import { Daytona } from '@daytonaio/sdk'
import path from 'path'

async function main() {
  const daytona = new Daytona()

  //  Create a new volume or get an existing one
  const volume = await daytona.volume.get('my-volume', true)

  // Mount the volume to the sandbox
  const mountDir1 = '/home/daytona/volume'

  const sandbox1 = await daytona.create({
    language: 'typescript',
    volumes: [{ volumeId: volume.id, mountPath: mountDir1 }],
  })

  // Create a new directory in the mount directory
  const newDir = path.join(mountDir1, 'new-dir')
  await sandbox1.fs.createFolder(newDir, '755')

  // Create a new file in the mount directory
  const newFile = path.join(mountDir1, 'new-file.txt')
  await sandbox1.fs.uploadFile(Buffer.from('Hello, World!'), newFile)

  // Create a new sandbox with the same volume
  // and mount it to the different path
  const mountDir2 = '/home/daytona/my-files'

  const sandbox2 = await daytona.create({
    language: 'typescript',
    volumes: [{ volumeId: volume.id, mountPath: mountDir2 }],
  })

  // List files in the mount directory
  const files = await sandbox2.fs.listFiles(mountDir2)
  console.log('Files:', files)

  // Get the file from the first sandbox
  const file = await sandbox1.fs.downloadFile(newFile)
  console.log('File:', file.toString())

  // Cleanup
  await daytona.delete(sandbox1)
  await daytona.delete(sandbox2)
  // await daytona.volume.delete(volume)
}

main()
