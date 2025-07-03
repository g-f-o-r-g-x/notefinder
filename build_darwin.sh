#!/bin/bash

set -e

APP_NAME="Notefinder"
APP_ID="org.notefinder"
APP_VERSION="0.1a"
ICON_PATH="../../images/notefinder.png"
DMG_NAME="${APP_NAME}-${APP_VERSION}.dmg"
BUILD_DIR="../../build"
STAGING_DIR="${BUILD_DIR}/dmg-staging"

echo "📦 Building macOS .app bundle for $APP_NAME..."
cd cmd/notefinder

# Clean previous artifacts
rm -rf "${BUILD_DIR}"
rm -f "${FINAL_DMG_PATH}"
mkdir -p "${STAGING_DIR}"

# Package with fyne
fyne package --os darwin --icon "${ICON_PATH}" --name "${APP_NAME}" --app-id "${APP_ID}"

# Move .app into staging folder
mv "${APP_NAME}.app" "${STAGING_DIR}/"

cd ../..

exit 0

echo "💽 Creating DMG using hdiutil..."

# Create the DMG
hdiutil create -volname "${APP_NAME}" \
  -srcfolder "${STAGING_DIR}" \
  -ov -format UDZO "${DMG_NAME}"
