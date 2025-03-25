#!/bin/bash
set -e

# This script detects changes in subdirectories with pkg.yaml files
# Usage: ./detect-changes.sh FROM_REF TO_REF

FROM_REF="$1"
TO_REF="$2"

if [ -z "$FROM_REF" ] || [ -z "$TO_REF" ]; then
  echo "Usage: $0 FROM_REF TO_REF"
  echo "Example: $0 main HEAD"
  exit 1
fi

echo "Detecting changes between $FROM_REF and $TO_REF"

# Find all changed files
if git rev-parse "$FROM_REF" >/dev/null 2>&1 && git rev-parse "$TO_REF" >/dev/null 2>&1; then
  CHANGED_FILES=$(git diff --name-only "$FROM_REF" "$TO_REF")
else
  echo "Warning: One or both refs not found. Checking all tracked files."
  CHANGED_FILES=$(git ls-files)
fi

# Debug: Show all changed files
echo "Changed files:"
echo "$CHANGED_FILES"

# Initialize array to store changed directories with pkg.yaml
declare -a CHANGED_DIRS

# Process each changed file - IMPORTANT: Don't use a pipe to while as it creates a subshell
# where the array modifications won't persist to the parent shell
while IFS= read -r file || [[ -n "$file" ]]; do
  # Skip empty lines
  if [ -z "$file" ]; then
    continue
  fi

  # Get the directory of the changed file (first level subdirectory)
  DIR=$(echo "$file" | cut -d'/' -f1)

  # Skip if it's not a direct subdirectory or if it's a dot directory
  if [[ "$DIR" == */* ]] || [[ "$DIR" == .* ]] || [[ "$DIR" == "build" ]]; then
    continue
  fi

  # Check if this directory has a pkg.yaml file
  if [[ -f "$DIR/pkg.yaml" ]]; then
    # Check if pkg.yaml itself changed or any file in this directory changed
    if [[ "$file" == "$DIR/pkg.yaml" ]] || [[ "$file" == "$DIR/"* ]]; then
      # Add to the list if not already there
      if [[ ! " ${CHANGED_DIRS[*]} " =~ " $DIR " ]]; then
        CHANGED_DIRS+=("$DIR")
        echo "Found changed directory with pkg.yaml: $DIR"
      fi
    fi
  fi
done <<< "$CHANGED_FILES"

# Output the changed directories and status
if [ ${#CHANGED_DIRS[@]} -gt 0 ]; then
  echo "CHANGED_DIRS=${CHANGED_DIRS[*]}"
  echo "has_changes=true"
else
  echo "CHANGED_DIRS="
  echo "has_changes=false"
fi

# If GITHUB_OUTPUT is set, write to it for GitHub Actions
if [ -n "$GITHUB_OUTPUT" ]; then
  if [ ${#CHANGED_DIRS[@]} -gt 0 ]; then
    echo "CHANGED_DIRS=${CHANGED_DIRS[*]}" >> $GITHUB_OUTPUT
    echo "has_changes=true" >> $GITHUB_OUTPUT
  else
    echo "CHANGED_DIRS=" >> $GITHUB_OUTPUT
    echo "has_changes=false" >> $GITHUB_OUTPUT
  fi
fi

exit 0
