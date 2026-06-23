import { Daytona, Image, Sandbox } from '@daytona/sdk'

/**
 * Runs `curl` against a URL inside the sandbox and reports whether the request
 * made it out through the sandbox's domain allow list or was blocked.
 */
async function checkAccess(sandbox: Sandbox, url: string) {
  const res = await sandbox.process.executeCommand(`curl -sS --max-time 10 -o /dev/null -w "%{http_code}" ${url}`)
  const allowed = res.exitCode === 0
  const detail = allowed ? `HTTP ${res.result.trim()}` : `exit ${res.exitCode}`
  console.log(`  ${allowed ? '✅ allowed' : '⛔ blocked'}  ${url}  (${detail})`)
}

async function main() {
  const daytona = new Daytona()

  // A domain allow list is a comma-separated list of domains the sandbox is
  // allowed to reach. Everything else is blocked. Wildcards match subdomains:
  //   - google.com    → the apex domain
  //   - *.google.com  → www.google.com, mail.google.com, ...
  const domainAllowList = 'google.com,*.google.com'

  // The image just needs curl so we can demonstrate the allow list in action.
  const sandbox = await daytona.create(
    {
      image: Image.base('ubuntu:22.04').runCommands(
        'apt-get update',
        'apt-get install -y --no-install-recommends curl ca-certificates',
      ),
      domainAllowList,
    },
    {
      timeout: 200,
      onSnapshotCreateLogs: console.log,
    },
  )

  try {
    console.log('domainAllowList:', sandbox.domainAllowList)

    console.log(`\nWith allow list "${domainAllowList}":`)
    await checkAccess(sandbox, 'https://www.google.com') // matches *.google.com
    await checkAccess(sandbox, 'https://google.com') // matches google.com
    await checkAccess(sandbox, 'https://example.com') // not on the list -> blocked

    // The allow list can also be changed at runtime without restarting the
    // sandbox. Here we swap it out so only example.com is reachable.
    console.log('\nUpdating allow list to "example.com":')
    await sandbox.updateNetworkSettings({ domainAllowList: 'example.com' })
    console.log('domainAllowList:', sandbox.domainAllowList)

    await checkAccess(sandbox, 'https://example.com') // now allowed
    await checkAccess(sandbox, 'https://www.google.com') // now blocked
  } finally {
    // cleanup
    await daytona.delete(sandbox)
  }
}

main().catch(console.error)
