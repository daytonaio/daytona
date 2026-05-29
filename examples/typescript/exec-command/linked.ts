import { Daytona } from '@daytona/sdk'

async function main() {
  const daytona = new Daytona()

  const owner = await daytona.create()
  console.log(`Owner sandbox ready: id=${owner.id} name=${owner.name}`)

  // Linked sandboxes must be ephemeral — `ephemeral: true` sets
  // `autoDeleteInterval=0` automatically.
  const follower = await daytona.create({
    linkedSandbox: owner.id,
    ephemeral: true,
  })
  console.log(`Follower sandbox ready: id=${follower.id} name=${follower.name}`)
  console.log(`  follower.linkedSandboxId = ${follower.linkedSandboxId}`)

  try {
    // Background the http server with nohup, then poll locally until it
    // binds — so the follower's curl below doesn't race startup.
    console.log(`\nStarting \`python3 -m http.server 3000\` in owner '${owner.name}'`)
    const startScript = `set -e
mkdir -p /tmp/lnk
echo 'hello from owner' > /tmp/lnk/index.html
cd /tmp/lnk
nohup python3 -m http.server 3000 > /tmp/lnk/srv.log 2>&1 &
for _ in $(seq 1 20); do
  if curl -sS --max-time 1 http://127.0.0.1:3000/ >/dev/null 2>&1; then
    echo READY
    exit 0
  fi
  sleep 0.5
done
echo "server failed to start"
cat /tmp/lnk/srv.log
exit 1
`
    const startRes = await owner.process.executeCommand(startScript, undefined, undefined, 30)
    if (startRes.exitCode !== 0) {
      throw new Error(`Failed to start server in owner: ${startRes.result}`)
    }
    console.log(startRes.result.trim())

    // The link network registers the owner under its sandbox name as a DNS
    // alias, so the follower can reach it by name.
    console.log(`\nReaching '${owner.name}' from the follower over the link network`)
    const curlRes = await follower.process.executeCommand(
      `curl -sS --max-time 5 http://${owner.name}:3000/`,
      undefined,
      undefined,
      10,
    )
    if (curlRes.exitCode !== 0) {
      throw new Error(`Follower could not reach owner: exit=${curlRes.exitCode} output=${curlRes.result}`)
    }
    console.log(`Response from owner: ${curlRes.result.trim()}`)
  } finally {
    console.log(`\nDeleting follower ${follower.id}`)
    await daytona.delete(follower)
    console.log(`Deleting owner ${owner.id}`)
    await daytona.delete(owner)
  }
}

main()
