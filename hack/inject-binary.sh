# Get machine architecture
ARCH=$(uname -m)
case $ARCH in
"arm64" | "ARM64")
  ARCH_VALUE="arm64"
  ;;
"x86_64" | "AMD64")
  ARCH_VALUE="amd64"
  ;;
"aarch64")
  ARCH_VALUE="arm64"
  ;;
*)
  err "Unsupported architecture: $ARCH"
  ;;
esac

# Get machine operating system
OS=$(uname -s)
case $OS in
"Linux")
  OUTPUT_FOLDER="$HOME/.config/daytona/server/binaries/v0.0.0-dev"
  ;;
"Darwin")
  OUTPUT_FOLDER="$HOME/Library/Application Support/daytona/server/binaries/v0.0.0-dev"
  ;;
*)
  echo "Unsupported operating system: $OS"
  exit 1
  ;;
esac

# Create output folder if it doesn't exist
mkdir -p "$OUTPUT_FOLDER" || { echo "Failed to create output folder"; exit 1; }

# Build the project container binary
GOOS=linux GOARCH=$ARCH_VALUE go build -o $OUTPUT_FOLDER/daytona-linux-$ARCH_VALUE cmd/daytona/main.go || { echo "Build failed"; exit 1; }
echo "Binary build successful"