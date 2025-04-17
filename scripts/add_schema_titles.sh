#!/bin/bash

# Default values
DIR="./docs"
FILES="swagger.json"

# Parse command-line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --dir) DIR="$2"; shift ;;
        --files) FILES="$2"; shift ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

echo "🔍 Scanning OpenAPI specs in '$DIR' for schemas and adding titles..."

update_json() {
    local file=$1
    echo "📝 Processing JSON file: $file"

    awk '
    BEGIN {
        in_schemas = 0
        def_depth = 0
        current_depth = 0
    }
    
    /{/ { current_depth++ }
    /}/ { 
        current_depth--
        if (in_schemas && current_depth < def_depth) {
            in_schemas = 0
        }
    }
    
    /"schemas": \{/ {
        in_schemas = 1
        def_depth = current_depth
        print
        next
    }
    
    in_schemas && current_depth == (def_depth + 1) && match($0, /^([[:space:]]*)"([^"]+)": \{/, arr) {
        schema_name = arr[2]
        indent = arr[1]
        print $0
        print indent "  \"title\": \"" schema_name " Schema\","
        next
    }
    
    { print }
    ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"

    echo "✅ JSON file updated: $file"
}

IFS=',' read -r -a FILE_LIST <<< "$FILES"

for file in "${FILE_LIST[@]}"; do
    file_path="$DIR/$file"
    if [ -f "$file_path" ]; then
        case "$file" in
            *.json) update_json "$file_path" ;;
            *) echo "⚠️ Unsupported file type: $file" ;;
        esac
    else
        echo "⚠️ File not found: $file_path"
    fi
done

echo "🎉 Schema titles added successfully!"