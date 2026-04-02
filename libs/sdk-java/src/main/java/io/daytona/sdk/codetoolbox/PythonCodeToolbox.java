// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.codetoolbox;

import java.util.Base64;
import java.nio.charset.StandardCharsets;
import java.util.regex.Pattern;

public class PythonCodeToolbox implements CodeToolbox {

    private static final Pattern[] MATPLOTLIB_PATTERNS = {
            Pattern.compile("^[^#]*import\\s+matplotlib", Pattern.MULTILINE),
            Pattern.compile("^[^#]*from\\s+matplotlib", Pattern.MULTILINE),
            Pattern.compile("^[^#]*__import__\\s*\\(\\s*['\"]matplotlib['\"]", Pattern.MULTILINE),
            Pattern.compile("^[^#]*importlib\\.import_module\\s*\\(\\s*['\"]matplotlib['\"]", Pattern.MULTILINE),
    };

    // Matplotlib wrapper: base64-encoded Python that intercepts plt.show() to extract chart
    // metadata as "dtn_artifact_k39fd2:{json}" lines. Kept in sync with TS/Python SDKs —
    // canonical source: libs/sdk-typescript/src/code-toolbox/SandboxPythonCodeToolbox.ts
    private static final String PYTHON_CODE_WRAPPER;
    static {
        String blob = null;
        try {
            java.io.InputStream is = PythonCodeToolbox.class.getResourceAsStream("/python_code_wrapper.b64");
            if (is != null) {
                blob = new String(is.readAllBytes(), StandardCharsets.UTF_8).trim();
            }
        } catch (Exception ignored) {
        }
        if (blob == null || blob.isEmpty()) {
            throw new RuntimeException("python_code_wrapper.b64 resource not found");
        }
        PYTHON_CODE_WRAPPER = blob;
    }

    @Override
    public String getRunCommand(String code) {
        String base64Code = Base64.getEncoder().encodeToString(code.getBytes(StandardCharsets.UTF_8));

        if (isMatplotlibImported(code)) {
            String wrapper = new String(Base64.getDecoder().decode(PYTHON_CODE_WRAPPER), StandardCharsets.UTF_8);
            wrapper = wrapper.replace("{encoded_code}", base64Code);
            base64Code = Base64.getEncoder().encodeToString(wrapper.getBytes(StandardCharsets.UTF_8));
        }

        return "printf '%s' '" + base64Code + "' | base64 -d | python3 -u -";
    }

    private static boolean isMatplotlibImported(String code) {
        for (Pattern pattern : MATPLOTLIB_PATTERNS) {
            if (pattern.matcher(code).find()) {
                return true;
            }
        }
        return false;
    }
}
