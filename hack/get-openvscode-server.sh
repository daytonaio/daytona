RELEASE_TAG="v1.86.2"
RELEASE_ORG="gitpod-io"
OPENVSCODE_SERVER_ROOT="$HOME/vscode-server"

if [ -d "$OPENVSCODE_SERVER_ROOT" ]; then
  echo "OpenVSCode Server is already installed. Skipping installation."
  exit 0
fi

# Downloading the latest VSC Server release and extracting the release archive
# Rename `openvscode-server` cli tool to `code` for convenience
if [ -z "$RELEASE_TAG" ]; then
    echo "The RELEASE_TAG build arg must be set." >&2 &&
    exit 1;
fi

arch=$(uname -m)

if [ "$arch" = "x86_64" ]; then
  arch="x64";
elif [ "$arch" = "aarch64" ]; then
  arch="arm64";
elif [ "$arch" = "armv7l" ]; then
  arch="armhf";
fi

wget https://github.com/$RELEASE_ORG/openvscode-server/releases/download/openvscode-server-$RELEASE_TAG/openvscode-server-$RELEASE_TAG-linux-$arch.tar.gz -O $HOME/openvscode-server-$RELEASE_TAG-linux-$arch.tar.gz
tar -xzf $HOME/openvscode-server-$RELEASE_TAG-linux-$arch.tar.gz -C $HOME
mv -f $HOME/openvscode-server-$RELEASE_TAG-linux-$arch $OPENVSCODE_SERVER_ROOT

cp $OPENVSCODE_SERVER_ROOT/bin/remote-cli/openvscode-server $OPENVSCODE_SERVER_ROOT/bin/remote-cli/code
rm -rf $HOME/openvscode-server-$RELEASE_TAG-linux-$arch
rm -f $HOME/openvscode-server-$RELEASE_TAG-linux-$arch.tar.gz
