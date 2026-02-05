#!/bin/bash
set -e

# Local npm publish script (requires OTP input for 2FA accounts)
# Usage: ./scripts/publish-npm-local.sh [version]

VERSION="${1:-}"
DIST_DIR="dist-download"
EXTRACTED_DIR="dist-extracted"
NPM_DIR="npm-dist"

if [[ -z "$VERSION" ]]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 0.1.0"
  exit 1
fi

echo "=== Publishing gitwm v$VERSION to npm ==="
echo ""

# Check npm login
if ! npm whoami &>/dev/null; then
  echo "Not logged in to npm. Please run: npm login"
  exit 1
fi

echo "Logged in as: $(npm whoami)"
echo ""

# Download from GitHub release
echo "=== Downloading binaries from GitHub release v$VERSION ==="
rm -rf "$DIST_DIR" "$EXTRACTED_DIR" "$NPM_DIR"
mkdir -p "$DIST_DIR"

gh release download "v$VERSION" -D "$DIST_DIR" --pattern '*.tar.gz' --pattern '*.zip'

# Extract
echo "=== Extracting binaries ==="
mkdir -p "$EXTRACTED_DIR"

for f in "$DIST_DIR"/*.tar.gz; do
  name=$(basename "$f" .tar.gz)
  mkdir -p "$EXTRACTED_DIR/$name"
  tar -xzf "$f" -C "$EXTRACTED_DIR/$name"
done

mkdir -p "$EXTRACTED_DIR/wm_windows_amd64"
unzip -q "$DIST_DIR/wm_windows_amd64.zip" -d "$EXTRACTED_DIR/wm_windows_amd64"

# Build npm packages
echo "=== Building npm packages ==="
mkdir -p "$NPM_DIR"

# darwin-arm64
mkdir -p "$NPM_DIR/gitwm-darwin-arm64/bin"
cp "$EXTRACTED_DIR/wm_darwin_arm64/wm" "$NPM_DIR/gitwm-darwin-arm64/bin/"
chmod +x "$NPM_DIR/gitwm-darwin-arm64/bin/wm"
cat > "$NPM_DIR/gitwm-darwin-arm64/package.json" << EOF
{"name":"gitwm-darwin-arm64","version":"$VERSION","os":["darwin"],"cpu":["arm64"],"license":"MIT","files":["bin"]}
EOF

# darwin-x64
mkdir -p "$NPM_DIR/gitwm-darwin-x64/bin"
cp "$EXTRACTED_DIR/wm_darwin_amd64/wm" "$NPM_DIR/gitwm-darwin-x64/bin/"
chmod +x "$NPM_DIR/gitwm-darwin-x64/bin/wm"
cat > "$NPM_DIR/gitwm-darwin-x64/package.json" << EOF
{"name":"gitwm-darwin-x64","version":"$VERSION","os":["darwin"],"cpu":["x64"],"license":"MIT","files":["bin"]}
EOF

# linux-arm64
mkdir -p "$NPM_DIR/gitwm-linux-arm64/bin"
cp "$EXTRACTED_DIR/wm_linux_arm64/wm" "$NPM_DIR/gitwm-linux-arm64/bin/"
chmod +x "$NPM_DIR/gitwm-linux-arm64/bin/wm"
cat > "$NPM_DIR/gitwm-linux-arm64/package.json" << EOF
{"name":"gitwm-linux-arm64","version":"$VERSION","os":["linux"],"cpu":["arm64"],"license":"MIT","files":["bin"]}
EOF

# linux-x64
mkdir -p "$NPM_DIR/gitwm-linux-x64/bin"
cp "$EXTRACTED_DIR/wm_linux_amd64/wm" "$NPM_DIR/gitwm-linux-x64/bin/"
chmod +x "$NPM_DIR/gitwm-linux-x64/bin/wm"
cat > "$NPM_DIR/gitwm-linux-x64/package.json" << EOF
{"name":"gitwm-linux-x64","version":"$VERSION","os":["linux"],"cpu":["x64"],"license":"MIT","files":["bin"]}
EOF

# win32-x64
mkdir -p "$NPM_DIR/gitwm-win32-x64/bin"
cp "$EXTRACTED_DIR/wm_windows_amd64/wm.exe" "$NPM_DIR/gitwm-win32-x64/bin/"
cat > "$NPM_DIR/gitwm-win32-x64/package.json" << EOF
{"name":"gitwm-win32-x64","version":"$VERSION","os":["win32"],"cpu":["x64"],"license":"MIT","files":["bin"]}
EOF

# main package
mkdir -p "$NPM_DIR/gitwm/bin"
cp npm/wm/bin/wm "$NPM_DIR/gitwm/bin/"
chmod +x "$NPM_DIR/gitwm/bin/wm"
cat > "$NPM_DIR/gitwm/package.json" << EOF
{"name":"gitwm","version":"$VERSION","description":"git worktree manager","license":"MIT","bin":{"wm":"bin/wm"},"files":["bin"],"optionalDependencies":{"gitwm-darwin-arm64":"$VERSION","gitwm-darwin-x64":"$VERSION","gitwm-linux-arm64":"$VERSION","gitwm-linux-x64":"$VERSION","gitwm-win32-x64":"$VERSION"}}
EOF

echo "=== Publishing to npm (OTP required for each package) ==="
echo ""

# Publish platform packages first
for pkg in gitwm-darwin-arm64 gitwm-darwin-x64 gitwm-linux-arm64 gitwm-linux-x64 gitwm-win32-x64; do
  echo "Publishing $pkg..."
  (cd "$NPM_DIR/$pkg" && npm publish --access public)
  echo ""
done

# Publish main package
echo "Publishing gitwm..."
(cd "$NPM_DIR/gitwm" && npm publish --access public)

echo ""
echo "=== Done! ==="
echo "Install with: npm install -g gitwm"

# Cleanup
rm -rf "$DIST_DIR" "$EXTRACTED_DIR" "$NPM_DIR"
