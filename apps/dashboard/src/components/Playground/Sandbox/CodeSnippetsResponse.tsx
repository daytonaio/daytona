/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { CodeLanguage } from '@daytonaio/sdk-typescript/src'
import { codeSnippetSupportedLanguages } from '@/enums/Playground'
import CodeBlock from '@/components/CodeBlock'
import { usePlaygroundSandboxParams } from './hook'
import { useMemo, useState } from 'react'

const SandboxCodeSnippetsResponse: React.FC = () => {
  const [codeSnippetLanguage, setCodeSnippetLanguage] = useState<CodeLanguage>(CodeLanguage.PYTHON)

  const { playgroundSandboxParametersState } = usePlaygroundSandboxParams()

  const sandboxCodeSnippetsData = useMemo(
    () => ({
      [CodeLanguage.PYTHON]: {
        //TODO -> OVDE CE IC PUNI KOD I KRAJ SVAKOG DIJELA KOJI OVISI O UNESENON PARAMETRU CE BIT CONDITIONAL RENDER -> TO CE BIT DEPENDENCIES U useMemo
        code: `
from daytona import Daytona, DaytonaConfig

# Define the configuration
config = DaytonaConfig(api_key="your-api-key")

# Initialize the Daytona client
daytona = Daytona(config)

# Create the Sandbox instance
sandbox = daytona.create()

# Run the code securely inside the Sandbox
response = sandbox.process.code_run('print("Hello World from code!")')
if response.exit_code != 0:
print(f"Error: {response.exit_code} {response.result}")
else:
  print(response.result)
from daytona import Daytona, DaytonaConfig

# Define the configuration
config = DaytonaConfig(api_key="your-api-key")

# Initialize the Daytona client
daytona = Daytona(config)

# Create the Sandbox instance
sandbox = daytona.create()

# Run the code securely inside the Sandbox
response = sandbox.process.code_run('print("Hello World from code!")')
if response.exit_code != 0:
print(f"Error: {response.exit_code} {response.result}")
else:
  print(response.result)
      `,
      },
      [CodeLanguage.TYPESCRIPT]: {
        code: `import { Daytona } from '@daytonaio/sdk'
    
// Initialize the Daytona client
const daytona = new Daytona({ apiKey: 'your-api-key' });

// Create the Sandbox instance
const sandbox = await daytona.create({
  language: 'typescript',
});

// Run the code securely inside the Sandbox
const response = await sandbox.process.codeRun('console.log("Hello World from code!")')
console.log(response.result);
`,
      },
      [CodeLanguage.JAVASCRIPT]: { code: '' }, // Currently to prevent ts error when indexing
    }),
    [],
  )
  return (
    <>
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Code</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs
            value={codeSnippetLanguage}
            className="flex flex-col"
            onValueChange={(languageValue) => setCodeSnippetLanguage(languageValue as CodeLanguage)}
          >
            <TabsList className="px-4 w-full flex-shrink-0">
              {codeSnippetSupportedLanguages.map((language) => (
                <TabsTrigger
                  key={language.value}
                  value={language.value}
                  className={codeSnippetLanguage === language.value ? 'bg-foreground/10' : ''}
                >
                  <div className="flex items-center text-sm">
                    <img src={language.icon} alt={language.icon} className="w-4 h-4" />
                    <span className="ml-2">{language.label}</span>
                  </div>
                </TabsTrigger>
              ))}
            </TabsList>
            {codeSnippetSupportedLanguages.map((language) => (
              <TabsContent key={language.value} value={language.value}>
                <CodeBlock
                  language={language.value}
                  code={sandboxCodeSnippetsData[language.value].code}
                  codeAreaClassName="overflow-y-scroll h-[350px]"
                />
              </TabsContent>
            ))}
          </Tabs>
        </CardContent>
      </Card>
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Response</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg">
            <pre className="max-w-full bg-zinc-900 text-zinc-100 h-[250px] p-4 rounded-lg overflow-x-auto overflow-y-auto text-sm font-mono">
              <code>{`Terminal output test text
Terminal output test text 2`}</code>
            </pre>
          </div>
        </CardContent>
      </Card>
    </>
  )
}

export default SandboxCodeSnippetsResponse
