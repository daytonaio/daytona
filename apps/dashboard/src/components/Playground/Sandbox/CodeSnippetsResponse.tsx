/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Daytona,
  Image,
  CodeLanguage,
  CreateSandboxFromSnapshotParams,
  CreateSandboxFromImageParams,
} from '@daytonaio/sdk'
import { codeSnippetSupportedLanguages } from '@/enums/Playground'
import CodeBlock from '@/components/CodeBlock'
import { Button } from '@/components/ui/button'
import { usePlayground } from '@/hooks/usePlayground'
import { Loader2, Play } from 'lucide-react'
import { useAuth } from 'react-oidc-context'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { useMemo, useState } from 'react'
import ResponseCard from '../ResponseCard'

const SandboxCodeSnippetsResponse: React.FC = () => {
  const [codeSnippetLanguage, setCodeSnippetLanguage] = useState<CodeLanguage>(CodeLanguage.PYTHON)
  const [codeSnippetOutput, setCodeSnippetOutput] = useState<string>('')
  const [isCodeSnippetRunning, setIsCodeSnippetRunning] = useState<boolean>(false)

  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()
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
  const useConfigObject = false

  const useLanguageParam = sandboxParametersState['language']

  const useResources = objectHasAnyValue(sandboxParametersState['resources'])
  const useResourcesCPU = useResources && 'cpu' in sandboxParametersState['resources']
  const useResourcesMemory = useResources && 'memory' in sandboxParametersState['resources']
  const useResourcesDisk = useResources && 'disk' in sandboxParametersState['resources']

  const createSandboxParamsExist = objectHasAnyValue(sandboxParametersState['createSandboxBaseParams'])
  const useAutoStopInterval =
    createSandboxParamsExist && 'autoStopInterval' in sandboxParametersState['createSandboxBaseParams']
  const useAutoArchiveInterval =
    createSandboxParamsExist && 'autoArchiveInterval' in sandboxParametersState['createSandboxBaseParams']
  const useAutoDeleteInterval =
    createSandboxParamsExist && 'autoDeleteInterval' in sandboxParametersState['createSandboxBaseParams']

  const useSandboxCreateParams = useLanguageParam || useResources || createSandboxParamsExist

  const sandboxCodeSnippetsData = useMemo(() => {
    let pythonCodeSnippet = `from daytona import Daytona${useConfigObject ? ', DaytonaConfig' : ''}${useSandboxCreateParams ? ', CreateSandboxFromImageParams' : ''}${useResources ? ', Resources' : ''}, Image\n`
    let typeScriptCodeSnippet = `import { Daytona${useConfigObject ? ', DaytonaConfig' : ''}, Image } from '@daytonaio/sdk'\n`
    typeScriptCodeSnippet += '\nasync function main() {'
    if (useConfigObject) {
      pythonCodeSnippet += '\n# Define the configuration\n'
      typeScriptCodeSnippet += '\n\t// Define the configuration\n'
    }
    pythonCodeSnippet += `\n# Initialize the Daytona client\n`
    typeScriptCodeSnippet += `\n\t// Initialize the Daytona client\n`
    pythonCodeSnippet += `daytona = Daytona(${useConfigObject ? 'config' : ''})\n`
    typeScriptCodeSnippet += `\tconst daytona = new Daytona(${useConfigObject ? 'config' : ''})\n`
    if (useResources) {
      pythonCodeSnippet += '\n# Create a Sandbox with custom resources\nresources = Resources(\n'
      if (useResourcesCPU)
        pythonCodeSnippet += `\tcpu=${sandboxParametersState['resources']['cpu']},  # ${sandboxParametersState['resources']['cpu']} CPU cores\n`
      if (useResourcesMemory)
        pythonCodeSnippet += `\tmemory=${sandboxParametersState['resources']['memory']},  # ${sandboxParametersState['resources']['memory']}GB RAM\n`
      if (useResourcesDisk)
        pythonCodeSnippet += `\tdisk=${sandboxParametersState['resources']['disk']},  # ${sandboxParametersState['resources']['disk']}GB disk space\n`
      pythonCodeSnippet += ')\n'
    }
    if (useSandboxCreateParams) {
      pythonCodeSnippet += '\nparams = CreateSandboxFromImageParams(\n\timage=Image.debian_slim("3.12"),\n'
      if (useResources) pythonCodeSnippet += '\tresources=resources,\n'
      if (useLanguageParam) pythonCodeSnippet += `\tlanguage="${sandboxParametersState['language']}"\n`
      if (createSandboxParamsExist) {
        if (useAutoStopInterval)
          pythonCodeSnippet += `\tauto_stop_interval=${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']},\t # ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] == 0 ? 'Disables the auto-stop feature' : `Sandbox will be stopped after ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']} minute${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] > 1 ? 's' : ''}`}\n`
        if (useAutoArchiveInterval)
          pythonCodeSnippet += `\tauto_archive_interval=${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']},\t # Auto-archive after a Sandbox has been stopped for ${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval'] == 0 ? '30 days' : `${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']} minutes`}\n`
        if (useAutoDeleteInterval)
          pythonCodeSnippet += `\tauto_delete_interval=${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']},\t # ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] == 0 ? 'Sandbox will be deleted immediately after stopping' : sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval'] == -1 ? 'Auto-delete functionality disabled' : `Auto-delete after a Sandbox has been stopped for ${sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']} minutes`}\n`
      }
      pythonCodeSnippet += ')\n'
    }
    pythonCodeSnippet += `\n# Create the Sandbox instance\nsandbox = daytona.create(${useSandboxCreateParams ? 'params' : ''})\nprint(f"Sandbox created:{sandbox.id}")\n`
    typeScriptCodeSnippet += '\ttry {\n'
    typeScriptCodeSnippet += `\t\t// Create the Sandbox instance\n\t\tconst sandbox = await daytona.create(${useSandboxCreateParams ? '{\n\t\t\timage: Image.debianSlim("3.13"),\n' : ''}`
    if (useResources) {
      typeScriptCodeSnippet += '\t\t\tresources: {\n'
      if (useResourcesCPU)
        typeScriptCodeSnippet += `\t\t\t\tcpu: ${sandboxParametersState['resources']['cpu']},  // ${sandboxParametersState['resources']['cpu']} CPU cores\n`
      if (useResourcesMemory)
        typeScriptCodeSnippet += `\t\t\t\tmemory: ${sandboxParametersState['resources']['memory']},  // ${sandboxParametersState['resources']['memory']}GB RAM\n`
      if (useResourcesDisk)
        typeScriptCodeSnippet += `\t\t\t\tdisk: ${sandboxParametersState['resources']['disk']},  // ${sandboxParametersState['resources']['disk']}GB disk space\n`
      typeScriptCodeSnippet += '\t\t\t},\n'
    }
    if (useLanguageParam) typeScriptCodeSnippet += `\t\t\tlanguage: '${sandboxParametersState['language']}',\n`
    if (createSandboxParamsExist) {
      if (useAutoStopInterval)
        typeScriptCodeSnippet += `\t\t\tautoStopInterval: ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']},\t // ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] == 0 ? 'Disables the auto-stop feature' : `Sandbox will be stopped after ${sandboxParametersState['createSandboxBaseParams']['autoStopInterval']} minute${sandboxParametersState['createSandboxBaseParams']['autoStopInterval'] > 1 ? 's' : ''}`}\n`
      if (useAutoArchiveInterval)
        typeScriptCodeSnippet += `\t\t\tautoArchiveInterval: ${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']},\t // Auto-archive after a Sandbox has been stopped for ${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval'] == 0 ? '30 days' : `${sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']} minutes`}\n`
      if (useAutoDeleteInterval)
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
    useResourcesCPU,
    useResourcesMemory,
    useResourcesDisk,
    createSandboxParamsExist,
    useAutoStopInterval,
    useAutoArchiveInterval,
    useAutoDeleteInterval,
    useSandboxCreateParams,
    sandboxParametersState,
  ])

  const runCodeSnippet = async () => {
    setIsCodeSnippetRunning(true)
    setCodeSnippetOutput('Running code...')

    try {
      const daytona = new Daytona({
        jwtToken: user?.access_token,
        apiUrl: import.meta.env.VITE_API_URL,
        organizationId: selectedOrganization?.id,
      })
      let createSandboxFromImageParams: CreateSandboxFromImageParams
      const createSandboxFromSnapshotParams: CreateSandboxFromSnapshotParams = { snapshot: 'vnc-snapshot' }
      const createSandboxFromImage = useSandboxCreateParams
      if (createSandboxFromImage) {
        // Set CreateSandboxFromImageParams specific params
        createSandboxFromImageParams = { image: Image.debianSlim('3.13') }
        if (useResources) {
          createSandboxFromImageParams.resources = {}
          if (useResourcesCPU) createSandboxFromImageParams.resources.cpu = sandboxParametersState['resources']['cpu']
          if (useResourcesMemory)
            createSandboxFromImageParams.resources.memory = sandboxParametersState['resources']['memory']
          if (useResourcesDisk)
            createSandboxFromImageParams.resources.disk = sandboxParametersState['resources']['disk']
        }
      }
      const createSandboxParams: CreateSandboxFromImageParams | CreateSandboxFromSnapshotParams = createSandboxFromImage
        ? createSandboxFromImageParams
        : createSandboxFromSnapshotParams
      // Set CreateSandboxBaseParams params whch are common for both params types
      if (useLanguageParam) createSandboxParams.language = sandboxParametersState['language']
      if (useAutoStopInterval)
        createSandboxParams.autoStopInterval = sandboxParametersState['createSandboxBaseParams']['autoStopInterval']
      if (useAutoArchiveInterval)
        createSandboxParams.autoArchiveInterval =
          sandboxParametersState['createSandboxBaseParams']['autoArchiveInterval']
      if (useAutoDeleteInterval)
        createSandboxParams.autoDeleteInterval = sandboxParametersState['createSandboxBaseParams']['autoDeleteInterval']
      createSandboxParams.labels = { 'daytona-playground': 'true' }
      if (useLanguageParam)
        createSandboxParams.labels['daytona-playground-language'] = sandboxParametersState['language']
      const sandbox = await daytona.create(createSandboxParams)
      setCodeSnippetOutput(
        `Sandbox successfully created: ${sandbox.id}${useLanguageParam ? '\nRunning code inside sandbox...' : ''}`,
      )
      if (useLanguageParam) {
        const codeRunResponse = await sandbox.process.codeRun(
          createSandboxParams.language === CodeLanguage.PYTHON ? 'print("Hello World!")' : 'console.log("Hello world")',
        )
        setCodeSnippetOutput(codeRunResponse.result)
      }
    } catch (error) {
      console.error(error)
      setCodeSnippetOutput(`Error: ${error instanceof Error ? error.message : String(error)}`)
    } finally {
      setIsCodeSnippetRunning(false)
    }
  }

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
              <Button disabled={isCodeSnippetRunning} variant="outline" className="ml-auto" onClick={runCodeSnippet}>
                {isCodeSnippetRunning ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="w-4 h-4" />} Run
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
      <ResponseCard responseText={codeSnippetOutput} />
    </>
  )
}

export default SandboxCodeSnippetsResponse
