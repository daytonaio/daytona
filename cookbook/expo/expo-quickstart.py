"""
Daytona Expo App Builder Example (Python)
This example demonstrates how to create and serve an Expo app in Daytona
with support for web preview, iOS/Android Expo Go, and proper networking.
"""

from daytona import Daytona, Image, Resources, CreateSandboxFromImageParams
import time

APP_TSX = '''import { StatusBar } from 'expo-status-bar';
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

APP_JSON = '''{
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

def main():
    print("Daytona Expo App Builder (Python)")
    print("This example will create a complete Expo development environment.")

    # 1. Initialize Daytona
    daytona = Daytona(api_key="your-api-key")  # Replace with your API key

    # 2. Create optimized Expo environment (Node 20 base image)
    expo_image = (
        Image.base("node:20")
        .run_commands(
            "npm install -g @expo/cli @expo/ngrok expo-doctor",
            "apt-get update && apt-get install -y git curl wget",
            "mkdir -p /home/daytona/expo-projects"
        )
        .workdir("/home/daytona/expo-projects")
        .env({
            "EXPO_NO_DOTENV": "1",
            "EXPO_NO_TELEMETRY": "1",
            "NODE_ENV": "development"
        })
    )

    # 3. Create sandbox with adequate resources
    sandbox = daytona.create(
        CreateSandboxFromImageParams(
            image=expo_image,
            resources=Resources(cpu=2, memory=4, disk=8),
            public=True,
            auto_stop_interval=0
        )
    )
    print(f"Sandbox created: {sandbox.id}")

    # 4. Setup Expo project
    print("Setting up Expo project...")
    result = sandbox.process.exec(
        "npx create-expo-app MyExpoApp --template blank-typescript",
        "/home/daytona/expo-projects"
    )
    if result.exit_code != 0:
        print(f"Failed to create Expo project: {result.result}")
        return
    print("Expo project created successfully")
    project_dir = "/home/daytona/expo-projects/MyExpoApp"

    # Install additional useful packages
    install_result = sandbox.process.exec(
        "npx expo install react-dom react-native-web @expo/metro-runtime",
        project_dir
    )
    print("Install result:", install_result)
    if install_result.exit_code != 0:
        print(f"Warning: Some packages failed to install: {install_result.result}")

    # Upload enhanced App.tsx
    sandbox.fs.upload_file(APP_TSX.encode(), f"{project_dir}/App.tsx")
    # Upload app.json
    sandbox.fs.upload_file(APP_JSON.encode(), f"{project_dir}/app.json")
    print("Project setup completed with enhanced components")

    # 5. Start development servers
    print("Starting Expo development servers...")
    # Web server session
    web_session = "expo-web"
    sandbox.process.create_session(web_session)
    web_command = sandbox.process.execute_session_command(web_session, {
        "command": f"cd {project_dir} && npx expo start --web --port 19006",
        "async": True
    })
    # Mobile server session
    mobile_session = "expo-mobile"
    sandbox.process.create_session(mobile_session)
    mobile_command = sandbox.process.execute_session_command(mobile_session, {
        "command": f"cd {project_dir} && npx expo start --port 8081 --tunnel",
        "async": True
    })
    print("Waiting for servers to initialize...")
    time.sleep(10)

    # 6. Get access URLs
    access_info = {}
    try:
        web_preview = sandbox.get_preview_link(19006)
        access_info["webUrl"] = web_preview.url
        print(f"Web Preview: {web_preview.url}")
    except Exception as e:
        print(f"Warning: Could not get web preview: {e}")
    try:
        mobile_preview = sandbox.get_preview_link(8081)
        access_info["mobileUrl"] = mobile_preview.url
        print(f"Mobile Server: {mobile_preview.url}")
    except Exception as e:
        print(f"Warning: Could not get mobile preview: {e}")
    try:
        terminal_preview = sandbox.get_preview_link(22222)
        access_info["terminalUrl"] = terminal_preview.url
        print(f"Terminal: {terminal_preview.url}")
    except Exception as e:
        print(f"Warning: Could not get terminal access: {e}")

    # 7. Display instructions
    print("\n" + "="*60)
    print("YOUR EXPO APP IS READY!")
    print("="*60)
    if access_info.get("webUrl"):
        print(f"Web App: {access_info['webUrl']}")
        print("   → Open this URL in your browser to see the web version")
    if access_info.get("mobileUrl"):
        print(f"Mobile Development Server: {access_info['mobileUrl']}")
        print("   → Open Expo Go app and scan the QR code from this URL")
    if access_info.get("terminalUrl"):
        print(f"Terminal Access: {access_info['terminalUrl']}")
        print("   → Use this for debugging and running commands")
    print("\nINSTRUCTIONS:")
    print("1. Open the web URL to see your app running in the browser")
    print('2. Install "Expo Go" app on your iOS/Android device')
    print("3. Open the mobile server URL and scan the QR code with Expo Go")
    print("4. Make changes to your code and see live updates!")
    print("\nTROUBLESHOOTING:")
    print("- If Expo Go can't connect, ensure your device is on the same network")
    print("- For CORS issues, the app is configured with proper headers")
    print("- Use the terminal URL for debugging and log inspection")
    print("\nNote: Keep this script running to maintain the servers!")
    print("="*60)

if __name__ == "__main__":
    main() 