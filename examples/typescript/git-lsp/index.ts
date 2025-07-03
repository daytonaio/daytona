import { Daytona, Image } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  //  first, create a sandbox
  const sandbox = await daytona.create(
    {
      image: Image.base('ubuntu:25.10').runCommands(
        'apt-get update && apt-get install -y --no-install-recommends nodejs npm coreutils',
        'curl -fsSL https://deb.nodesource.com/setup_20.x | bash -',
        'apt-get install -y nodejs',
        'npm install -g ts-node typescript typescript-language-server',
      ),
      language: 'typescript',
    },
    {
      onSnapshotCreateLogs: console.log,
    },
  )

  try {
    const projectDir = 'learn-typescript'

    //  clone the repository
    await sandbox.git.clone('https://github.com/panaverse/learn-typescript', projectDir, 'master')

    //  search for the file we want to work on
    const matches = await sandbox.fs.findFiles(projectDir, 'var obj1 = new Base();')
    console.log('Matches:', matches)

    //  start the language server
    const lsp = await sandbox.createLspServer('typescript', projectDir)
    await lsp.start()

    //  notify the language server of the document we want to work on
    await lsp.didOpen(matches[0].file!)

    //  get symbols in the document
    const symbols = await lsp.documentSymbols(matches[0].file!)
    console.log('Symbols:', symbols)

    //  fix the error in the document
    await sandbox.fs.replaceInFiles([matches[0].file!], 'var obj1 = new Base();', 'var obj1 = new E();')

    //  notify the language server of the document change
    await lsp.didClose(matches[0].file!)
    await lsp.didOpen(matches[0].file!)

    //  get completions at a specific position
    const completions = await lsp.completions(matches[0].file!, {
      line: 12,
      character: 18,
    })
    console.log('Completions:', completions)
  } catch (error) {
    console.error('Error creating sandbox:', error)
  } finally {
    //  cleanup
    await daytona.delete(sandbox)
  }
}

main()
