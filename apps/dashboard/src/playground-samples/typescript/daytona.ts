export const daytona_typescript = `import { Daytona } from '@daytonaio/sdk'

// Initialize the Daytona client
const daytona = new Daytona()

// Create the sandbox instance
const sandbox = await daytona.create()

// Run some python code securely inside the sandbox
const response = await sandbox.process.codeRun('print("Hello World!")')
console.log(response.result)

// Run a shell command securely inside the sandbox
const cmdResult = await sandbox.process.executeCommand('echo "Hello World from CMD!"')
console.log(cmdResult.result)

// delete the sandbox
await sandbox.delete()
`
