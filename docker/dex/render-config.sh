#!/bin/sh

set -eu

template_path=${1:-/etc/dex/config.yaml.tmpl}
output_path=${2:-/tmp/dex-config.yaml}

require_env() {
  var_name=$1
  eval "value=\${$var_name:-}"
  if [ -z "$value" ]; then
    echo "Missing required environment variable: $var_name" >&2
    exit 1
  fi
}

escape_sed() {
  printf '%s' "$1" | sed 's/[\/&]/\\&/g'
}

require_env DAYTONA_DASHBOARD_BASE_URL
require_env DAYTONA_PROXY_BASE_URL
require_env DAYTONA_OIDC_ISSUER

dashboard_base_url_escaped=$(escape_sed "$DAYTONA_DASHBOARD_BASE_URL")
proxy_base_url_escaped=$(escape_sed "$DAYTONA_PROXY_BASE_URL")
oidc_issuer_escaped=$(escape_sed "$DAYTONA_OIDC_ISSUER")

sed \
  -e "s/__DAYTONA_DASHBOARD_BASE_URL__/$dashboard_base_url_escaped/g" \
  -e "s/__DAYTONA_PROXY_BASE_URL__/$proxy_base_url_escaped/g" \
  -e "s/__DAYTONA_OIDC_ISSUER__/$oidc_issuer_escaped/g" \
  "$template_path" > "$output_path"
