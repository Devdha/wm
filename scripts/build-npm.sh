#!/bin/bash
set -e

VERSION="${1:-0.0.0}"
DIST_DIR="${2:-dist}"
NPM_DIR="npm-dist"

# Platform mapping: goreleaser output -> npm package
declare -A PLATFORMS=(
  ["darwin_arm64"]="darwin-arm64 darwin arm64"
  ["darwin_amd64"]="darwin-x64 darwin x64"
  ["linux_arm64"]="linux-arm64 linux arm64"
  ["linux_amd64"]="linux-x64 linux x64"
  ["windows_amd64"]="win32-x64 win32 x64"
)

rm -rf "$NPM_DIR"
mkdir -p "$NPM_DIR"

# Build platform-specific packages
for goreleaser_name in "${!PLATFORMS[@]}"; do
  read -r npm_suffix os cpu <<< "${PLATFORMS[$goreleaser_name]}"

  pkg_name="@devdha/wm-${npm_suffix}"
  pkg_dir="$NPM_DIR/wm-${npm_suffix}"

  echo "Building $pkg_name..."

  mkdir -p "$pkg_dir/bin"

  # Find the binary
  if [[ "$os" == "win32" ]]; then
    binary_name="wm.exe"
    archive=$(find "$DIST_DIR" -name "wm_windows_amd64*" -type d | head -1)
  else
    binary_name="wm"
    archive=$(find "$DIST_DIR" -name "wm_${goreleaser_name}*" -type d | head -1)
  fi

  if [[ -z "$archive" ]]; then
    echo "Warning: No binary found for $goreleaser_name, skipping..."
    continue
  fi

  cp "$archive/$binary_name" "$pkg_dir/bin/"
  chmod +x "$pkg_dir/bin/$binary_name"

  # Create package.json
  cat > "$pkg_dir/package.json" << EOF
{
  "name": "$pkg_name",
  "version": "$VERSION",
  "description": "wm binary for $os $cpu",
  "repository": {
    "type": "git",
    "url": "https://github.com/Devdha/wm.git"
  },
  "license": "MIT",
  "os": ["$os"],
  "cpu": ["$cpu"],
  "bin": {
    "wm": "bin/$binary_name"
  },
  "files": ["bin"]
}
EOF

  echo "Created $pkg_name"
done

# Build main package
echo "Building @devdha/wm..."
pkg_dir="$NPM_DIR/wm"
mkdir -p "$pkg_dir/bin"
cp npm/wm/bin/wm "$pkg_dir/bin/"
chmod +x "$pkg_dir/bin/wm"

# Update version in main package
cat > "$pkg_dir/package.json" << EOF
{
  "name": "@devdha/wm",
  "version": "$VERSION",
  "description": "git worktree manager",
  "repository": {
    "type": "git",
    "url": "https://github.com/Devdha/wm.git"
  },
  "license": "MIT",
  "bin": {
    "wm": "bin/wm"
  },
  "files": ["bin"],
  "optionalDependencies": {
    "@devdha/wm-darwin-arm64": "$VERSION",
    "@devdha/wm-darwin-x64": "$VERSION",
    "@devdha/wm-linux-arm64": "$VERSION",
    "@devdha/wm-linux-x64": "$VERSION",
    "@devdha/wm-win32-x64": "$VERSION"
  }
}
EOF

echo "Created @devdha/wm"
echo ""
echo "All packages built in $NPM_DIR/"
ls -la "$NPM_DIR/"
