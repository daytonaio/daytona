#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <file>"
  exit 1
fi

FILE="$1"
TEMP_FILE=$(mktemp)

inside_code_block=false
indent_level=0
buffer=""

while IFS= read -r line || [[ -n "$line" ]]; do
  # Check if line contains a code fence (```)
  if [[ "$line" =~ ^[[:space:]]*\`\`\`.* ]]; then
    # Extract just the language identifier if present (everything after ```)
    lang_identifier=$(echo "$line" | sed -E 's/^[[:space:]]*```(.*)/\1/')

    if $inside_code_block; then
      # Closing fence: print dedented buffer and the closing fence
      if [[ -n "$buffer" ]]; then
        # Remove trailing newline from buffer
        buffer=${buffer%$'\n'}
        while IFS= read -r buf_line; do
          if [[ -n "$buf_line" ]]; then
            # Only dedent non-empty lines
            dedented_line=$(echo "$buf_line" | sed -E "s/^[[:space:]]{$indent_level}//")
            echo "$dedented_line"
          else
            # Preserve empty lines
            echo ""
          fi
        done <<<"$buffer"
      fi
      buffer=""
      echo '```'
      inside_code_block=false
      indent_level=0
    else
      # Opening fence: print it with language identifier if present
      if [[ -n "$lang_identifier" ]]; then
        echo '```'"${lang_identifier}"
      else
        echo '```'
      fi
      inside_code_block=true
      indent_level=9999 # Reset for new block
    fi
  elif $inside_code_block; then
    # Inside code block: collect lines and track minimum indentation
    if [[ -n "$line" ]]; then
      # Count leading spaces for non-empty lines
      current_indent=$(echo "$line" | sed -E 's/^([[:space:]]*).*$/\1/')
      indent_length=${#current_indent}
      if [[ $indent_length -lt $indent_level ]]; then
        indent_level=$indent_length
      fi
    fi
    buffer+="$line"$'\n'
  else
    # Outside code block: print normally
    echo "$line"
  fi
done <"$FILE" >"$TEMP_FILE"

mv "$TEMP_FILE" "$FILE"

echo "Fixed code indentation in: $FILE"
