import { Daytona, Image } from '@daytona/sdk'

function section(title: string) {
  console.log(`\n=== ${title} ===`)
}

async function main() {
  const daytona = new Daytona()

  // Custom image with a TypeScript language server (for the LSP showcase) and git.
  const sandbox = await daytona.create(
    {
      image: Image.base('ubuntu:25.10').runCommands(
        'apt-get update && apt-get install -y --no-install-recommends nodejs npm coreutils git',
        'curl -fsSL https://deb.nodesource.com/setup_20.x | bash -',
        'apt-get install -y nodejs',
        'npm install -g ts-node typescript typescript-language-server',
      ),
      language: 'typescript',
    },
    {
      timeout: 200,
      onSnapshotCreateLogs: console.log,
    },
  )

  try {
    const git = sandbox.git
    const proc = sandbox.process
    const repo = 'demo-repo'

    // ----------------------------- Git operations -----------------------------
    console.log('git version:', (await proc.executeCommand('git --version')).result.trim())

    section('init')
    await git.init(repo, false, 'main')
    console.log('initialized repo at', repo)

    section('configureUser + getConfig (local scope)')
    await git.configureUser('Ada Lovelace', 'ada@example.com', 'local', repo)
    console.log('user.name  =', await git.getConfig('user.name', 'local', repo))
    console.log('user.email =', await git.getConfig('user.email', 'local', repo))

    section('setConfig / getConfig (local) + unset key')
    await git.setConfig('core.editor', 'nano', 'local', repo)
    console.log('core.editor     =', await git.getConfig('core.editor', 'local', repo))
    console.log('user.signingkey =', await git.getConfig('user.signingkey', 'local', repo), '(unset -> undefined)')

    section('remoteAdd / remotes / remoteGet')
    await git.remoteAdd(repo, 'origin', 'https://github.com/panaverse/learn-typescript.git')
    console.log(
      'remotes       =',
      (await git.remotes(repo)).remotes.map((r) => [r.name, r.url]),
    )
    console.log('remoteGet     =', await git.remoteGet(repo, 'origin'))
    console.log('remoteGet(?)  =', await git.remoteGet(repo, 'upstream'), '(missing -> undefined)')

    section('add / commit')
    await sandbox.fs.uploadFile(Buffer.from('line1\n'), `${repo}/a.txt`)
    await git.add(repo, ['a.txt'])
    const commit = await git.commit(repo, 'first commit', 'Ada Lovelace', 'ada@example.com')
    console.log('commit sha =', commit.sha)

    section('branches (current marker)')
    const branches = await git.branches(repo)
    console.log('branches =', branches.branches, '| current =', branches.current)

    section('status (detached / upstream / current)')
    const status = await git.status(repo)
    console.log(
      `current_branch=${status.currentBranch} detached=${status.detached} ` +
        `upstream=${JSON.stringify(status.upstream)} ahead=${status.ahead} behind=${status.behind}`,
    )

    section('createBranch + deleteBranch')
    await git.createBranch(repo, 'feature')
    await git.checkoutBranch(repo, 'main')
    await git.deleteBranch(repo, 'feature')
    console.log("deleted branch 'feature'")

    section('reset (mixed) -> unstage')
    await sandbox.fs.uploadFile(Buffer.from('staged\n'), `${repo}/b.txt`)
    await git.add(repo, ['b.txt'])
    console.log(
      'staged before reset:',
      (await git.status(repo)).fileStatus.map((f) => [f.name, f.staging]),
    )
    await git.reset(repo)
    console.log(
      'staged after reset :',
      (await git.status(repo)).fileStatus.map((f) => [f.name, f.staging]),
    )

    section('restore (worktree) -> discard local changes')
    await sandbox.fs.uploadFile(Buffer.from('corrupted\n'), `${repo}/a.txt`)
    console.log('a.txt before restore:', (await proc.executeCommand('cat a.txt', repo)).result.trim())
    await git.restore(repo, ['a.txt'])
    console.log('a.txt after restore :', (await proc.executeCommand('cat a.txt', repo)).result.trim())

    section('reset (keep)')
    await sandbox.fs.uploadFile(Buffer.from('v2\n'), `${repo}/a.txt`)
    await git.add(repo, ['a.txt'])
    await git.commit(repo, 'second commit', 'Ada Lovelace', 'ada@example.com')
    await git.reset(repo, 'keep', 'HEAD~1')
    console.log('a.txt after keep reset to HEAD~1:', (await proc.executeCommand('cat a.txt', repo)).result.trim())

    section('clone (shallow, depth=1)')
    await git.clone(
      'https://github.com/panaverse/learn-typescript',
      'shallow',
      'master',
      undefined,
      undefined,
      undefined,
      undefined,
      1,
    )
    console.log(
      'shallow clone commit count (expect 1) =',
      (await proc.executeCommand('git rev-list --count HEAD', 'shallow')).result.trim(),
    )

    section('pull (remote + branch)')
    await git.pull('shallow', undefined, undefined, 'master', 'origin')
    console.log('pulled origin/master into shallow clone (already up to date)')

    section('dangerouslyAuthenticate')
    await git.dangerouslyAuthenticate('ci-bot', 'ghp_faketoken', 'example.com')
    console.log('credential.helper (global) =', await git.getConfig('credential.helper', 'global'))

    console.log('\nAll new git operations exercised successfully ✅')

    // --------------------------------- LSP -----------------------------------
    const projectDir = 'learn-typescript'

    section('clone project for LSP')
    //  clone the repository
    await git.clone('https://github.com/panaverse/learn-typescript', projectDir, 'master')

    //  search for the file we want to work on
    const matches = await sandbox.fs.findFiles(projectDir, 'var obj1 = new Base();')
    console.log('Matches:', matches)

    section('LSP: document symbols + completions')
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
    console.error('Error executing example:', error)
    throw error
  } finally {
    await daytona.delete(sandbox)
  }
}

main().catch(() => process.exit(1))
