# Screen Recording Dependencies for Linux Sandbox

This document lists the dependencies required for the screen recording feature to work in the Linux-based sandbox.

## Required Packages

### FFmpeg (Required)

FFmpeg is the core dependency for screen recording. It must be installed with x11grab support.

**Ubuntu/Debian:**

```bash
apt-get update && apt-get install -y ffmpeg
```

**Alpine:**

```bash
apk add ffmpeg
```

**Fedora/RHEL:**

```bash
dnf install -y ffmpeg
```

### X11 Libraries (Required for x11grab)

FFmpeg's x11grab input device requires X11 libraries:

**Ubuntu/Debian:**

```bash
apt-get install -y libx11-6 libxext6 libxfixes3
```

**Alpine:**

```bash
apk add libx11 libxext libxfixes
```

### Verification

To verify FFmpeg is installed with x11grab support:

```bash
ffmpeg -devices 2>&1 | grep x11grab
```

Expected output should include:

```
 D  x11grab         X11 screen capture, using XCB
```

## Optional Packages

### libxcb (Recommended)

Modern FFmpeg uses XCB (X11 C Bindings) for x11grab which is more efficient:

**Ubuntu/Debian:**

```bash
apt-get install -y libxcb1 libxcb-shm0 libxcb-shape0 libxcb-xfixes0
```

### H.264 Encoder (Required for MP4 output)

The recording feature uses libx264 for H.264 encoding:

**Ubuntu/Debian:**

```bash
apt-get install -y libx264-dev
```

Note: Most FFmpeg packages include libx264 by default.

## Complete Installation Command

### Ubuntu/Debian (Recommended)

```bash
apt-get update && apt-get install -y \
    ffmpeg \
    libx11-6 \
    libxext6 \
    libxfixes3 \
    libxcb1 \
    libxcb-shm0 \
    libxcb-shape0 \
    libxcb-xfixes0
```

### Alpine Linux

```bash
apk add --no-cache \
    ffmpeg \
    libx11 \
    libxext \
    libxfixes \
    libxcb
```

## Environment Requirements

### DISPLAY Variable

The screen recording feature requires a valid X11 display. The `DISPLAY` environment variable must be set:

```bash
export DISPLAY=:0
```

If running in a headless environment with Xvfb:

```bash
# Start Xvfb
Xvfb :99 -screen 0 1920x1080x24 &
export DISPLAY=:99
```

### X Server Access

The daemon process must have permission to connect to the X server. If running as a different user:

```bash
xhost +local:
```

Or more securely, grant access to a specific user:

```bash
xhost +SI:localuser:daytona
```

## Dockerfile Example

```dockerfile
FROM ubuntu:22.04

# Install recording dependencies
RUN apt-get update && apt-get install -y \
    ffmpeg \
    libx11-6 \
    libxext6 \
    libxfixes3 \
    libxcb1 \
    libxcb-shm0 \
    libxcb-shape0 \
    libxcb-xfixes0 \
    && rm -rf /var/lib/apt/lists/*

# Set default display
ENV DISPLAY=:0
```

## Storage Location

Recordings are stored in:

```
~/.daytona/recordings/
```

Ensure this directory is writable by the daemon process and has sufficient disk space for video files.

## FFmpeg Command Used

The recording feature uses the following FFmpeg command:

```bash
ffmpeg -f x11grab -framerate 30 -i :0 -c:v libx264 -preset ultrafast -pix_fmt yuv420p -y output.mp4
```

Parameters:

- `-f x11grab`: X11 screen capture input
- `-framerate 30`: 30 frames per second
- `-i :0`: Capture from display :0
- `-c:v libx264`: H.264 video codec
- `-preset ultrafast`: Fast encoding for real-time capture
- `-pix_fmt yuv420p`: Standard pixel format for compatibility
- `-y`: Overwrite output file if exists

## Troubleshooting

### "ffmpeg not found"

FFmpeg is not installed or not in PATH:

```bash
which ffmpeg
apt-get install -y ffmpeg
```

### "Cannot open display"

DISPLAY environment variable is not set or X server is not accessible:

```bash
echo $DISPLAY
export DISPLAY=:0
```

### "Permission denied" on X server

Grant X server access:

```bash
xhost +local:
```

### Recording starts but file is empty/corrupt

Check if FFmpeg has x11grab support:

```bash
ffmpeg -devices 2>&1 | grep x11grab
```

If not present, reinstall FFmpeg with proper dependencies.
