/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { CodeLanguage } from '@daytonaio/sdk'
import { codeSnippetSupportedLanguages } from '@/enums/Playground'
import CodeBlock from '@/components/CodeBlock'
import { Button } from '@/components/ui/button'
import { usePlayground } from '@/hooks/usePlayground'
import { Play } from 'lucide-react'
import { useMemo, useState } from 'react'
import ResponseCard from '../ResponseCard'

const SandboxCodeSnippetsResponse: React.FC = () => {
  const [codeSnippetLanguage, setCodeSnippetLanguage] = useState<CodeLanguage>(CodeLanguage.PYTHON)

  const { sandboxParametersState } = usePlayground()

  const objectHasAnyValue = (obj: object) => Object.values(obj).some((v) => v !== '' && v !== undefined)
  const indentString = (string: string, indentationCount: number) => {
    let indentationString = ''
    for (let i = 0; i < indentationCount; i++) indentationString += '\t'
    return string
      .split('\n')
      .map((line) => indentationString + line)
      .join('\n')
  }

  sandboxParametersState['languageCodeToRun'] = `function greet(name: string): string {
    return \`Hello, \${name}!\`;
}
console.log(greet("Daytona"));`
  sandboxParametersState['shellCodeToRun'] = 'ls -la'
  const useConfigObject = sandboxParametersState['apiKey']
  const useLanguageParam = sandboxParametersState['language']
  const useResources = objectHasAnyValue(sandboxParametersState['resources'])
  const createSandboxParamsExist = objectHasAnyValue(sandboxParametersState['createSandboxBaseParams'])
  const useSandboxCreateParams = useLanguageParam || useResources || createSandboxParamsExist

  const sandboxCodeSnippetsData = useMemo(() => {
    let pythonCodeSnippet = `from daytona import Daytona${useConfigObject ? ', DaytonaConfig' : ''}${useSandboxCreateParams ? ', CreateSandboxFromImageParams' : ''}${useResources ? ', Resources' : ''}, Image\n`
    let typeScriptCodeSnippet = `import { Daytona${useConfigObject ? ', DaytonaConfig' : ''}, Image } from '@daytonaio/sdk'\n`
    typeScriptCodeSnippet += '\nasync function main() {'
    if (useConfigObject) {
      pythonCodeSnippet += '\n# Define the configuration\n'
      pythonCodeSnippet += `config = DaytonaConfig(api_key="${sandboxParametersState['apiKey']}")\n`
      typeScriptCodeSnippet += '\n\t// Define the configuration\n'
      typeScriptCodeSnippet += `\tconst config: DaytonaConfig = { apiKey: "${sandboxParametersState['apiKey']}" }\n`
    }
    pythonCodeSnippet += `\n# Initialize the Daytona client\n`
    typeScriptCodeSnippet += `\n\t// Initialize the Daytona client\n`
    pythonCodeSnippet += `daytona = Daytona(${useConfigObject ? 'config' : ''})\n`
    typeScriptCodeSnippet += `\tconst daytona = new Daytona(${useConfigObject ? 'config' : ''})\n`
    if (useResources) {
      pythonCodeSnippet += '\n# Create a Sandbox with custom resources\nresources = Resources(\n'
      if (sandboxParametersState['resources']['cpu'])
        pythonCodeSnippet += `\tcpu=${sandboxParametersState['resources']['cpu']},  # ${sandboxParametersState['resources']['cpu']} CPU cores\n`
      if (sandboxParametersState['resources']['memory'])
        pythonCodeSnippet += `\tmemory=${sandboxParametersState['resources']['memory']},  # ${sandboxParametersState['resources']['memory']}GB RAM\n`
      if (sandboxParametersState['resources']['disk'])
        pythonCodeSnippet += `\tdisk=${sandboxParametersState['resources']['disk']},  # ${sandboxParametersState['resources']['disk']}GB disk space\n`
      pythonCodeSnippet += ')\n'
    }
    if (useSandboxCreateParams) {
      pythonCodeSnippet += '\nparams = CreateSandboxFromImageParams(\n\timage=Image.debian_slim("3.12"),\n'
      if (useResources) pythonCodeSnippet += '\tresources=resources,\n'
      if (useLanguageParam) pythonCodeSnippet += `\tlanguage="${sandboxParametersState['language']}"\n`
      if (createSandboxParamsExist) {
        if (sandboxParametersState['createSandboxBaseParams']['autoStopInterval'])
          pythonCodeSnippet += `\tauto_stop_interval=${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']},\t # ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] == 0 ? 'Disables the auto-stop feature' : `Sandbox will be stopped after ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']} minute${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] > 1 ? 's' : ''}`}\n`
        if (sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval'])
          pythonCodeSnippet += `\tauto_archive_interval=${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']},\t # Auto-archive after a Sandbox has been stopped for ${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval'] == 0 ? '30 days' : `${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']} minutes`}\n`
        if (sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'])
          pythonCodeSnippet += `\tauto_delete_interval=${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']},\t # ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] == 0 ? 'Sandbox will be deleted immediately after stopping' : sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] == -1 ? 'Auto-delete functionality disabled' : `Auto-delete after a Sandbox has been stopped for ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']} minutes`}\n`
      }
      pythonCodeSnippet += ')\n'
    }
    pythonCodeSnippet += `\n# Create the Sandbox instance\nsandbox = daytona.create(${useSandboxCreateParams ? 'params' : ''})\nprint(f"Sandbox created:{sandbox.id}")\n`
    typeScriptCodeSnippet += '\ttry {\n'
    typeScriptCodeSnippet += `\t\t// Create the Sandbox instance\n\t\tconst sandbox = await daytona.create(${useSandboxCreateParams ? '{\n\t\t\timage: Image.debianSlim("3.13"),\n' : ''}`
    if (useResources) {
      typeScriptCodeSnippet += '\t\t\tresources: {\n'
      if (sandboxParametersState['resources']['cpu'])
        typeScriptCodeSnippet += `\t\t\t\tcpu: ${sandboxParametersState['resources']['cpu']},  // ${sandboxParametersState['resources']['cpu']} CPU cores\n`
      if (sandboxParametersState['resources']['memory'])
        typeScriptCodeSnippet += `\t\t\t\tmemory: ${sandboxParametersState['resources']['memory']},  // ${sandboxParametersState['resources']['memory']}GB RAM\n`
      if (sandboxParametersState['resources']['disk'])
        typeScriptCodeSnippet += `\t\t\t\tdisk: ${sandboxParametersState['resources']['disk']},  // ${sandboxParametersState['resources']['disk']}GB disk space\n`
      typeScriptCodeSnippet += '\t\t\t},\n'
    }
    if (useLanguageParam) typeScriptCodeSnippet += `\t\t\tlanguage: '${sandboxParametersState['language']}',\n`
    if (createSandboxParamsExist) {
      if (sandboxParametersState['createSandboxBaseParams']['autoStopInterval'])
        typeScriptCodeSnippet += `\t\t\tautoStopInterval: ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']},\t // ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] == 0 ? 'Disables the auto-stop feature' : `Sandbox will be stopped after ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']} minute${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] > 1 ? 's' : ''}`}\n`
      if (sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval'])
        typeScriptCodeSnippet += `\t\t\tautoArchiveInterval: ${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']},\t // Auto-archive after a Sandbox has been stopped for ${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval'] == 0 ? '30 days' : `${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']} minutes`}\n`
      if (sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'])
        typeScriptCodeSnippet += `\t\t\tautoDeleteInterval: ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']},\t // ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] == 0 ? 'Sandbox will be deleted immediately after stopping' : sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] == -1 ? 'Auto-delete functionality disabled' : `Auto-delete after a Sandbox has been stopped for ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']} minutes`}\n`
    }
    if (useSandboxCreateParams) typeScriptCodeSnippet += '\t\t})\n'
    else typeScriptCodeSnippet += ')\n'
    if (sandboxParametersState['languageCodeToRun']) {
      pythonCodeSnippet += '\n# Run code securely inside the Sandbox\n'
      pythonCodeSnippet += `response = sandbox.process.code_run('\n${indentString(sandboxParametersState.languageCodeToRun, 1)}\n')\nif response.exit_code != 0:\n\tprint(f"Error: {response.exit_code} {response.result}")\nelse:\n\tprint(response.result)\n`
      typeScriptCodeSnippet += `\n\t\t// Run code securely inside the Sandbox\n`
      typeScriptCodeSnippet += `\t\tconst response = await sandbox.process.codeRun('\n${indentString(sandboxParametersState.languageCodeToRun, 3)}\n\t\t')\n\t\tif (response.exitCode !== 0) {\n\t\t\tconsole.error("Error running code:", response.exitCode, response.result)\n\t\t} else {\n\t\t\tconsole.log(response.result)\n\t\t}\n`
    }
    if (sandboxParametersState['shellCodeToRun']) {
      pythonCodeSnippet += '\n# Execute shell commands\n'
      pythonCodeSnippet += `response = sandbox.process.exec("${sandboxParametersState['shellCodeToRun']}")\nprint(response.result)\n`
      typeScriptCodeSnippet += '\n\t\t// Execute shell commands\n'
      typeScriptCodeSnippet += `\t\tconst response = await sandbox.process.executeCommand('${sandboxParametersState['shellCodeToRun']}')\n\t\tconsole.log(response.result)\n`
    }
    typeScriptCodeSnippet += '\t} catch (error) {\n\t\tconsole.error("Sandbox flow error:", error)\n\t}\n'
    typeScriptCodeSnippet += '}\n'
    typeScriptCodeSnippet += '\nmain().catch(console.error)'
    return {
      [CodeLanguage.PYTHON]: {
        code: pythonCodeSnippet,
      },
      [CodeLanguage.TYPESCRIPT]: {
        code: typeScriptCodeSnippet,
      },
      [CodeLanguage.JAVASCRIPT]: { code: '' }, // Currently to prevent ts error when indexing
    }
  }, [
    useConfigObject,
    useLanguageParam,
    useResources,
    createSandboxParamsExist,
    useSandboxCreateParams,
    sandboxParametersState,
  ])
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
              <Button variant="outline" className="ml-auto">
                <Play className="w-4 h-4" /> Run
              </Button>
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
      <ResponseCard responseText="Response test text" />
    </>
  )
}

export default SandboxCodeSnippetsResponse
