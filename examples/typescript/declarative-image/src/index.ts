import { Daytona, Image } from '@daytonaio/sdk'
import fs from 'fs'

async function main() {
  const daytona = new Daytona()

  // Generate unique name for the image to avoid conflicts
  const imageName = `node-example:${Date.now()}`
  console.log(`Creating image with name: ${imageName}`)

  // Create a local file with some data
  const localFilePath = 'file_example.txt'
  const localFileContent = 'Hello, World!'
  fs.writeFileSync(localFilePath, localFileContent)

  // Create a Python image with common data science packages
  const image = Image.debianSlim('3.12')
    .pipInstall(['numpy', 'pandas', 'matplotlib', 'scipy', 'scikit-learn'])
    .runCommands(['apt-get update && apt-get install -y git', 'mkdir -p /home/daytona/workspace'])
    .workdir('/home/daytona/workspace')
    .env({
      MY_ENV_VAR: 'My Environment Variable',
    })
    .addLocalFile(localFilePath, '/home/daytona/workspace/file_example.txt')

  // Create the image
  console.log(`=== Creating Image: ${imageName} ===`)
  await daytona.createImage(imageName, image, { onLogs: console.log })

  // Create first sandbox using the pre-built image
  console.log('\n=== Creating Sandbox from Pre-built Image ===')
  const sandbox1 = await daytona.create(
    {
      image: imageName,
    },
    {
      onImageBuildLogs: console.log,
    },
  )

  try {
    // Verify the first sandbox environment
    console.log('Verifying sandbox from pre-built image:')
    const nodeResponse = await sandbox1.process.executeCommand('python --version && pip list')
    console.log('Python environment:')
    console.log(nodeResponse.result)

    // Verify the file was added to the image
    const fileContent = await sandbox1.process.executeCommand('cat file_example.txt')
    console.log('File content:')
    console.log(fileContent.result)
  } finally {
    // Clean up first sandbox
    await daytona.remove(sandbox1)
  }

  // Create second sandbox with a new dynamic image
  console.log('\n=== Creating Sandbox with Dynamic Image ===')

  // Define a new dynamic image for the second sandbox
  const dynamicImage = Image.debianSlim('3.13')
    .pipInstall(['pytest', 'pytest-cov', 'black', 'isort', 'mypy', 'ruff'])
    .runCommands(['apt-get update && apt-get install -y git', 'mkdir -p /home/daytona/project'])
    .workdir('/home/daytona/project')
    .env({
      NODE_ENV: 'development',
    })

  // Create sandbox with the dynamic image
  const sandbox2 = await daytona.create(
    {
      image: dynamicImage,
    },
    {
      timeout: 0,
      onImageBuildLogs: console.log,
    },
  )

  try {
    // Verify the second sandbox environment
    console.log('Verifying sandbox with dynamic image:')
    const toolsResponse = await sandbox2.process.executeCommand('pip list | grep -E "pytest|black|isort|mypy|ruff"')
    console.log('Development tools:')
    console.log(toolsResponse.result)
  } finally {
    // Clean up second sandbox
    await daytona.remove(sandbox2)
  }
}

main().catch((error) => {
  console.error('Error:', error)
  process.exit(1)
})
