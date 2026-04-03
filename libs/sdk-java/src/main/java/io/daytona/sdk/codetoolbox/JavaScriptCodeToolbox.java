// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.codetoolbox;

import java.util.Base64;
import java.nio.charset.StandardCharsets;

public class JavaScriptCodeToolbox implements CodeToolbox {
    @Override
    public String getRunCommand(String code) {
        String base64Code = Base64.getEncoder().encodeToString(
                ("process.argv.splice(1, 1);\n" + code).getBytes(StandardCharsets.UTF_8));
        return "printf '%s' '" + base64Code + "' | base64 -d | node -";
    }
}
