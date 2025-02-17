#!/bin/bash

# This script creates a zip archive of the source code of the project removing
# unnecessary files and directories. This is useful for sharing the project with an AI

echo "🔍 Starting project cleanup and packaging..."

# Define variables
PROJECT_NAME=$(basename "$PWD")                # Get current directory name
TMP_ROOT="./.tmp"                              # Use the project's .tmp folder as a base
TEMP_DIR="$TMP_ROOT/${PROJECT_NAME}_temp"      # Temporary directory inside .tmp folder
ZIP_FILE="./source_code.zip"                   # Final zip file in the project directory

# Ensure the .tmp folder exists
mkdir -p "$TMP_ROOT"

# Define directories and files to delete
DIRS_TO_DELETE=(
    ".git"
    ".tmp"
    ".vscode"
    "dev-ssl-certificates"
    "node_modules"
    "static"
)
FILES_TO_DELETE=(
    "go.sum"
    "package-lock.json"
    "*.db*"
    "zip_source_code.sh"
)

# Step 1: Create a temporary directory and copy the project
echo "📂 Creating temporary directory: $TEMP_DIR"
rm -rf "$TEMP_DIR"  # Clean any existing temp folder
rm -f "$ZIP_FILE"   # Remove any existing zip file
mkdir -p "$TEMP_DIR"
# Copy the project to the temporary directory, excluding the zip file and the .tmp folder
rsync -a --progress ./ "$TEMP_DIR" \
    --exclude "$(basename "$ZIP_FILE")" \
    --exclude ".tmp/" > /dev/null

# Step 2: Remove unnecessary directories from the temporary copy
echo "📂 Moving into temporary directory..."
cd "$TEMP_DIR" || exit 1
for dir in "${DIRS_TO_DELETE[@]}"; do
    echo "🗑️  Removing directories named: $dir"
    find . -type d -name "$dir" -exec rm -rf {} + 2>/dev/null
done

# Step 3: Remove unnecessary files from the temporary copy
for file in "${FILES_TO_DELETE[@]}"; do
    echo "🗑️  Removing files matching: $file"
    find . -type f -name "$file" -delete 2>/dev/null
done

# Step 4: Zip the cleaned project contents (without including the temp folder itself)
echo "📦 Creating zip archive: $ZIP_FILE"
# The zip command below zips all visible and hidden files from TEMP_DIR
# into a zip file created in the original project folder.
zip -r "$OLDPWD/$(basename "$ZIP_FILE")" -- * .[!.]* >/dev/null

# Step 5: Remove the temporary directory
echo "🧹 Cleaning up temporary directory..."
cd "$OLDPWD" || exit 1
rm -rf "$TEMP_DIR"

echo "✅ Cleanup and packaging complete! Zip file created at: $ZIP_FILE"
