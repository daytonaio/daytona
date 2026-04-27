// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.model.SandboxState;
import org.junit.jupiter.api.Test;

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class DaytonaUtilitiesTest {

    @Test
    void urlEncodePathSegmentEscapesSpacesAndSlashes() {
        assertThat(Daytona.urlEncodePathSegment("dir name/file.txt")).isEqualTo("dir%20name%2Ffile.txt");
    }

    @Test
    void urlEncodeQueryTreatsNullAsEmptyString() {
        assertThat(Daytona.urlEncodeQuery(null)).isEmpty();
        assertThat(Daytona.urlEncodeQuery("a b")).isEqualTo("a+b");
    }

    @Test
    void castStringMapConvertsKeysAndNullValues() {
        Map<Object, Object> raw = new HashMap<Object, Object>();
        raw.put(1, null);
        raw.put("two", 2);

        assertThat(Daytona.castStringMap(raw))
                .containsEntry("1", "")
                .containsEntry("two", "2");
    }

    @Test
    void castStringMapReturnsEmptyMapForNullInput() {
        assertThat(Daytona.castStringMap(null)).isEmpty();
    }

    @Test
    void sandboxToMapHandlesNullSandboxAndMissingFields() {
        assertThat(Daytona.sandboxToMap(null)).isEmpty();

        io.daytona.api.client.model.Sandbox sandbox = TestSupport.mainSandbox("sb-utility", SandboxState.STARTED);
        sandbox.setCpu(null);
        sandbox.setGpu(null);
        sandbox.setMemory(null);
        sandbox.setDisk(null);
        sandbox.setState(null);

        assertThat(Daytona.sandboxToMap(sandbox))
                .containsEntry("id", "sb-utility")
                .containsEntry("state", null)
                .containsEntry("cpu", 0)
                .containsEntry("gpu", 0)
                .containsEntry("memory", 0)
                .containsEntry("disk", 0);
    }

    @Test
    void shutdownHttpClientAcceptsClientWithNoCache() {
        okhttp3.OkHttpClient client = new okhttp3.OkHttpClient();

        Daytona.shutdownHttpClient(client);

        assertThat(client.dispatcher().executorService().isShutdown()).isTrue();
    }
}
