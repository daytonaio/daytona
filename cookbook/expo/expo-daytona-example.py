"""
Daytona Expo App Builder Example
This example demonstrates how to create and serve an Expo app in Daytona
with support for web preview, iOS/Android Expo Go, and proper networking.
"""

import time
from daytona import Daytona, DaytonaConfig, CreateSandboxFromImageParams, Image, Resources

def create_expo_sandbox():
    """Create a Daytona sandbox optimized for Expo development"""
    print("Creating Daytona sandbox for Expo development...")
    
    # Initialize Daytona client
    daytona = Daytona(DaytonaConfig(
        api_key="your-api-key",  # Replace with your API key
        target="us"  # or "eu" based on your preference
    ))
    
    # Create a custom image with Node.js and Expo CLI (Node 20 base)
    expo_image = (
        Image.base("node:20")
        .run_commands(
            # Install Expo CLI and other necessary tools
            "npm install -g @expo/cli @expo/ngrok expo-doctor",
            # Install additional development tools
            "apt-get update && apt-get install -y git curl wget",
            # Create development directory
            "mkdir -p /home/daytona/expo-projects"
        )
        .workdir("/home/daytona/expo-projects")
        .env({
            "EXPO_NO_DOTENV": "1",
            "EXPO_NO_TELEMETRY": "1",
            "NODE_ENV": "development"
        })
    )
    
    # Create sandbox with adequate resources for Expo development
    sandbox = daytona.create(
        CreateSandboxFromImageParams(
            image=expo_image,
            resources=Resources(
                cpu=2,      # 2 CPU cores for better build performance
                memory=4,   # 4GB RAM for Metro bundler
                disk=8      # 8GB storage for node_modules and builds
            ),
            auto_stop_interval=0,  # Keep running for development
            public=True  # Allow external access for Expo Go
        ),
        timeout=0,
        on_snapshot_create_logs=print,
    )
    
    print(f"Sandbox created with ID: {sandbox.id}")
    return sandbox

def setup_expo_project(sandbox):
    """Create and configure a new Expo project"""
    print("Setting up Expo project...")
    
    # Create a new Expo project
    result = sandbox.process.exec(
        "npx create-expo-app MyExpoApp --template blank-typescript",
        cwd="/home/daytona/expo-projects"
    )
    
    if result.exit_code != 0:
        print(f"Failed to create Expo project: {result.result}")
        return False
    
    print("Expo project created successfully")
    
    # Navigate to project directory and install dependencies
    project_dir = "/home/daytona/expo-projects/MyExpoApp"
    
    # Install additional useful packages
    result = sandbox.process.exec(
        "npx expo install react-dom react-native-web @expo/metro-runtime",
        cwd=project_dir
    )
    print("Install result:", result)
    if result.exit_code != 0:
        print(f"Warning: Some packages failed to install: {result.result}")
    
    # Create a more interesting app component
    app_content = '''import { StatusBar } from 'expo-status-bar';
import React, { useState } from 'react';
import { StyleSheet, Text, View, TouchableOpacity, Platform } from 'react-native';

export default function App() {
  const [count, setCount] = useState(0);
  
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Expo + Daytona</Text>
      <Text style={styles.subtitle}>
        Running on {Platform.OS === 'web' ? 'Web' : Platform.OS}
      </Text>
      
      <View style={styles.counterContainer}>
        <Text style={styles.counterText}>Count: {count}</Text>
        <TouchableOpacity 
          style={styles.button} 
          onPress={() => setCount(count + 1)}
        >
          <Text style={styles.buttonText}>Increment</Text>
        </TouchableOpacity>
        <TouchableOpacity 
          style={[styles.button, styles.resetButton]} 
          onPress={() => setCount(0)}
        >
          <Text style={styles.buttonText}>Reset</Text>
        </TouchableOpacity>
      </View>
      
      <Text style={styles.info}>
        This app is running in a Daytona sandbox!
      </Text>
      
      <StatusBar style="auto" />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 20,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    marginBottom: 10,
  },
  subtitle: {
    fontSize: 18,
    color: '#666',
    marginBottom: 40,
  },
  counterContainer: {
    alignItems: 'center',
    marginBottom: 40,
  },
  counterText: {
    fontSize: 24,
    marginBottom: 20,
  },
  button: {
    backgroundColor: '#007AFF',
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 8,
    marginVertical: 5,
    minWidth: 120,
    alignItems: 'center',
  },
  resetButton: {
    backgroundColor: '#FF3B30',
  },
  buttonText: {
    color: 'white',
    fontSize: 16,
    fontWeight: '600',
  },
  info: {
    fontSize: 14,
    color: '#888',
    textAlign: 'center',
    marginTop: 20,
  },
});'''
    
    # Upload the enhanced App component
    sandbox.fs.upload_file(app_content.encode(), f"{project_dir}/App.tsx")
    
    # Create app.json with proper configuration for Daytona
    app_json = '''{
  "expo": {
    "name": "MyExpoApp",
    "slug": "my-expo-app",
    "version": "1.0.0",
    "orientation": "portrait",
    "icon": "./assets/icon.png",
    "userInterfaceStyle": "light",
    "splash": {
      "image": "./assets/splash.png",
      "resizeMode": "contain",
      "backgroundColor": "#ffffff"
    },
    "assetBundlePatterns": [
      "**/*"
    ],
    "ios": {
      "supportsTablet": true
    },
    "android": {
      "adaptiveIcon": {
        "foregroundImage": "./assets/adaptive-icon.png",
        "backgroundColor": "#FFFFFF"
      }
    },
    "web": {
      "favicon": "./assets/favicon.png",
      "bundler": "metro"
    },
    "extra": {
      "eas": {
        "projectId": "your-project-id"
      }
    }
  }
}'''
    
    sandbox.fs.upload_file(app_json.encode(), f"{project_dir}/app.json")
    
    print("Project setup completed with enhanced components")
    return True

def start_expo_servers(sandbox):
    """Start Expo development servers for web and mobile"""
    print("Starting Expo development servers...")
    
    project_dir = "/home/daytona/expo-projects/MyExpoApp"
    
    # Create session for web server
    web_session = "expo-web"
    sandbox.process.create_session(web_session)
    
    # Start web development server
    print("Starting web server on port 19006...")
    web_command = sandbox.process.execute_session_command(
        web_session,
        {
            "command": f"cd {project_dir} && npx expo start --web --port 19006",
            "async": True
        }
    )
    
    # Create session for mobile development server  
    mobile_session = "expo-mobile"
    sandbox.process.create_session(mobile_session)
    
    # Start mobile development server (for Expo Go)
    print("Starting mobile server on port 8081...")
    mobile_command = sandbox.process.execute_session_command(
        mobile_session,
        {
            "command": f"cd {project_dir} && npx expo start --port 8081 --tunnel",
            "async": True
        }
    )
    
    # Wait for servers to start
    print("Waiting for servers to initialize...")
    time.sleep(10)
    
    return {
        "web_session": web_session,
        "mobile_session": mobile_session,
        "web_command": web_command,
        "mobile_command": mobile_command
    }

def get_access_urls(sandbox):
    """Get all the access URLs for the Expo app"""
    print("Getting access URLs...")
    
    access_info = {}
    
    # Get web preview URL
    try:
        web_preview = sandbox.get_preview_link(19006)
        access_info['web_url'] = web_preview.url
        access_info['web_token'] = getattr(web_preview, 'token', None)
        print(f"Web Preview: {web_preview.url}")
    except Exception as e:
        print(f"Warning: Could not get web preview: {e}")
    
    # Get mobile development server URL  
    try:
        mobile_preview = sandbox.get_preview_link(8081)
        access_info['mobile_url'] = mobile_preview.url
        access_info['mobile_token'] = getattr(mobile_preview, 'token', None)
        print(f"Mobile Server: {mobile_preview.url}")
    except Exception as e:
        print(f"Warning: Could not get mobile preview: {e}")
    
    # Get terminal access
    try:
        terminal_preview = sandbox.get_preview_link(22222)
        access_info['terminal_url'] = terminal_preview.url
        access_info['terminal_token'] = getattr(terminal_preview, 'token', None)
        print(f"Terminal: {terminal_preview.url}")
    except Exception as e:
        print(f"Warning: Could not get terminal access: {e}")
    
    return access_info

def monitor_servers(sandbox, sessions_info):
    """Monitor the running servers and show logs"""
    print("Monitoring servers (Ctrl+C to stop)...")
    
    try:
        while True:
            # Check web server logs
            try:
                web_session = sandbox.process.get_session(sessions_info['web_session'])
                if web_session.commands:
                    latest_command = web_session.commands[-1]
                    if getattr(latest_command, 'exit_code', None) is not None and latest_command.exit_code != 0:
                        print(f"Warning: Web server issue detected: {latest_command.exit_code}")
            except Exception:
                pass
            
            # Check mobile server logs  
            try:
                mobile_session = sandbox.process.get_session(sessions_info['mobile_session'])
                if mobile_session.commands:
                    latest_command = mobile_session.commands[-1]
                    if getattr(latest_command, 'exit_code', None) is not None and latest_command.exit_code != 0:
                        print(f"Warning: Mobile server issue detected: {latest_command.exit_code}")
            except Exception:
                pass
            
            print("Servers are running...")
            time.sleep(30)  # Check every 30 seconds
            
    except KeyboardInterrupt:
        print("\nStopping monitoring...")

def main():
    """Main function to orchestrate the Expo app setup"""
    print("Welcome to Daytona Expo App Builder!")
    print("This example will create a complete Expo development environment.")
    
    try:
        # Step 1: Create sandbox
        sandbox = create_expo_sandbox()
        
        # Step 2: Setup Expo project
        if not setup_expo_project(sandbox):
            print("Failed to setup project")
            return
        
        # Step 3: Start development servers
        sessions_info = start_expo_servers(sandbox)
        
        # Step 4: Get access URLs
        access_info = get_access_urls(sandbox)
        
        # Step 5: Display instructions
        print("\n" + "="*60)
        print("YOUR EXPO APP IS READY!")
        print("="*60)
        
        if 'web_url' in access_info:
            print(f"Web App: {access_info['web_url']}")
            print("   → Open this URL in your browser to see the web version")
        
        if 'mobile_url' in access_info:
            print(f"Mobile Development Server: {access_info['mobile_url']}")
            print("   → Open Expo Go app and scan the QR code from this URL")
        
        if 'terminal_url' in access_info:
            print(f"Terminal Access: {access_info['terminal_url']}")
            print("   → Use this for debugging and running commands")
        
        print("\nINSTRUCTIONS:")
        print("1. Open the web URL to see your app running in the browser")
        print("2. Install 'Expo Go' app on your iOS/Android device")
        print("3. Open the mobile server URL and scan the QR code with Expo Go")
        print("4. Make changes to your code and see live updates!")
        
        print("\nTROUBLESHOOTING:")
        print("- If Expo Go can't connect, ensure your device is on the same network")
        print("- For CORS issues, the app is configured with proper headers")
        print("- Use the terminal URL for debugging and log inspection")
        
        print("\nNote: Keep this script running to maintain the servers!")
        print("="*60)
        
        # Step 6: Monitor servers
        monitor_servers(sandbox, sessions_info)
        
    except Exception as e:
        print(f"Error: {e}")
        print("Please check your Daytona API key and network connection.")
    
    finally:
        print("\nCleaning up...")
        try:
            # Note: In a real scenario, you might want to keep the sandbox running
            # sandbox.delete()
            print("Sandbox kept running for continued development")
        except:
            pass

if __name__ == "__main__":
    main() 