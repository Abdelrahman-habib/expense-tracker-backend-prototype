#!/bin/bash

set -e  # Exit on error

# Parse arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --dir) DOCS_DIR="$2"; shift ;;
        --files) FILES="$2"; shift ;;
        --packages) PACKAGES="$2"; shift ;;
        *) echo "❌ Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

# Default values if not provided
DOCS_DIR="${DOCS_DIR:-./docs}"
FILES="${FILES:-swagger.json,swagger.yaml,docs.go}"
PACKAGES="${PACKAGES:-types,payloads,errors}"

echo "🧹 Cleaning package names from Swagger schemas in '$DOCS_DIR'..."

# Convert comma-separated lists to arrays
IFS=',' read -ra FILES_ARRAY <<< "$FILES"
IFS=',' read -ra PACKAGES_ARRAY <<< "$PACKAGES"

# Construct the regex pattern for package names
PATTERN="\b($(IFS='|'; echo "${PACKAGES_ARRAY[*]}"))\.([a-zA-Z0-9_]+)\b"

# Iterate over files and clean
for FILE in "${FILES_ARRAY[@]}"; do
    FULL_PATH="$DOCS_DIR/$FILE"
    
    if [[ -f "$FULL_PATH" ]]; then
        # Replace package names
        sed -i -E "s/$PATTERN/\2/g" "$FULL_PATH"
        
        # Replace "x-nullable" with "nullable"
        sed -i 's/"x-nullable"/"nullable"/g' "$FULL_PATH"
        
        echo "✅ Cleaned: $FULL_PATH"
    else
        echo "⚠️ File not found: $FULL_PATH"
    fi
done

echo "🎉 Swagger docs cleaned successfully!"