#!/bin/bash
set -e

NPM_DIR="${1:-npm-dist}"

echo "Publishing npm packages..."

# Publish platform packages first
for pkg_dir in "$NPM_DIR"/wm-*; do
  if [[ -d "$pkg_dir" ]]; then
    echo "Publishing $(basename "$pkg_dir")..."
    (cd "$pkg_dir" && npm publish --access public)
  fi
done

# Publish main package last
echo "Publishing wm..."
(cd "$NPM_DIR/wm" && npm publish --access public)

echo "Done!"
