// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaAuthenticationException;
import io.daytona.sdk.exception.DaytonaBadRequestException;
import io.daytona.sdk.exception.DaytonaConflictException;
import io.daytona.sdk.exception.DaytonaConnectionException;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.exception.DaytonaForbiddenException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaRateLimitException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.sdk.exception.DaytonaTimeoutException;
import io.daytona.sdk.exception.DaytonaValidationException;
import org.junit.jupiter.api.Test;

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class ExceptionTypesTest {

    @Test
    void baseExceptionStoresStatusAndImmutableHeaders() {
        Map<String, String> headers = new HashMap<String, String>();
        headers.put("x", "1");

        DaytonaException exception = new DaytonaException(499, "oops", headers);

        assertThat(exception.getStatusCode()).isEqualTo(499);
        assertThat(exception.getHeaders()).containsEntry("x", "1");
        assertThatThrownBy(() -> exception.getHeaders().put("y", "2"))
                .isInstanceOf(UnsupportedOperationException.class);
    }

    @Test
    void baseExceptionStoresCause() {
        IllegalStateException cause = new IllegalStateException("boom");

        DaytonaException exception = new DaytonaException("message", cause);

        assertThat(exception.getStatusCode()).isZero();
        assertThat(exception.getCause()).isSameAs(cause);
    }

    @Test
    void httpExceptionsExposeExpectedStatusCodes() {
        assertThat(new DaytonaBadRequestException("bad").getStatusCode()).isEqualTo(400);
        assertThat(new DaytonaAuthenticationException("auth").getStatusCode()).isEqualTo(401);
        assertThat(new DaytonaForbiddenException("forbidden").getStatusCode()).isEqualTo(403);
        assertThat(new DaytonaNotFoundException("missing").getStatusCode()).isEqualTo(404);
        assertThat(new DaytonaConflictException("conflict").getStatusCode()).isEqualTo(409);
        assertThat(new DaytonaValidationException("invalid").getStatusCode()).isEqualTo(422);
        assertThat(new DaytonaRateLimitException("slow down").getStatusCode()).isEqualTo(429);
        assertThat(new DaytonaServerException(503, "server").getStatusCode()).isEqualTo(503);
    }

    @Test
    void connectionAndTimeoutExceptionsUseGenericStatusCode() {
        assertThat(new DaytonaConnectionException("offline").getStatusCode()).isZero();
        assertThat(new DaytonaTimeoutException("late").getStatusCode()).isZero();
    }

    @Test
    void simpleConstructorsExposeMessages() {
        assertThat(new DaytonaConnectionException("offline", new RuntimeException("cause")).getCause())
                .hasMessage("cause");
        assertThat(new DaytonaTimeoutException("late").getMessage()).isEqualTo("late");
        assertThat(new DaytonaException("plain").getHeaders()).isEqualTo(Collections.<String, String>emptyMap());
    }
}
