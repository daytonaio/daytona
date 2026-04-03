// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.codetoolbox;

import java.util.Base64;
import java.nio.charset.StandardCharsets;

public class TypeScriptCodeToolbox implements CodeToolbox {
    @Override
    public String getRunCommand(String code) {
        String base64Code = Base64.getEncoder().encodeToString(
                ("process.argv.splice(1, 1);\n" + code).getBytes(StandardCharsets.UTF_8));
        return "_f=/tmp/dtn_$$.ts; "
                + "printf '%s' '" + base64Code + "' | base64 -d > \"$_f\"; "
                + "npm_config_loglevel=error npx ts-node -T --ignore-diagnostics 5107 -O '{\"module\":\"CommonJS\"}' \"$_f\"; "
                + "_dtn_ec=$?; rm -f \"$_f\"; exit $_dtn_ec";
    }
}
