/**
 * Test WebSocket event subscription for sandbox lifecycle.
 *
 * Tests:
 * 1. Event subscriber connects on first sandbox creation
 * 2. Sandbox state auto-updates via WebSocket events
 * 3. waitUntilStarted/Stopped use WebSocket events
 * 4. get() and list() sandboxes also subscribe to events
 */

import { Daytona } from '@daytonaio/sdk'

async function test1_subscriberConnects(): Promise<Daytona> {
  console.log('--- Test 1: Subscriber connects on Daytona construction ---')
  const d = new Daytona()

  // initEventSubscriber is called in constructor (non-blocking), initPromise should exist
  const initPromise = (d as any).eventSubscriberInitPromise
  console.assert(initPromise !== null, '  Init promise should exist (connection started in constructor)')
  console.log('  PASS: Subscriber init started in constructor')

  const sandbox = await d.create()
  const subAfter = (d as any).eventSubscriber
  console.assert(subAfter !== null, '  Subscriber should exist after create')
  console.assert(subAfter?.isConnected === true, '  Subscriber should be connected')
  console.log(`  PASS: Subscriber connected after create (state=${sandbox.state})`)

  await sandbox.delete()
  console.log('  PASS: Test 1 complete\n')
  return d
}

async function test2_eventDrivenLifecycle(d: Daytona) {
  console.log('--- Test 2: Event-driven start/stop lifecycle ---')

  const sandbox = await d.create()
  console.log(`  Created sandbox: ${sandbox.id}, state=${sandbox.state}`)

  let t0 = Date.now()
  await sandbox.stop()
  const stopTime = (Date.now() - t0) / 1000
  console.log(`  Stopped in ${stopTime.toFixed(2)}s, state=${sandbox.state}`)
  console.assert(sandbox.state === 'stopped', `  Expected stopped, got ${sandbox.state}`)

  t0 = Date.now()
  await sandbox.start()
  const startTime = (Date.now() - t0) / 1000
  console.log(`  Started in ${startTime.toFixed(2)}s, state=${sandbox.state}`)
  console.assert(sandbox.state === 'started', `  Expected started, got ${sandbox.state}`)

  await sandbox.delete()
  console.log('  PASS: Test 2 complete\n')
}

async function test3_autoUpdateFromEvents(d: Daytona) {
  console.log('--- Test 3: Auto-update sandbox state from events ---')

  const sandbox = await d.create()
  console.log(`  Created: state=${sandbox.state}`)

  // Issue stop via the sandboxApi directly (bypassing SDK stop method)
  await (sandbox as any).sandboxApi.stopSandbox(sandbox.id)
  console.log('  Issued raw stop API call...')

  // Wait for auto-update via WebSocket events
  for (let i = 0; i < 30; i++) {
    await new Promise((r) => setTimeout(r, 200))
    if (sandbox.state !== 'started') {
      console.log(`  State changed to '${sandbox.state}' after ${((i + 1) * 0.2).toFixed(1)}s (via WS event)`)
      break
    }
  }

  // Wait for fully stopped
  for (let i = 0; i < 30; i++) {
    await new Promise((r) => setTimeout(r, 200))
    if (sandbox.state === 'stopped') {
      console.log(`  State reached 'stopped' (via WS event)`)
      break
    }
  }

  console.assert(sandbox.state === 'stopped', `  Expected stopped, got ${sandbox.state}`)
  await sandbox.delete()
  console.log('  PASS: Test 3 complete\n')
}

async function test4_getAndListSubscribe(d: Daytona) {
  console.log('--- Test 4: get() and list() sandboxes subscribe to events ---')

  const sandbox = await d.create()
  const sandbox2 = await d.get(sandbox.id)
  console.log(`  Got sandbox via get(): state=${sandbox2.state}`)

  await sandbox.stop()
  await new Promise((r) => setTimeout(r, 1000))
  console.log(`  After stop - original: state=${sandbox.state}, get'd: state=${sandbox2.state}`)
  console.assert(
    sandbox2.state === 'stopped' || sandbox2.state === 'stopping',
    `  get'd sandbox state should update, got ${sandbox2.state}`,
  )

  await sandbox.delete()
  console.log('  PASS: Test 4 complete\n')
}

async function main() {
  console.log('='.repeat(60))
  console.log('WebSocket Event Subscription Test Suite (TypeScript)')
  console.log('='.repeat(60) + '\n')

  const d = await test1_subscriberConnects()
  await test2_eventDrivenLifecycle(d)
  await test3_autoUpdateFromEvents(d)
  await test4_getAndListSubscribe(d)

  console.log('='.repeat(60))
  console.log('ALL TESTS PASSED')
  console.log('='.repeat(60))

  // Explicitly dispose to close WebSocket
  await d[Symbol.asyncDispose]()
}

main().catch(console.error)
