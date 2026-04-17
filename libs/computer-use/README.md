# ComputerUse - Process Management for VNC Desktop Environment

This package provides a Computer Use plugin used by the Daytona Daemon to allow agents to control VNC desktop environments.

## Overview

The `ComputerUse` package manages five main processes in the correct order:

1. **Xvfb** (X Virtual Framebuffer) - Provides a virtual display
2. **xfce4** (Desktop Environment) - Starts the XFCE desktop environment
3. **at-spi-bus-launcher** (Accessibility bus) - Launches `org.a11y.Bus` so GTK/Qt/Electron apps publish their widget trees for the accessibility API
4. **x11vnc** (VNC Server) - Exposes the desktop via VNC protocol
5. **novnc** (Web-based VNC client) - Provides web access to the VNC server

## Features

- **Process Management**: Automatic startup, monitoring, and shutdown of processes
- **Priority-based Startup**: Processes start in the correct order based on dependencies
- **Auto-restart**: Failed processes are automatically restarted
- **Logging**: Individual log files for each process
- **Status Monitoring**: Check the status of all processes
- **Graceful Shutdown**: Proper cleanup when stopping processes
- **Individual Control**: Start, stop, or restart individual processes

## Dockerfile Requirements

To use the ComputerUse package, your Dockerfile must include the following VNC-related packages and setup:

### Required Packages

```dockerfile
# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive
ENV DISPLAY=:1
ENV VNC_PORT=5901
ENV NO_VNC_PORT=6080
ENV VNC_RESOLUTION=1280x720

# Install VNC and desktop environment packages
RUN apt-get update && apt-get install -y \
    wget \
    git \
    vim \
    xfce4 \
    xfce4-terminal \
    dbus-x11 \
    xfonts-base \
    xfonts-100dpi \
    xfonts-75dpi \
    xfonts-scalable \
    x11vnc \
    novnc \
    supervisor \
    net-tools \
    locales \
    xvfb \
    x11-utils \
    x11-xserver-utils \
    gnome-screenshot \
    scrot \
    imagemagick \
    xdotool \
    xautomation \
    wmctrl \
    build-essential \
    libx11-dev \
    libxext-dev \
    libxtst-dev \
    libxinerama-dev \
    libx11-xcb-dev \
    libxkbcommon-dev \
    libxkbcommon-x11-dev \
    libxcb-xkb-dev \
    libpng-dev \
    chromium \
    # AT-SPI (accessibility API)
    at-spi2-core \
    libatspi2.0-0 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
```

### VNC Setup

```dockerfile
# Setup VNC
RUN mkdir -p /home/daytona/.vnc && \
    chown -R daytona:daytona /home/daytona/.vnc

# NoVNC setup
RUN ln -sf /usr/share/novnc/vnc.html /usr/share/novnc/index.html && \
    sed -i 's/websockify =/websockify = --heartbeat 30/' /usr/share/novnc/utils/launch.sh
```

### Launch Script

The NoVNC launch script (`/usr/share/novnc/utils/launch.sh`) is used to start the web-based VNC client. The Dockerfile modifies this script to add heartbeat support:

```bash
# Add heartbeat to websockify for better connection stability
sed -i 's/websockify =/websockify = --heartbeat 30/' /usr/share/novnc/utils/launch.sh
```

This ensures that the WebSocket connection remains stable during long-running sessions.

### Additional Tools

The Dockerfile also installs several useful tools for VNC desktop interaction:

- **xdotool**: For mouse and keyboard automation
- **xautomation**: Additional automation tools
- **wmctrl**: Window manager control
- **scrot**: Screenshot capture
- **imagemagick**: Image processing
- **gnome-screenshot**: Alternative screenshot tool
- **chromium**: Web browser for testing

These tools are used by the toolbox API endpoints for desktop interaction.

### Environment Variables

The following environment variables should be set in your Dockerfile:

```dockerfile
ENV DEBIAN_FRONTEND=noninteractive
ENV DISPLAY=:1
ENV VNC_PORT=5901
ENV NO_VNC_PORT=6080
ENV VNC_RESOLUTION=1280x720
# AT-SPI accessibility API — ensures GTK/Qt apps publish their widget trees
ENV GTK_MODULES=gail:atk-bridge
ENV QT_ACCESSIBILITY=1
ENV NO_AT_BRIDGE=0
```

App-side accessibility caveats:

- **GTK apps** (xfce4, xfce4-terminal, Thunar, most of GNOME) publish their tree automatically once `at-spi2-core` is installed and `GTK_MODULES=gail:atk-bridge` is set. No per-app work.
- **Qt apps** need `QT_ACCESSIBILITY=1` in their environment.
- **Chromium and Electron** need `--force-renderer-accessibility` on the command line. Without it they expose only a minimal AT-SPI subtree. This plugin does not enforce the flag — launch Chromium with it when you want agents to read browser DOM via AT-SPI, or use CDP instead.

### User Setup

Ensure you have a non-root user with proper permissions:

```dockerfile
# Create the Daytona user and configure sudo access
RUN useradd -m daytona && echo "daytona ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/91-daytona

# Switch to the user for VNC operations
USER daytona
```

## Configuration

### Environment Variables

- `VNC_RESOLUTION`: Set the VNC resolution (default: "1920x1080")
- `VNC_PORT`: VNC server port (default: 5901)
- `NO_VNC_PORT`: NoVNC web port (default: 6080)
- `DISPLAY`: X display (default: ":1")
- `VNC_USER`: User to run VNC processes (default: "daytona")

### Process Configuration

The processes are configured with the following settings based on environment variables:

| Process | Command                                                                            | Priority | Auto-restart | Log Files | Environment                                           |
| ------- | ---------------------------------------------------------------------------------- | -------- | ------------ | --------- | ----------------------------------------------------- |
| xvfb    | `/usr/bin/Xvfb $DISPLAY -screen 0 $VNC_RESOLUTIONx24`                              | 100      | Yes          | No        | `DISPLAY`                                             |
| xfce4   | `/usr/bin/startxfce4`                                                              | 200      | Yes          | Yes       | `DISPLAY`, `HOME`, `USER`, `DBUS_SESSION_BUS_ADDRESS` |
| atspi   | `/usr/libexec/at-spi-bus-launcher --launch-immediately`                            | 250      | Yes          | Yes       | `DISPLAY`, `HOME`, `USER`, `DBUS_SESSION_BUS_ADDRESS` |
| x11vnc  | `/usr/bin/x11vnc -display $DISPLAY -forever -shared -rfbport $VNC_PORT`            | 300      | Yes          | No        | `DISPLAY`                                             |
| novnc   | `/usr/share/novnc/utils/launch.sh --vnc localhost:$VNC_PORT --listen $NO_VNC_PORT` | 400      | Yes          | No        | `DISPLAY`                                             |

**Default Values:**

- `DISPLAY`: `:1`
- `VNC_RESOLUTION`: `1920x1080`
- `VNC_PORT`: `5901`
- `NO_VNC_PORT`: `6080`
- `VNC_USER`: `daytona`

### Log Files

Log files are stored in `~/.daytona/computeruse/`:

- `xfce4.log` - Standard output from xfce4
- `xfce4.err` - Error output from xfce4

## Integration with Toolbox

The `ComputerUse` package is integrated into the toolbox server and provides HTTP endpoints for:

- Screenshot functionality
- Mouse control
- Keyboard control
- Display information
- **Accessibility (AT-SPI) — read the widget tree, query nodes by role/name/state, and act on them semantically**

These endpoints are available under the `/computeruse` route group in the toolbox API.

## Accessibility API (AT-SPI)

Agents can read the Linux desktop accessibility tree and act on nodes (focus, invoke, set value) without relying on screenshots or pixel coordinates. Useful for text-only agents and as a complement to vision for multimodal ones.

### Endpoints

All under `/computeruse/a11y/`:

| Method | Path               | Description                                                    |
| ------ | ------------------ | -------------------------------------------------------------- |
| GET    | `/tree`            | Walk the accessibility tree under a scope (focused/pid/all).   |
| POST   | `/find`            | Return a flat list of nodes matching role/name/state filters.  |
| POST   | `/node/focus`      | Move keyboard focus to a node (via `Component.GrabFocus`).     |
| POST   | `/node/invoke`     | Fire a node's default action (button press, menu activate...). |
| POST   | `/node/value`      | Set a node's value (text entry, slider, combo box).            |

### Node shape

```json
{
  "id":          "<bus-name>:<object-path>",
  "role":        "push button",
  "name":        "Submit",
  "description": "",
  "bounds":      { "x": 120, "y": 340, "width": 96, "height": 32 },
  "states":      ["enabled","visible","showing","sensitive","focusable"],
  "actions":     ["press","click","release"],
  "children":    [ ... ]
}
```

`children` is present in tree responses and omitted in find results. Node IDs are AT-SPI accessible paths (D-Bus unique name + object path) — stable within a session, directly usable in subsequent action calls.

### Filters (`/find`)

- `scope`: `"focused"` (default) | `"pid"` | `"all"`. When `pid`, pass `pid` as an integer.
- `role`: exact, case-insensitive match against the AT-SPI role name.
- `name` + `nameMatch`: `"exact"`, `"substring"` (default), or `"regex"`. Substring and regex are case-sensitive.
- `states[]`: node must have all listed states.
- `limit`: max matches returned (default 500, ceiling 5000). Response includes `"truncated": true` when the cap is hit.
- All filter fields are AND-ed.

### Error model

| Status | Condition                                                            |
| ------ | -------------------------------------------------------------------- |
| 200    | Success.                                                             |
| 400    | Bad request — invalid scope, unsupported action on the node, etc.    |
| 404    | Node ID refers to a widget/window that has gone away.                |
| 503    | `org.a11y.Bus` unreachable. `"code": "A11Y_UNAVAILABLE"` in body.    |
| 500    | Unclassified internal error.                                         |

On 503, the rest of the computeruse API (mouse/keyboard/screenshot) stays up — agents can fall back to vision or pixel input.

### App-side caveats

AT-SPI only reflects apps that publish their widget tree on the a11y bus:

- **GTK apps** (xfce4, xfce4-terminal, Thunar, most GNOME apps) — publish automatically once `at-spi2-core` is installed and `GTK_MODULES=gail:atk-bridge` is set.
- **Qt apps** — need `QT_ACCESSIBILITY=1` in their environment.
- **Chromium and Electron** — need `--force-renderer-accessibility` on the command line. Without it they expose only a minimal subtree. This plugin does not enforce the flag; launch Chromium with it when you want AT-SPI coverage, or use CDP for browser automation instead.

### What's not in v1

- No event subscriptions, streaming, or long-poll "wait until match" endpoint. Agents poll.
- No dedicated Table-interface endpoint (row/column coordinate lookup). Individual table cells that appear in the tree are reachable via `/find` + `/node/*` like any other node.
- No multi-select Selection-interface endpoints. Use `focus` + shift-click via the mouse endpoints.
- No drag-and-drop via AT-SPI — use the mouse drag endpoint.

## Error Handling

The implementation includes comprehensive error handling:

- Process startup failures are logged and retried (if auto-restart is enabled)
- Log file access errors are handled gracefully
- Process termination is handled properly with context cancellation
- Mutex-based thread safety for concurrent access

## Thread Safety

All operations are thread-safe using appropriate mutexes:

- `sync.RWMutex` for the main `ComputerUse` struct
- `sync.Mutex` for individual `Process` structs

This allows safe concurrent access from multiple goroutines.
