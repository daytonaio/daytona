import React, { useMemo, useState, useCallback, useEffect } from 'react'
import Editor from '@monaco-editor/react'
import { Button } from '../components/ui/button'
import { Card, CardContent } from '../components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select'
import { Label } from '../components/ui/label'
import { Daytona, SandboxTsCodeToolbox, SandboxPythonCodeToolbox } from '@daytonaio/sdk'
import '../components/monaco'
import { useAuth } from 'react-oidc-context'
import { useSelectedOrganization } from '@/hooks/useSelectedOrganization'
import { CodeLanguage, SAMPLES } from '@/playground-samples'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { useXTerm } from 'react-xtermjs'
import '@xterm/xterm/css/xterm.css'

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
  const [isRunning, setIsRunning] = useState<boolean>(false)
  const { user } = useAuth()
  const { selectedOrganization } = useSelectedOrganization()

  const { instance: terminal, ref: terminalRef } = useXTerm()

  const fitAddon = new FitAddon()
  const webLinksAddon = new WebLinksAddon()

  useEffect(() => {
    // Load the fit addon
    terminal?.loadAddon(fitAddon)
    terminal?.loadAddon(webLinksAddon)

    fitAddon.fit()

    // Hide cursor
    terminal?.write('\x1b[?25l')

    const handleResize = () => fitAddon.fit()

    // Handle resize event
    window.addEventListener('resize', handleResize)
    return () => {
      window.removeEventListener('resize', handleResize)
      terminal?.dispose()
      fitAddon.dispose()
      webLinksAddon.dispose()
    }
  }, [terminalRef, terminal])

  const clearTerminal = useCallback(() => {
    if (terminal) {
      terminal.clear()
    }
  }, [terminal])

  const handleSampleChange = useCallback((value: string, language: CodeLanguage) => {
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
  }, [])

  const handleReload = useCallback(() => {
    clearTerminal()
    switch (language) {
      case CodeLanguage.TypeScript:
        handleSampleChange(selectedSample, language)
        break
      case CodeLanguage.Python:
        handleSampleChange(selectedSample, language)
        break
      case CodeLanguage.Bash:
        handleSampleChange(selectedSample, language)
        break
    }
  }, [clearTerminal, language, handleSampleChange, selectedSample])

  const handleLanguageChange = useCallback(
    (value: string) => {
      clearTerminal()
      setLanguage(value as CodeLanguage)
      setLanguageSamples(Object.keys(SAMPLES[value as CodeLanguage]))
      handleSampleChange('Default', value as CodeLanguage)
    },
    [clearTerminal, handleSampleChange],
  )

  const handleCodeChange = useCallback(
    (value: string | undefined) => {
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
    },
    [language],
  )

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

  const runCode = useCallback(async () => {
    if (!terminal) return

    setIsRunning(true)

    // Clear terminal and show running message
    terminal.clear()
    terminal.write('Running code...\r\n')

    try {
      const daytona = new Daytona({
        jwtToken: user?.access_token,
        apiUrl: import.meta.env.VITE_API_URL,
        organizationId: selectedOrganization?.id,
      })

      terminal.write('Creating sandbox...\r\n')
      const sandbox = await daytona.create({
        language: language === CodeLanguage.Bash ? 'python' : language,
        labels: {
          'daytona-playground': 'true',
          'daytona-playground-language': language,
          'daytona-playground-sample': selectedSample,
        },
        autoDeleteInterval: 0,
      })

      terminal.write('Sandbox created. Executing code...\r\n\r\n')

      let command = ''
      switch (language) {
        case CodeLanguage.Bash:
          command = bashCode
          break
        case CodeLanguage.TypeScript:
          command = new SandboxTsCodeToolbox().getRunCommand(tsCode)
          break
        case CodeLanguage.Python:
          command = new SandboxPythonCodeToolbox().getRunCommand(pyCode)
          break
      }

      await sandbox.process.createSession('exec-session')
      const response = await sandbox.process.executeSessionCommand('exec-session', {
        command,
        runAsync: true,
      })
      if (response.cmdId) {
        await sandbox.process.getSessionCommandLogs(
          'exec-session',
          response.cmdId,
          (stdout) => {
            terminal.write(stdout.replace('\n', '\r\n'))
          },
          (stderr) => {
            terminal.write(stderr.replace('\n', '\r\n'))
          },
        )
      }

      terminal.write('\r\n\r\nCleaning up sandbox...\r\n')
      await sandbox.delete()
      terminal.write('Done!\r\n')
    } catch (error) {
      console.error(error)
      terminal.write(`\r\nError: ${error instanceof Error ? error.message : String(error)}\r\n`)
    } finally {
      setIsRunning(false)
    }
  }, [terminal, language, tsCode, pyCode, bashCode, selectedSample, user?.access_token, selectedOrganization?.id])

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

        <Button onClick={clearTerminal} variant="outline" className="mt-auto">
          Clear Terminal
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 h-full">
        <div className="space-y-2 h-full">
          <h2 className="text-xl font-semibold">Editor</h2>
          <Card>
            <CardContent className="p-0">
              <div className="h-[500px] rounded-md overflow-hidden">
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

                    editor.layout()
                  }}
                  height="500px"
                  // height="10px"
                  language={language}
                  value={editorValue}
                  onChange={handleCodeChange}
                  theme="vs-dark"
                  options={{
                    minimap: { enabled: false },
                    fontSize: 14,
                    automaticLayout: true,
                  }}
                />
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-2 h-full">
          <h2 className="text-xl font-semibold">Output</h2>
          <Card>
            <CardContent className="p-0 h-[500px]">
              <div ref={terminalRef} className="h-full w-full rounded-md overflow-hidden" />
            </CardContent>
          </Card>
        </div>

        <div className="text-sm text-muted-foreground col-span-2">
          <p>Note: Running code will create a new sandbox and delete it after the code has finished executing.</p>
        </div>
      </div>
    </div>
  )
}

export default Playground
