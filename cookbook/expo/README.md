# Daytona + Expo Development Examples

This repository contains comprehensive examples demonstrating how to use **Daytona** to build and serve **Expo applications** with full support for web preview, iOS/Android Expo Go, and proper networking configuration.

## What This Solves

These examples address common challenges when running Expo apps in sandbox environments:

- CORS Issues: Properly configured networking and headers
- Multi-Platform Support: Web, iOS, and Android from a single codebase
- External Access: Public URLs that work with Expo Go on mobile devices
- Development Experience: Live reloading, debugging, and monitoring
- Resource Optimization: Properly sized containers for Expo development

## Example Files

### 1. **Complete Example** (`expo-daytona-example.py`)
- **Language**: Python
- **Features**: Full-featured implementation with monitoring, error handling, and enhanced UI
- **Best for**: Production-ready setups, learning all Daytona features

### 2. **TypeScript Version** (`expo-daytona-example.ts`)
- **Language**: TypeScript
- **Features**: Same functionality as Python version, demonstrating TypeScript SDK
- **Best for**: TypeScript/JavaScript developers

### 3. **Quick Start** (`expo-quickstart.py`)
- **Language**: Python
- **Features**: Minimal example showing core concepts
- **Best for**: Quick evaluation, understanding the essentials

## Quick Start

### Prerequisites

1. **Daytona Account**: Sign up at [app.daytona.io](https://app.daytona.io)
2. **API Key**: Generate one from the [Dashboard](https://app.daytona.io/dashboard/keys)
3. **Python/Node.js**: Install the Daytona SDK

### Installation

```bash
# Python
pip install daytona

# TypeScript/JavaScript
npm install @daytonaio/sdk
```

### Running the Examples

1. **Replace API Key**: Update `your-api-key` in the examples with your actual key
2. **Run the script**:

```bash
# Python examples
python expo-quickstart.py
# or
python expo-daytona-example.py

# TypeScript example
npx tsx expo-daytona-example.ts
```

## Key Components Explained

### 1. **Sandbox Creation**
```python
# Custom image with Node.js and Expo CLI pre-installed
expo_image = (
    Image.debian_slim("3.12")
    .run_commands(
        "curl -fsSL https://deb.nodesource.com/setup_18.x | bash -",
        "apt-get install -y nodejs",
        "npm install -g @expo/cli"
    )
    .env({"EXPO_NO_TELEMETRY": "1"})
)

# Sandbox with adequate resources for Expo development
sandbox = daytona.create(
    CreateSandboxFromImageParams(
        image=expo_image,
        resources=Resources(cpu=2, memory=4, disk=8),
        public=True,  # CRITICAL: Enables external access for Expo Go
        auto_stop_interval=0  # Keep running during development
    )
)
```

### 2. **Dual Server Setup**
```python
# Web development server (port 19006)
sandbox.process.execute_session_command(web_session, {
    "command": "npx expo start --web --port 19006 --host 0.0.0.0",
    "async": True
})

# Mobile development server (port 8081) 
sandbox.process.execute_session_command(mobile_session, {
    "command": "npx expo start --port 8081 --host 0.0.0.0 --tunnel",
    "async": True
})
```

### 3. **Preview URLs**
```python
# Get accessible URLs for all platforms
web_preview = sandbox.get_preview_link(19006)
mobile_preview = sandbox.get_preview_link(8081) 
terminal_preview = sandbox.get_preview_link(22222)
```

## Platform Access

### Web Browser
- Direct access via the web preview URL
- Full React Native Web functionality
- Hot reloading enabled

### Mobile (iOS/Android)
1. Install **Expo Go** from App Store/Play Store
2. Open the mobile server URL in your browser
3. Scan the QR code with Expo Go
4. App loads directly on your device

### Terminal Access
- Full shell access for debugging
- Install additional packages
- Monitor logs and processes

## Networking & CORS Solutions

### Public Sandbox
```python
sandbox = daytona.create(
    CreateSandboxFromImageParams(
        public=True  # Essential for Expo Go connectivity
    )
)
```

### Proper Host Binding
```bash
# Always use 0.0.0.0 to allow external connections
npx expo start --host 0.0.0.0
```

### Tunnel Support
```bash
# --tunnel flag creates external tunnel for mobile access
npx expo start --tunnel
```

## Advanced Features

### File System Operations
```python
# Upload custom components
sandbox.fs.upload_file(app_content.encode(), f"{project_dir}/App.tsx")

# Create configuration files
sandbox.fs.upload_file(app_json.encode(), f"{project_dir}/app.json")
```

### Process Management
```python
# Create persistent sessions for long-running servers
sandbox.process.create_session("expo-web")
sandbox.process.create_session("expo-mobile")

# Monitor server status
sessions = sandbox.process.list_sessions()
```

### Resource Monitoring
```python
# Check server health
web_session = sandbox.process.get_session("expo-web")
if web_session.commands:
    latest_command = web_session.commands[-1]
    if latest_command.exit_code != 0:
        print("Server issue detected!")
```

## Resource Requirements

### Recommended Specifications
- **CPU**: 2 cores (for Metro bundler performance)
- **Memory**: 4GB (handles large node_modules)
- **Disk**: 8GB (project files + dependencies)

### Cost Optimization
```python
# Auto-stop when not in use
auto_stop_interval=60  # minutes

# Use smaller resources for simple apps
resources=Resources(cpu=1, memory=2, disk=4)
```

## Troubleshooting

### Common Issues

#### Expo Go Can't Connect
```python
# Ensure sandbox is public
public=True

# Check tunnel is working
command="npx expo start --tunnel"
```

#### Server Won't Start
```python
# Increase wait time
time.sleep(15)  

# Check logs
session = sandbox.process.get_session("expo-mobile")
print(session.commands[-1].result)
```

#### CORS Errors
```python
# Web configuration in app.json
"web": {
  "bundler": "metro"
}

# Proper host binding
--host 0.0.0.0
```

### Debug Commands
```python
# Terminal access for debugging
terminal_url = sandbox.get_preview_link(22222).url

# Check running processes
sandbox.process.exec("ps aux | grep expo")

# View network configuration
sandbox.process.exec("netstat -tulpn")
```

## Development Workflow

1. **Initial Setup**: Run example script to create environment
2. **Development**: Use terminal URL for file editing and debugging
3. **Testing**: 
   - Web: Use browser URL
   - Mobile: Use Expo Go with QR code
4. **Iteration**: Modify code and see live updates
5. **Deployment**: Build and deploy using standard Expo workflows

## Production Considerations

### Environment Variables
```python
.env({
    "EXPO_NO_TELEMETRY": "1",
    "NODE_ENV": "development",
    "EXPO_NO_DOTENV": "1"
})
```

### Security
```python
# Use environment variables for sensitive data
sandbox.process.exec("export API_KEY=$DAYTONA_API_KEY")
```

### Scaling
```python
# Create multiple sandboxes for team development
for developer in developers:
    sandbox = create_expo_sandbox(f"expo-{developer}")
```

## Additional Resources

- [Daytona Documentation](https://docs.daytona.io)
- [Expo Documentation](https://docs.expo.dev)
- [Daytona Python SDK](https://github.com/daytonaio/sdk)
- [Daytona TypeScript SDK](https://github.com/daytonaio/sdk)

## Support

For issues with:
- **Daytona Platform**: [support@daytona.io](mailto:support@daytona.io)
- **This Example**: Create an issue in this repository
- **Expo Questions**: [Expo Community](https://expo.dev/community)

## Key Benefits

✅ **Zero Setup Time**: Environment ready in minutes
✅ **Cross-Platform**: Web, iOS, Android from single codebase  
✅ **External Access**: Real device testing with Expo Go
✅ **Collaborative**: Share preview URLs with team members
✅ **Scalable**: Create multiple environments as needed
✅ **Cost-Effective**: Pay only for active development time 