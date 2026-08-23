#!/bin/bash
# Download and extract Buster extension for Playwright
# Run this once before first use: ./scripts/download-buster.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
EXT_DIR="$ROOT_DIR/extensions/buster"
EXT_ID="mpbjkejclgfgadiemmefgebjfooflfhl"
CHROME_VERSION="125.0.6422.60"

if [ -f "$EXT_DIR/manifest.json" ]; then
  echo "Buster extension already exists at $EXT_DIR"
  exit 0
fi

echo "Downloading Buster extension from Chrome Web Store..."

CRX_URL="https://clients2.google.com/service/update2/crx?response=redirect&os=mac&arch=arm64&os_arch=arm64&nacl_arch=arm&prod=chromecrx&prodchannel=&prodversion=${CHROME_VERSION}&lang=en&acceptformat=crx3&x=id%3D${EXT_ID}%26installsource%3Dondemand%26uc"

TMP_CRX="$ROOT_DIR/extensions/buster.crx"
mkdir -p "$ROOT_DIR/extensions"

curl -sL "$CRX_URL" -o "$TMP_CRX"

if ! file "$TMP_CRX" | grep -q "Chrome extension"; then
  echo "ERROR: Downloaded file is not a valid CRX."
  echo "Try downloading manually from Chrome Web Store and extracting to $EXT_DIR"
  rm -f "$TMP_CRX"
  exit 1
fi

echo "Extracting..."
mkdir -p "$EXT_DIR"
unzip -qo "$TMP_CRX" -d "$EXT_DIR"
rm -f "$TMP_CRX"

# Clean up Chrome Web Store metadata (causes issues with sideloading)
rm -rf "$EXT_DIR/_metadata"
# Remove update_url from manifest (prevents auto-update attempts)
sed -i '' 's/"update_url": "[^"]*",//' "$EXT_DIR/manifest.json" 2>/dev/null || \
  sed -i 's/"update_url": "[^"]*",//' "$EXT_DIR/manifest.json"

echo "Done. Buster extension installed at: $EXT_DIR"
echo "Version: $(grep -o '"version":"[^"]*"' "$EXT_DIR/manifest.json")"
