---
title: VNC Access
---

import { TabItem, Tabs } from '@astrojs/starlight/components'

VNC (Virtual Network Computing) access provides a graphical desktop environment for your Daytona Sandbox directly in the browser. This allows you to interact with GUI applications, desktop tools, and visual interfaces running inside your sandbox.

VNC and [Computer Use](/docs/en/computer-use) work together to enable both manual and automated desktop interactions. VNC provides the visual interface for users to manually interact with the desktop, while Computer Use provides the programmatic API for AI agents to automate mouse, keyboard, and screenshot operations. Through VNC, you can observe AI agents performing automated tasks via Computer Use in real-time.

- **GUI application development**: build and test desktop applications with visual interfaces
- **Browser testing**: run and debug web applications in a full browser environment
- **Visual debugging**: inspect graphical output and UI behavior in real-time
- **Desktop tool access**: use graphical IDEs, design tools, or other desktop software
- **Agent observation**: watch AI agents perform automated tasks through Computer Use

:::note[Sandbox image requirement]
VNC and Computer Use require a sandbox with the default image. Sandboxes created with custom images do not include VNC support unless you install the [required packages](#required-packages).
:::

## Access VNC from Dashboard

Access the VNC desktop environment directly from the [Daytona Dashboard ↗](https://app.daytona.io/dashboard/sandboxes).

1. Navigate to [Daytona Sandboxes ↗](https://app.daytona.io/dashboard/sandboxes)
2. Locate the sandbox you want to access via VNC
3. Click the options menu (**⋮**) next to the sandbox
4. Select **VNC** from the dropdown menu

This opens a VNC viewer in your browser with a **Connect** button.

5. Click **Connect** to establish the VNC session

Once connected, a full desktop environment loads in your browser, providing mouse and keyboard control over the sandbox's graphical interface.

:::note
VNC sessions remain active as long as the sandbox is running. If the sandbox auto-stops due to inactivity, you need to start the sandbox again before reconnecting via VNC.
:::

## Programmatic VNC management

Daytona provides methods to [start](#start-vnc), [stop](#stop-vnc), and [monitor](#get-vnc-status) VNC sessions and processes programmatically using the [Computer Use](/docs/en/computer-use) references as part of automated workflows.

### Start VNC

Start all VNC processes (Xvfb, xfce4, x11vnc, novnc) in the sandbox to enable desktop access.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
result = sandbox.computer_use.start()
print("VNC processes started:", result.message)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const result = await sandbox.computerUse.start();
console.log('VNC processes started:', result.message);
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
result = sandbox.computer_use.start
puts "VNC processes started: #{result.message}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
err := sandbox.ComputerUse.Start(ctx)
if err != nil {
	log.Fatal(err)
}
defer sandbox.ComputerUse.Stop(ctx)

fmt.Println("VNC processes started")
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/computeruse/start' \
  --request POST
```

</TabItem>
</Tabs>

For more information, see the [Computer Use](/docs/en/computer-use#start-computer-use) reference.

> [**start (Python SDK)**](/docs/en/python-sdk/sync/computer-use#computerusestart)
>
> [**start (TypeScript SDK)**](/docs/en/typescript-sdk/computer-use#start)
>
> [**start (Ruby SDK)**](/docs/en/ruby-sdk/computer-use#start)
>
> [**Start (Go SDK)**](/docs/en/go-sdk/daytona#ComputerUseService.Start)
>
> [**Start Computer Use Processes (API)**](/docs/en/tools/api/#daytona-toolbox/tag/computer-use/POST/computeruse/start)

### Stop VNC

Stop all VNC processes in the sandbox.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
result = sandbox.computer_use.stop()
print("VNC processes stopped:", result.message)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const result = await sandbox.computerUse.stop();
console.log('VNC processes stopped:', result.message);
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
result = sandbox.computer_use.stop
puts "VNC processes stopped: #{result.message}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
err := sandbox.ComputerUse.Stop(ctx)
if err != nil {
	log.Fatal(err)
}

fmt.Println("VNC processes stopped")
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/computeruse/stop' \
  --request POST
```

</TabItem>
</Tabs>

For more information, see the [Computer Use](/docs/en/computer-use#stop-computer-use) reference.

> [**stop (Python SDK)**](/docs/en/python-sdk/sync/computer-use#computerusestop)
>
> [**stop (TypeScript SDK)**](/docs/en/typescript-sdk/computer-use#stop)
>
> [**stop (Ruby SDK)**](/docs/en/ruby-sdk/computer-use#stop)
>
> [**Stop (Go SDK)**](/docs/en/go-sdk/daytona#ComputerUseService.Stop)
>
> [**Stop Computer Use Processes (API)**](/docs/en/tools/api/#daytona-toolbox/tag/computer-use/POST/computeruse/stop)

### Get VNC status

Check the status of VNC processes to verify they are running.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
response = sandbox.computer_use.get_status()
print("VNC status:", response.status)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const status = await sandbox.computerUse.getStatus();
console.log('VNC status:', status.status);
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">

```ruby
response = sandbox.computer_use.status
puts "VNC status: #{response.status}"
```

</TabItem>
<TabItem label="Go" icon="seti:go">

```go
status, err := sandbox.ComputerUse.GetStatus(ctx)
if err != nil {
	log.Fatal(err)
}

fmt.Printf("VNC status: %v\n", status["status"])
```

</TabItem>
<TabItem label="API" icon="seti:json">

```bash
curl 'https://proxy.app.daytona.io/toolbox/{sandboxId}/computeruse/status'
```

</TabItem>
</Tabs>

For more information, see the [Computer Use](/docs/en/computer-use#get-status) reference.

> [**get_status (Python SDK)**](/docs/en/python-sdk/sync/computer-use#computeruseget_status)
>
> [**getStatus (TypeScript SDK)**](/docs/en/typescript-sdk/computer-use#getstatus)
>
> [**status (Ruby SDK)**](/docs/en/ruby-sdk/computer-use#status)
>
> [**GetStatus (Go SDK)**](/docs/en/go-sdk/daytona#ComputerUseService.GetStatus)
>
> [**Get Computer Use Status (API)**](/docs/en/tools/api/#daytona-toolbox/tag/computer-use/GET/computeruse/status)

For additional process management operations including restarting individual processes and viewing logs, see the [Computer Use](/docs/en/computer-use) reference.

## Automating desktop interactions

Once VNC is running, you can automate desktop interactions using Computer Use. This enables AI agents to programmatically control the mouse, keyboard, and capture screenshots within the VNC session.

**Available operations:**

- **Mouse**: click, move, drag, scroll, and get cursor position
- **Keyboard**: type text, press keys, and execute hotkey combinations
- **Screenshot**: capture full screen, regions, or compressed images
- **Display**: get display information and list open windows

For complete documentation on automating desktop interactions, see [Computer Use](/docs/en/computer-use).

> **Example**: Automated browser interaction

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
# Start VNC processes
sandbox.computer_use.start()

# Click to open browser
sandbox.computer_use.mouse.click(50, 50)

# Type a URL
sandbox.computer_use.keyboard.type("https://www.daytona.io/docs/")
sandbox.computer_use.keyboard.press("Return")

# Take a screenshot
screenshot = sandbox.computer_use.screenshot.take_full_screen()
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
// Start VNC processes
await sandbox.computerUse.start();

// Click to open browser
await sandbox.computerUse.mouse.click(50, 50);

// Type a URL
await sandbox.computerUse.keyboard.type('https://www.daytona.io/docs/');
await sandbox.computerUse.keyboard.press('Return');

// Take a screenshot
const screenshot = await sandbox.computerUse.screenshot.takeFullScreen();
```

</TabItem>
</Tabs>

## Required packages

The default sandbox image includes all packages required for VNC and Computer Use. If you are using a custom image, you need to install the following packages.

### VNC and desktop environment

| Package              | Description                                |
| -------------------- | ------------------------------------------ |
| **`xvfb`**           | X Virtual Framebuffer for headless display |
| **`xfce4`**          | Desktop environment                        |
| **`xfce4-terminal`** | Terminal emulator                          |
| **`x11vnc`**         | VNC server                                 |
| **`novnc`**          | Web-based VNC client                       |
| **`dbus-x11`**       | D-Bus session support                      |

### X11 libraries

| Library           | Description                                 |
| ----------------- | ------------------------------------------- |
| **`libx11-6`**    | X11 client library                          |
| **`libxrandr2`**  | X11 RandR extension (display configuration) |
| **`libxext6`**    | X11 extensions library                      |
| **`libxrender1`** | X11 rendering extension                     |
| **`libxfixes3`**  | X11 fixes extension                         |
| **`libxss1`**     | X11 screen saver extension                  |
| **`libxtst6`**    | X11 testing extension (input simulation)    |
| **`libxi6`**      | X11 input extension                         |
