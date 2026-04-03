#!/usr/bin/env bash
#
# Post-processes protoc-gen-go output to align Go json struct tags with the
# protobuf json_name (camelCase) instead of the default proto field name
# (snake_case).  This is necessary so that embedding proto types in DTOs
# produces correct JSON for encoding/json (v0 REST controllers) while
# remaining compatible with protojson (v2 executor).
#
# Usage: ./fix-go-json-tags.sh <file.pb.go>
set -euo pipefail

FILE="${1:?usage: fix-go-json-tags.sh <file.pb.go>}"

perl -i -pe '
  if (/protobuf:"[^"]*json=(\w+)/) {
    my $camel = $1;
    s/\bjson:"\w+(,omitempty)?"/json:"$camel$1"/;
  }
' "$FILE"
