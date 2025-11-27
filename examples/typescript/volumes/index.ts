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

  // Mount a specific subpath within the volume
  // This is useful for isolating data or implementing multi-tenancy
  const mountDir3 = '/home/daytona/subpath'

  const sandbox3 = await daytona.create({
    language: 'typescript',
    volumes: [{ volumeId: volume.id, mountPath: mountDir3, subpath: 'users/alice' }],
  })

  // This sandbox will only see files within the 'users/alice' subpath
  // Create a file in the subpath
  const subpathFile = path.join(mountDir3, 'alice-file.txt')
  await sandbox3.fs.uploadFile(Buffer.from("Hello from Alice's subpath!"), subpathFile)

  // The file is stored at: volume-root/users/alice/alice-file.txt
  // but appears at: /home/daytona/subpath/alice-file.txt in the sandbox

  // Cleanup
  await daytona.delete(sandbox1)
  await daytona.delete(sandbox2)
  await daytona.delete(sandbox3)
  // await daytona.volume.delete(volume)
}

main()
