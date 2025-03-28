#!/usr/bin/env python3
"""
This script detects changes in subdirectories with lib.yaml files
Usage: ./detect-changes.py FROM_REF TO_REF
"""

import os
import sys
import subprocess
import yaml

def run_command(cmd):
    """Run a shell command and return its output"""
    try:
        result = subprocess.run(cmd, shell=True, check=True, capture_output=True, text=True)
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        print(f"Error running command: {cmd}")
        print(f"Error: {e.stderr}")
        return ""

def check_pkg_name_match(directory):
    """Check if lib.yaml name matches directory name"""
    pkg_yaml_path = os.path.join(directory, "lib.yaml")
    if os.path.isfile(pkg_yaml_path):
        try:
            with open(pkg_yaml_path, 'r') as f:
                pkg_data = yaml.safe_load(f)
                pkg_name = pkg_data.get('name', '')

                if pkg_name != directory:
                    print(f"Error: Package name '{pkg_name}' in {pkg_yaml_path} does not match directory name '{directory}'")
                    return False
        except Exception as e:
            print(f"Error reading {pkg_yaml_path}: {e}")
            return False
    return True

def main():
    if len(sys.argv) < 3:
        print(f"Usage: {sys.argv[0]} FROM_REF TO_REF")
        print(f"Example: {sys.argv[0]} main HEAD")
        sys.exit(1)

    from_ref = sys.argv[1]
    to_ref = sys.argv[2]

    print(f"Detecting changes between {from_ref} and {to_ref}")

    # Find all changed files
    if run_command(f"git rev-parse {from_ref} > /dev/null 2>&1") and run_command(f"git rev-parse {to_ref} > /dev/null 2>&1"):
        changed_files = run_command(f"git diff --name-only {from_ref} {to_ref}")
    else:
        print("Warning: One or both refs not found. Checking all tracked files.")
        changed_files = run_command("git ls-files")

    # Debug: Show all changed files
    print("Changed files:")
    print(changed_files)

    # Initialize set to store changed directories with lib.yaml
    changed_dirs = set()

    # Process each changed file
    for file in changed_files.splitlines():
        if not file:
            continue

        # Get the directory of the changed file (first level subdirectory)
        directory = file.split('/')[0] if '/' in file else file

        # Skip if it's not a direct subdirectory or if it's a dot directory
        if '/' in directory or directory.startswith('.') or directory == "build":
            continue

        # Check if this directory has a lib.yaml file
        pkg_yaml_path = os.path.join(directory, "lib.yaml")
        if os.path.isfile(pkg_yaml_path):
            # Check if lib.yaml name matches directory name
            if not check_pkg_name_match(directory):
                sys.exit(1)

            # Check if lib.yaml itself changed or any file in this directory changed
            if file == pkg_yaml_path or file.startswith(f"{directory}/"):
                # Add to the list if not already there
                if directory not in changed_dirs:
                    changed_dirs.add(directory)
                    print(f"Found changed directory with lib.yaml: {directory}")

    # Check all directories with lib.yaml files, not just changed ones
    for directory in [d for d in os.listdir('.') if os.path.isdir(d) and not d.startswith('.') and d != 'build']:
        if os.path.isfile(os.path.join(directory, "lib.yaml")):
            if not check_pkg_name_match(directory):
                sys.exit(1)

    # Output the changed directories and status
    if changed_dirs:
        changed_dirs_str = ' '.join(changed_dirs)
        print(f"CHANGED_DIRS={changed_dirs_str}")
        print("has_changes=true")
    else:
        print("CHANGED_DIRS=")
        print("has_changes=false")

    # If GITHUB_OUTPUT is set, write to it for GitHub Actions
    github_output = os.environ.get('GITHUB_OUTPUT')
    if github_output:
        with open(github_output, 'a') as f:
            if changed_dirs:
                changed_dirs_str = ' '.join(changed_dirs)
                f.write(f"CHANGED_DIRS={changed_dirs_str}\n")
                f.write("has_changes=true\n")
            else:
                f.write("CHANGED_DIRS=\n")
                f.write("has_changes=false\n")

    sys.exit(0)

if __name__ == "__main__":
    main()
