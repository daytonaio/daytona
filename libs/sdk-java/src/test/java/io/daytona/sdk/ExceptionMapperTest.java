// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaAuthenticationException;
import io.daytona.sdk.exception.DaytonaBadRequestException;
import io.daytona.sdk.exception.DaytonaConflictException;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.exception.DaytonaForbiddenException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaRateLimitException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.sdk.exception.DaytonaValidationException;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class ExceptionMapperTest {

    @Test
    void callMainMapsBadRequest() {
        assertThatThrownBy(() -> ExceptionMapper.callMain(() -> {
            throw new io.daytona.api.client.ApiException(400, "bad", null, "{\"message\":\"invalid\"}");
        })).isInstanceOf(DaytonaBadRequestException.class).hasMessage("invalid");
    }

    @Test
    void callMainMapsAuthentication() {
        assertThatThrownBy(() -> ExceptionMapper.callMain(() -> {
            throw new io.daytona.api.client.ApiException(401, "auth", null, "{\"message\":\"denied\"}");
        })).isInstanceOf(DaytonaAuthenticationException.class).hasMessage("denied");
    }

    @Test
    void callToolboxMapsForbiddenAndNotFound() {
        assertThatThrownBy(() -> ExceptionMapper.callToolbox(() -> {
            throw new io.daytona.toolbox.client.ApiException(403, "forbidden", null, "{\"error\":\"blocked\"}");
        })).isInstanceOf(DaytonaForbiddenException.class).hasMessage("blocked");

        assertThatThrownBy(() -> ExceptionMapper.callToolbox(() -> {
            throw new io.daytona.toolbox.client.ApiException(404, "missing", null, "{\"message\":\"gone\"}");
        })).isInstanceOf(DaytonaNotFoundException.class).hasMessage("gone");
    }

    @Test
    void mapsConflictValidationAndRateLimit() {
        assertThatThrownBy(() -> ExceptionMapper.callMain(() -> {
            throw new io.daytona.api.client.ApiException(409, "conflict", null, "{\"message\":\"exists\"}");
        })).isInstanceOf(DaytonaConflictException.class).hasMessage("exists");

        assertThatThrownBy(() -> ExceptionMapper.callMain(() -> {
            throw new io.daytona.api.client.ApiException(422, "invalid", null, "{\"message\":\"bad data\"}");
        })).isInstanceOf(DaytonaValidationException.class).hasMessage("bad data");

        assertThatThrownBy(() -> ExceptionMapper.callMain(() -> {
            throw new io.daytona.api.client.ApiException(429, "limit", null, "{\"message\":\"too many\"}");
        })).isInstanceOf(DaytonaRateLimitException.class).hasMessage("too many");
    }

    @Test
    void mapsServerAndGenericStatuses() {
        assertThatThrownBy(() -> ExceptionMapper.callMain(() -> {
            throw new io.daytona.api.client.ApiException(503, "server", null, "{\"message\":\"retry\"}");
        })).isInstanceOf(DaytonaServerException.class).hasMessage("retry");

        assertThatThrownBy(() -> ExceptionMapper.callMain(() -> {
            throw new io.daytona.api.client.ApiException(418, "teapot", null, "raw body");
        })).isInstanceOf(DaytonaException.class).satisfies(error -> {
            DaytonaException exception = (DaytonaException) error;
            assertThat(exception.getStatusCode()).isEqualTo(418);
            assertThat(exception.getMessage()).isEqualTo("raw body");
        });
    }

    @Test
    void usesFallbackMessageWhenBodyMissing() {
        assertThatThrownBy(() -> ExceptionMapper.callToolbox(() -> {
            throw new io.daytona.toolbox.client.ApiException(500, "server", null, null);
        })).isInstanceOf(DaytonaServerException.class).hasMessage("Request failed with status 500");
    }

    @Test
    void extractsErrorFieldAndRawBodyWhenMessageMissing() {
        assertThatThrownBy(() -> ExceptionMapper.callMain(() -> {
            throw new io.daytona.api.client.ApiException(404, "missing", null, "{\"error\":\"gone\"}");
        })).isInstanceOf(DaytonaNotFoundException.class).hasMessage("gone");

        assertThatThrownBy(() -> ExceptionMapper.callToolbox(() -> {
            throw new io.daytona.toolbox.client.ApiException(418, "teapot", null, "not-json");
        })).isInstanceOf(DaytonaException.class).hasMessage("not-json");
    }

    @Test
    void preservesEscapedJsonMessageContent() {
        assertThatThrownBy(() -> ExceptionMapper.callMain(() -> {
            throw new io.daytona.api.client.ApiException(400, "bad", null, "{\"message\":\"invalid \\\"value\\\"\"}");
        })).isInstanceOf(DaytonaBadRequestException.class).hasMessage("invalid \\\"value\\\"");
    }

    @Test
    void runHelpersMapApiExceptions() {
        assertThatThrownBy(() -> ExceptionMapper.runMain(() -> {
            throw new io.daytona.api.client.ApiException(409, "conflict", null, "{\"message\":\"exists\"}");
        })).isInstanceOf(DaytonaConflictException.class).hasMessage("exists");

        assertThatThrownBy(() -> ExceptionMapper.runToolbox(() -> {
            throw new io.daytona.toolbox.client.ApiException(403, "forbidden", null, "{\"message\":\"blocked\"}");
        })).isInstanceOf(DaytonaForbiddenException.class).hasMessage("blocked");
    }

    @Test
    void runHelpersExecuteSuccessfulCallbacks() {
        String value = ExceptionMapper.callMain(() -> "ok");
        ExceptionMapper.runMain(() -> { });
        ExceptionMapper.runToolbox(() -> { });

        assertThat(value).isEqualTo("ok");
    }
}
