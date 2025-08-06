import React, { useMemo, useState } from 'react'
import Editor from '@monaco-editor/react'
import { Button } from '../components/ui/button'
import { Card, CardContent } from '../components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select'
import { Label } from '../components/ui/label'
import { Daytona } from '@daytonaio/sdk'
import '../components/monaco'
import { useAuth } from 'react-oidc-context'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { CodeLanguage, SAMPLES } from '@/playground-samples'

const files = import.meta.glob('../../../../dist/libs/sdk-typescript/**/*.d.ts', {
  eager: true,
  query: '?raw',
  import: 'default',
})

const Playground: React.FC = () => {
  const [language, setLanguage] = useState<CodeLanguage>(CodeLanguage.TypeScript)
  const [languageSamples, setLanguageSamples] = useState<string[]>(Object.keys(SAMPLES[CodeLanguage.TypeScript]))
  const [selectedSample, setSelectedSample] = useState<string>('Default')

  const [tsCode, setTsCode] = useState<string>(SAMPLES[CodeLanguage.TypeScript].Default)
  const [pyCode, setPyCode] = useState<string>(SAMPLES[CodeLanguage.Python].Default)
  const [bashCode, setBashCode] = useState<string>(SAMPLES[CodeLanguage.Bash].Default)
  const [output, setOutput] = useState<string>('')
  const [isRunning, setIsRunning] = useState<boolean>(false)
  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()

  const handleReload = () => {
    switch (language) {
      case CodeLanguage.TypeScript:
        handleSampleChange('Default', language)
        break
      case CodeLanguage.Python:
        handleSampleChange('Default', language)
        break
      case CodeLanguage.Bash:
        handleSampleChange('Default', language)
        break
    }
  }

  const handleLanguageChange = (value: string) => {
    setLanguage(value as CodeLanguage)
    setLanguageSamples(Object.keys(SAMPLES[value as CodeLanguage]))
    handleSampleChange('Default', value as CodeLanguage)
  }

  const handleSampleChange = (value: string, language: CodeLanguage) => {
    setSelectedSample(value)
    switch (language) {
      case CodeLanguage.TypeScript:
        setTsCode(SAMPLES[language][value])
        break
      case CodeLanguage.Python:
        setPyCode(SAMPLES[language][value])
        break
      case CodeLanguage.Bash:
        setBashCode(SAMPLES[language][value])
        break
    }
  }

  const handleCodeChange = (value: string | undefined) => {
    if (value !== undefined) {
      switch (language) {
        case CodeLanguage.TypeScript:
          setTsCode(value)
          break
        case CodeLanguage.Python:
          setPyCode(value)
          break
        case CodeLanguage.Bash:
          setBashCode(value)
          break
      }
    }
  }

  const editorValue = useMemo(() => {
    switch (language) {
      case CodeLanguage.TypeScript:
        return tsCode
      case CodeLanguage.Python:
        return pyCode
      case CodeLanguage.Bash:
        return bashCode
      default:
        return ''
    }
  }, [language, tsCode, pyCode, bashCode])

  const runCode = async () => {
    setIsRunning(true)
    setOutput('Running code...')

    try {
      const daytona = new Daytona({
        jwtToken: user?.access_token,
        apiUrl: import.meta.env.VITE_API_URL,
        organizationId: selectedOrganization?.id,
      })

      const sandbox = await daytona.create({
        language: language === CodeLanguage.Bash ? 'python' : language,
        labels: {
          'daytona-playground': 'true',
          'daytona-playground-language': language,
          'daytona-playground-sample': selectedSample,
        },
      })

      if (language === CodeLanguage.Bash) {
        await sandbox.process.createSession('exec-session')
        const response = await sandbox.process.executeSessionCommand('exec-session', {
          command: bashCode,
        })
        setOutput(response.output ?? 'No output')
      } else {
        const response = await sandbox.process.codeRun(language === 'typescript' ? tsCode : pyCode)
        setOutput(response.result)
      }

      await sandbox.delete()
    } catch (error) {
      console.error(error)
      setOutput(`Error: ${error instanceof Error ? error.message : String(error)}`)
    } finally {
      setIsRunning(false)
    }
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Playground</h1>
        <p className="text-muted-foreground mt-2">Try out executing code in a Daytona sandbox.</p>
      </div>

      <div className="flex items-center gap-4">
        <div className="space-y-2">
          <Label htmlFor="language-select">Language</Label>
          <Select value={language} onValueChange={handleLanguageChange}>
            <SelectTrigger id="language-select" className="w-[180px]">
              <SelectValue placeholder="Select language" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="typescript">TypeScript</SelectItem>
              <SelectItem value="python">Python</SelectItem>
              <SelectItem value="bash">Bash</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label htmlFor="sample-select">Sample</Label>
          <Select value={selectedSample} onValueChange={(value) => handleSampleChange(value, language)}>
            <SelectTrigger id="sample-select" className="w-[180px]">
              <SelectValue placeholder="Select sample" />
            </SelectTrigger>
            <SelectContent>
              {languageSamples.map((sample) => (
                <SelectItem key={sample} value={sample}>
                  {sample}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <Button onClick={runCode} disabled={isRunning} className="mt-auto">
          {isRunning ? 'Running...' : 'Run Code'}
        </Button>

        <Button onClick={handleReload} className="mt-auto">
          Reload Example
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="space-y-2">
          <h2 className="text-xl font-semibold">Editor</h2>
          <Card>
            <CardContent className="p-0">
              <div className="h-[400px] rounded-md overflow-hidden">
                <Editor
                  onMount={(editor, monaco) => {
                    let mergedContext = ''

                    Object.entries(files).forEach(async ([key, value]) => {
                      // remove the import lines so all the definitions are in the same file
                      const content = (value as string).replace(/^import .*;$/gm, '')
                      mergedContext += content
                    })

                    monaco.languages.typescript.typescriptDefaults.addExtraLib(
                      `declare module '@daytonaio/sdk' {
                         ${mergedContext}
                       }`,
                    )

                    monaco.languages.typescript.typescriptDefaults.setCompilerOptions({
                      ...monaco.languages.typescript.typescriptDefaults.getCompilerOptions(),
                      target: monaco.languages.typescript.ScriptTarget.ESNext,
                      module: monaco.languages.typescript.ModuleKind.ESNext,
                    })
                  }}
                  height="400px"
                  language={language}
                  value={editorValue}
                  onChange={handleCodeChange}
                  theme="vs-dark"
                  options={{
                    minimap: { enabled: false },
                    fontSize: 14,
                  }}
                />
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-2">
          <h2 className="text-xl font-semibold">Output</h2>
          <Card>
            <CardContent className="p-0">
              <div className="h-[400px] bg-black text-white font-mono p-4 overflow-auto rounded-md">
                <pre className="m-0">{output}</pre>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <div className="text-sm text-muted-foreground">
        <p>Note: Running code will create a new sandbox and delete it after the code has finished executing.</p>
      </div>
    </div>
  )
}

export default Playground
