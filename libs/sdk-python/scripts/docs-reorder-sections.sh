#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# Check if a file argument is provided
[ -z "$1" ] && echo "Usage: $0 <file.mdx> [section_title]" && exit 1

# Check if section title is provided, if not exit
[ -z "$2" ] && echo "No section title provided, exiting." && exit 0

FILE="$1"
SECTION_TITLE="$2"
TEMP_FILE=$(mktemp)
RESULT_FILE=$(mktemp)

# First, extract the frontmatter and content before any sections or anchors
awk '
BEGIN { in_frontmatter = 0; found_content = 0 }
/^---$/ { 
    if (in_frontmatter) {
        print
        in_frontmatter = 0
        next
    } else {
        in_frontmatter = 1
        print
        next
    }
}
in_frontmatter { print; next }
/<a id=|^## / { exit }
{ print }
' "$FILE" >"$TEMP_FILE"

# Now process the sections and their content
awk -v section="$SECTION_TITLE" '
BEGIN {
    in_target = 0
    found_target = 0
    target_content = ""
    other_content = ""
    anchor = ""
    in_frontmatter = 0
    passed_frontmatter = 0
}

# Skip frontmatter
/^---$/ {
    if (!in_frontmatter) {
        in_frontmatter = 1
        next
    } else {
        in_frontmatter = 0
        passed_frontmatter = 1
        next
    }
}
in_frontmatter { next }
!passed_frontmatter { next }

# Store potential anchor for next section
/<a id=/ {
    anchor = $0
    next
}

# When we hit a section header
/^## / {
    # If this is our target section
    if ($0 ~ "^## " section "$") {
        # Store this section (with its anchor if we just saw one)
        if (anchor != "") {
            target_content = anchor "\n" $0 "\n"
            anchor = ""
        } else {
            target_content = $0 "\n"
        }
        in_target = 1
        found_target = 1
    } else {
        # For other sections, add them to other_content
        if (anchor != "") {
            other_content = other_content anchor "\n" $0 "\n"
            anchor = ""
        } else {
            other_content = other_content $0 "\n"
        }
        in_target = 0
    }
    next
}

# Handle content lines
{
    if (in_target) {
        target_content = target_content $0 "\n"
    } else if (passed_frontmatter) {
        other_content = other_content $0 "\n"
    }
}

END {
    # Print in desired order
    printf "%s", target_content
    printf "%s", other_content
}' "$FILE" >"$RESULT_FILE"

# Combine the files
cat "$TEMP_FILE" "$RESULT_FILE" >"$FILE"

# Cleanup
rm -f "$TEMP_FILE" "$RESULT_FILE"

echo "Successfully reordered sections in $FILE, moved '$SECTION_TITLE' to the top."
