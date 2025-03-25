#!/bin/bash
# package.sh - Package prebuilt libraries for distribution
# Usage: package.sh <prebuilt_dir> <output_dir> [package_name] [version]
# Example: package.sh _prebuilt /tmp mypackage 1.0.0

set -e

PREBUILT_DIR="$1"
OUTPUT_DIR="$2"
PACKAGE_NAME="$3"
VERSION="$4"

if [ -z "$PREBUILT_DIR" ] || [ -z "$OUTPUT_DIR" ]; then
  echo "Usage: package.sh <prebuilt_dir> <output_dir> [package_name] [version]"
  exit 1
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Check if prebuilt directory exists
if [ ! -d "$PREBUILT_DIR" ]; then
  echo "Error: Prebuilt directory $PREBUILT_DIR does not exist"
  exit 1
fi

# Get current directory name if package name not provided
if [ -z "$PACKAGE_NAME" ]; then
  PACKAGE_NAME=$(basename "$(pwd)")
fi

echo "Packaging targets from $PREBUILT_DIR for package $PACKAGE_NAME"

# Find all subdirectories in the prebuilt directory (these are the targets)
for TARGET_DIR in "$PREBUILT_DIR"/*; do
  if [ -d "$TARGET_DIR" ]; then
    TARGET_NAME=$(basename "$TARGET_DIR")
    
    # Create artifact name
    if [ -n "$VERSION" ]; then
      ARTIFACT_NAME="${PACKAGE_NAME}-${VERSION}-${TARGET_NAME}"
    else
      ARTIFACT_NAME="${PACKAGE_NAME}-${TARGET_NAME}"
    fi
    
    echo "Packaging target: $TARGET_NAME as $ARTIFACT_NAME.tar.gz"
    
    # Create tarball
    tar -czf "$OUTPUT_DIR/${ARTIFACT_NAME}.tar.gz" -C "$PREBUILT_DIR" "$TARGET_NAME"
    
    echo "Created: $OUTPUT_DIR/${ARTIFACT_NAME}.tar.gz"
  fi
done

echo "Packaging complete. Files are available in $OUTPUT_DIR"
