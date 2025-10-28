#!/usr/bin/env node
/**
 * WebSocket client that connects to the Go server and handles close code 4008
 */

import WebSocket from 'ws'

async function testWebSocketClient(): Promise<void> {
  const uri = 'ws://localhost:8088/ws'

  console.log(`Connecting to ${uri}...`)

  const ws = new WebSocket(uri)

  // Connection opened
  ws.on('open', () => {
    console.log('Connected successfully!')
  })

  // Listen for messages
  ws.on('message', (data: WebSocket.Data) => {
    console.log(`Received message: ${data}`)
  })

  // Listen for close event
  ws.on('close', (code: number, reason: Buffer) => {
    console.log('\n🔴 Connection closed!')
    console.log(`   Close code: ${code}`)
    console.log(`   Close reason: ${reason.toString()}`)

    if (code === 4008) {
      console.log('   ✅ Successfully received close code 4008!')
    } else {
      console.log(`   ❌ Expected 4008, but got ${code}`)
    }
  })

  // Listen for errors
  ws.on('error', (error: Error) => {
    console.error(`Error: ${error.name}: ${error.message}`)
  })
}

// Main execution
console.log('='.repeat(60))
console.log('WebSocket Client (TypeScript) - Testing Close Code 4008')
console.log('='.repeat(60))

testWebSocketClient().catch((error) => {
  console.error('Unhandled error:', error)
  process.exit(1)
})

// Keep the process alive for a bit to receive the close event
setTimeout(() => {
  console.log('='.repeat(60))
  process.exit(0)
}, 5000)
