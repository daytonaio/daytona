---
title: "PtyHandle"
hideTitleOnPage: true
---


## PtyHandle

PTY session handle for managing a single PTY session.

**Properties**:

- `sessionId` _string_



Provides methods for sending input, resizing the terminal, waiting for completion,
and managing the WebSocket connection to a PTY session.

**Example:**

```typescript
// Create a PTY session
const ptyHandle = await process.createPty({
  id: 'my-session',
  cols: 120,
  rows: 30,
  onData: (data) => {
    const text = new TextDecoder().decode(data);
    process.stdout.write(text);
  },
});

// Send commands
await ptyHandle.sendInput('ls -la\n');
await ptyHandle.sendInput('exit\n');

// Wait for completion
const result = await ptyHandle.wait();
console.log(`PTY exited with code: ${result.exitCode}`);

// Clean up
await ptyHandle.disconnect();
```

### Accessors

#### error

##### Get Signature

```ts
get error(): string
```

Error message if the PTY failed

**Returns**:

- `string`

***

#### exitCode

##### Get Signature

```ts
get exitCode(): number
```

Exit code of the PTY process (if terminated)

**Returns**:

- `number`

### Constructors

#### new PtyHandle()

```ts
new PtyHandle(
   ws: WebSocket, 
   handleResize: (cols: number, rows: number) => Promise<PtySessionInfo>, 
   handleKill: () => Promise<void>, 
   onPty: (data: Uint8Array) => void | Promise<void>, 
   sessionId: string): PtyHandle
```

**Parameters**:

- `ws` _WebSocket_
- `handleResize` _\(cols: number, rows: number\) =\> Promise\<PtySessionInfo\>_
- `handleKill` _\(\) =\> Promise\<void\>_
- `onPty` _\(data: Uint8Array\) =\> void \| Promise\<void\>_
- `sessionId` _string_


**Returns**:

- `PtyHandle`

### Methods

#### disconnect()

```ts
disconnect(): Promise<void>
```

Disconnect from the PTY session and clean up resources.

Closes the WebSocket connection and releases any associated resources.
Should be called when done with the PTY session.

**Returns**:

- `Promise<void>`

**Example:**

```ts
// Always clean up when done
try {
  // ... use PTY session
} finally {
  await ptyHandle.disconnect();
}
```

***

#### isConnected()

```ts
isConnected(): boolean
```

Check if connected to the PTY session

**Returns**:

- `boolean`

***

#### kill()

```ts
kill(): Promise<void>
```

Kill the PTY process and terminate the session.

Forcefully terminates the PTY session and its associated process.
This operation is irreversible and will cause the PTY to exit immediately.

**Returns**:

- `Promise<void>`

**Throws**:

If the kill operation fails

**Example:**

```ts
// Kill a long-running process
await ptyHandle.kill();

// Wait to confirm termination
const result = await ptyHandle.wait();
console.log(`Process terminated with exit code: ${result.exitCode}`);
```

***

#### resize()

```ts
resize(cols: number, rows: number): Promise<PtySessionInfo>
```

Resize the PTY terminal dimensions.

Changes the terminal size which will notify terminal applications
about the new dimensions via SIGWINCH signal.

**Parameters**:

- `cols` _number_ - New number of terminal columns
- `rows` _number_ - New number of terminal rows


**Returns**:

- `Promise<PtySessionInfo>`

**Example:**

```ts
// Resize to 120x30
await ptyHandle.resize(120, 30);
```

***

#### sendInput()

```ts
sendInput(data: string | Uint8Array<ArrayBufferLike>): Promise<void>
```

Send input data to the PTY session.

Sends keyboard input or commands to the terminal session. The data will be
processed as if it was typed in the terminal.

**Parameters**:

- `data` _Input data to send \(commands, keystrokes, etc.\)_ - `string` | `Uint8Array`\<`ArrayBufferLike`\>


**Returns**:

- `Promise<void>`

**Throws**:

If PTY is not connected or sending fails

**Example:**

```ts
// Send a command
await ptyHandle.sendInput('ls -la\n');

// Send raw bytes
await ptyHandle.sendInput(new Uint8Array([3])); // Ctrl+C
```

***

#### wait()

```ts
wait(): Promise<PtyResult>
```

Wait for the PTY process to exit and return the result.

This method blocks until the PTY process terminates and returns
information about how it exited.

**Returns**:

- `Promise<PtyResult>` - Result containing exit code and error information

**Example:**

```ts
// Wait for process to complete
const result = await ptyHandle.wait();

if (result.exitCode === 0) {
  console.log('Process completed successfully');
} else {
  console.log(`Process failed with code: ${result.exitCode}`);
  if (result.error) {
    console.log(`Error: ${result.error}`);
  }
}
```

***

#### waitForConnection()

```ts
waitForConnection(): Promise<void>
```

Wait for the WebSocket connection to be established.

This method ensures the PTY session is ready to receive input and send output.
It waits for the server to confirm the connection is established.

**Returns**:

- `Promise<void>`

**Throws**:

If connection times out (10 seconds) or connection fails