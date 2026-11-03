#!/bin/sh
# Builds Time Pilot.app (universal: Apple Silicon + Intel) on a Mac.
# Needs Go 1.24+ and the Xcode command line tools (xcode-select --install).
set -e
cd "$(dirname "$0")/.."
OUT=dist/mac
rm -rf "$OUT"; mkdir -p "$OUT/Time Pilot.app/Contents/MacOS" "$OUT/Time Pilot.app/Contents/Resources"
CGO_ENABLED=1 GOARCH=arm64 go build -ldflags "-s -w" -o "$OUT/tp-arm64" ./cmd/timepilot
CGO_ENABLED=1 GOARCH=amd64 go build -ldflags "-s -w" -o "$OUT/tp-amd64" ./cmd/timepilot
lipo -create -output "$OUT/Time Pilot.app/Contents/MacOS/timepilot" "$OUT/tp-arm64" "$OUT/tp-amd64"
rm "$OUT/tp-arm64" "$OUT/tp-amd64"
cat > "$OUT/Time Pilot.app/Contents/Info.plist" <<PL
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleExecutable</key><string>timepilot</string>
<key>CFBundleIdentifier</key><string>local.timepilot</string>
<key>CFBundleName</key><string>Time Pilot</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleVersion</key><string>1</string>
<key>NSHighResolutionCapable</key><true/>
</dict></plist>
PL
codesign --force --deep -s - "$OUT/Time Pilot.app"
echo "Built $OUT/Time Pilot.app -- put timeplt.zip next to it (or in ~)."
