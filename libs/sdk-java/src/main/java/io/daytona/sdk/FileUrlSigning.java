// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaException;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.security.InvalidKeyException;
import java.security.NoSuchAlgorithmException;
import java.time.Instant;
import java.util.Base64;

final class FileUrlSigning {
    private static final String SIGNATURE_V1_PREFIX = "v1_";
    private static final long DEFAULT_TTL_SECONDS = 3600;

    private FileUrlSigning() {
    }

    static String computeFileUrlSignature(String signingKey, String method, String path, long expires) {
        String canonical = "v1:files:" + method + ":" + path + ":" + expires;
        try {
            Mac mac = Mac.getInstance("HmacSHA256");
            mac.init(new SecretKeySpec(signingKey.getBytes(StandardCharsets.UTF_8), "HmacSHA256"));
            byte[] digest = mac.doFinal(canonical.getBytes(StandardCharsets.UTF_8));
            return SIGNATURE_V1_PREFIX + Base64.getUrlEncoder().withoutPadding().encodeToString(digest);
        } catch (NoSuchAlgorithmException | InvalidKeyException e) {
            throw new DaytonaException("Failed to compute file URL signature", e);
        }
    }

    static long resolveExpires(Long ttlSeconds) {
        if (ttlSeconds == null) {
            return Instant.now().getEpochSecond() + DEFAULT_TTL_SECONDS;
        }
        if (ttlSeconds <= 0) {
            return 0;
        }
        return Instant.now().getEpochSecond() + ttlSeconds;
    }

    static String buildSignedFileUrl(
            String toolboxProxyUrl,
            String sandboxId,
            String operationPath,
            String method,
            String filePath,
            String signingKey,
            Long ttlSeconds) {
        if (signingKey == null || signingKey.isEmpty()) {
            throw new DaytonaException(
                    "Sandbox signing key is not available. Call refreshData() or fetch the sandbox by ID to load it.");
        }

        long expires = resolveExpires(ttlSeconds);
        String signature = computeFileUrlSignature(signingKey, method, filePath, expires);
        String encodedPath = URLEncoder.encode(filePath, StandardCharsets.UTF_8);
        String encodedExpires = URLEncoder.encode(String.valueOf(expires), StandardCharsets.UTF_8);
        String encodedSignature = URLEncoder.encode(signature, StandardCharsets.UTF_8);

        return trimTrailingSlash(toolboxProxyUrl) + "/" + sandboxId + operationPath
                + "?path=" + encodedPath
                + "&expires=" + encodedExpires
                + "&signature=" + encodedSignature;
    }

    private static String trimTrailingSlash(String value) {
        if (value == null) {
            return "";
        }
        String output = value;
        while (output.endsWith("/")) {
            output = output.substring(0, output.length() - 1);
        }
        return output;
    }
}
