import { useState } from 'react'
import { Tabs, TabsList, TabsTrigger } from './ui/tabs'
import pythonIcon from '../assets/python.svg'
import typescriptIcon from '../assets/typescript.svg'
import CodeBlock from './CodeBlock'

const GettingStarted: React.FC = () => {
  const [language, setLanguage] = useState<'typescript' | 'python'>('python')

  return (
    <div className="min-h-screen p-8">
      <div className="max-w-3xl mx-auto">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-2xl font-bold mb-2">Get Started</h1>
            <p className="text-muted-foreground">Install and get your Sandboxes running.</p>
          </div>
          <div className="flex items-center space-x-2">
            <Tabs value={language} onValueChange={(value) => setLanguage(value as 'typescript' | 'python')}>
              <TabsList className="bg-foreground/10">
                <TabsTrigger value="python">
                  <img src={pythonIcon} alt="Python" className="w-4 h-4" />
                </TabsTrigger>
                <TabsTrigger value="typescript">
                  <img src={typescriptIcon} alt="TypeScript" className="w-4 h-4" />
                </TabsTrigger>
              </TabsList>
            </Tabs>
          </div>
        </div>

        <div className="relative">
          {/* Timeline line */}
          <div className="absolute left-[15px] top-[40px] bottom-0 w-[2px] bg-muted-foreground/50" />

          {/* Steps */}
          <div className="space-y-12">
            {/* Step 1 */}
            <div className="relative pl-12">
              <div className="absolute left-0 w-8 h-8 text-background rounded-full bg-muted-foreground flex items-center justify-center text-sm">
                1
              </div>
              <h2 className="text-xl font-semibold mb-4">Install the SDK</h2>
              <p className="mb-4">Run the following command in your terminal to install the Daytona SDK:</p>
              <div className="transition-all duration-500">
                <CodeBlock
                  code={codeExamples[language].install}
                  language={language === 'typescript' ? 'bash' : 'bash'}
                  showCopy
                />
              </div>
            </div>

            {/* Step 2 */}
            <div className="relative pl-12">
              <div className="absolute left-0 w-8 h-8 text-background rounded-full bg-muted-foreground flex items-center justify-center text-sm">
                2
              </div>
              <h2 className="text-xl font-semibold mb-4">Create your first Sandbox</h2>
              <p className="mb-4">The example below will create a Sandbox and run a simple code snippet:</p>
              <div className="transition-all duration-500">
                <CodeBlock code={codeExamples[language].example} language={language} showCopy />
              </div>
            </div>

            {/* Step 3 */}
            <div className="relative pl-12">
              <div className="absolute left-0 w-8 h-8 text-background rounded-full bg-muted-foreground flex items-center justify-center text-sm">
                3
              </div>
              <h2 className="text-xl font-semibold mb-4">That's it</h2>
              <p className="text-muted-foreground">
                It's as easy as that. For more examples check out the{' '}
                <a href="https://daytona.io/docs" target="_blank" rel="noopener noreferrer" className="text-primary">
                  Docs
                </a>
                .
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

const codeExamples = {
  typescript: {
    install: `npm install @daytonaio/sdk`,
    example: `import { Daytona } from '@daytonaio/sdk'

// Initialize the Daytona client
const daytona = new Daytona({
  apiKey: 'your-api-key',
})

// Create the sandbox instance
const sandbox = await daytona.create()

// Run some python code securely inside the sandbox
const response = await sandbox.process.codeRun('print("Hello World!")')
console.log(response.result)

// Run a shell command securely inside the sandbox
const cmdResult = await sandbox.process.executeCommand('echo "Hello World from CMD!"')
console.log(cmdResult.result)

//  add a new file to the sandbox
const fileContent = new File([Buffer.from('Hello, World!')], 'data.txt', {
type: 'text/plain',
})
await sandbox.fs.uploadFile('/home/daytona/data.txt', fileContent)

// delete the sandbox
await sandbox.delete()
`,
  },
  python: {
    install: `pip install daytona-sdk`,
    example: `from daytona_sdk import Daytona, DaytonaConfig

# Initialize the Daytona client
config = DaytonaConfig(api_key="your-api-key")
daytona = Daytona(config)

# Create the sandbox instance
sandbox = daytona.create()

# Run the code securely inside the sandbox
response = sandbox.process.code_run('print("Hello World!")')
print(response.result)

# Execute an os command in the sandbox
response = sandbox.process.exec('echo "Hello World from exec!"', cwd="/home/daytona", timeout=10)
print(response.result)

# Add a new file to the sandbox
file_content = b"Hello, World!"
sandbox.fs.upload_file("/home/daytona/data.txt", file_content)

# delete the sandbox
daytona.remove(sandbox)
`,
  },
}

export default GettingStarted
