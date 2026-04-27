// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaBadRequestException;
import io.daytona.sdk.exception.DaytonaForbiddenException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaRateLimitException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.toolbox.client.api.LspApi;
import io.daytona.toolbox.client.model.CompletionList;
import io.daytona.toolbox.client.model.LspSymbol;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Arrays;
import java.util.Collections;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class LspServerTest {

    @Mock
    private LspApi lspApi;

    private LspServer lspServer;

    @BeforeEach
    void setUp() {
        lspServer = new LspServer(lspApi);
    }

    @Test
    void startStopAndDocumentLifecycleBuildRequests() {
        lspServer.start("python", "/project");
        lspServer.stop("python", "/project");
        lspServer.didOpen("python", "/project", "file:///main.py");
        lspServer.didClose("python", "/project", "file:///main.py");

        verify(lspApi).start(argThat(request -> "python".equals(request.getLanguageId()) && "/project".equals(request.getPathToProject())));
        verify(lspApi).stop(argThat(request -> "python".equals(request.getLanguageId()) && "/project".equals(request.getPathToProject())));
        verify(lspApi).didOpen(argThat(request -> "file:///main.py".equals(request.getUri())));
        verify(lspApi).didClose(argThat(request -> "file:///main.py".equals(request.getUri())));
    }

    @Test
    void completionsBuildPositionRequest() {
        CompletionList completionList = new CompletionList();
        when(lspApi.completions(any())).thenReturn(completionList);

        CompletionList result = lspServer.completions("typescript", "/project", "file:///a.ts", 2, 7);

        assertThat(result).isSameAs(completionList);
        ArgumentCaptor<io.daytona.toolbox.client.model.LspCompletionParams> captor = ArgumentCaptor.forClass(io.daytona.toolbox.client.model.LspCompletionParams.class);
        verify(lspApi).completions(captor.capture());
        assertThat(captor.getValue().getPosition().getLine()).isEqualTo(2);
        assertThat(captor.getValue().getPosition().getCharacter()).isEqualTo(7);
    }

    @Test
    void symbolQueriesDelegate() {
        LspSymbol symbol = new LspSymbol();
        symbol.setName("main");
        when(lspApi.documentSymbols("python", "/project", "file:///main.py")).thenReturn(Collections.singletonList(symbol));
        when(lspApi.workspaceSymbols("query", "python", "/project")).thenReturn(Arrays.asList(symbol));

        assertThat(lspServer.documentSymbols("python", "/project", "file:///main.py")).singleElement().extracting(LspSymbol::getName).isEqualTo("main");
        assertThat(lspServer.workspaceSymbols("query", "python", "/project")).hasSize(1);
        assertThat(LspServer.LspLanguageId.TYPESCRIPT.getValue()).isEqualTo("typescript");
    }

    @Test
    void completionsAndSymbolsAllowNullResults() {
        when(lspApi.completions(any())).thenReturn(null);
        when(lspApi.documentSymbols("python", "/project", "file:///main.py")).thenReturn(null);
        when(lspApi.workspaceSymbols("query", "python", "/project")).thenReturn(null);

        assertThat(lspServer.completions("python", "/project", "file:///main.py", 0, 0)).isNull();
        assertThat(lspServer.documentSymbols("python", "/project", "file:///main.py")).isNull();
        assertThat(lspServer.workspaceSymbols("query", "python", "/project")).isNull();
        assertThat(LspServer.LspLanguageId.PYTHON.getValue()).isEqualTo("python");
        assertThat(LspServer.LspLanguageId.JAVASCRIPT.getValue()).isEqualTo("javascript");
    }

    @ParameterizedTest
    @MethodSource("mappedToolboxExceptions")
    void startMapsToolboxErrors(int status, Class<? extends RuntimeException> type) {
        org.mockito.Mockito.doThrow(new io.daytona.toolbox.client.ApiException(status, "boom", null, "{\"message\":\"mapped\"}"))
                .when(lspApi).start(any());

        assertThatThrownBy(() -> lspServer.start("python", "/project"))
                .isInstanceOf(type)
                .hasMessage("mapped");
    }

    private static Stream<Arguments> mappedToolboxExceptions() {
        return Stream.of(
                Arguments.of(400, DaytonaBadRequestException.class),
                Arguments.of(403, DaytonaForbiddenException.class),
                Arguments.of(404, DaytonaNotFoundException.class),
                Arguments.of(429, DaytonaRateLimitException.class),
                Arguments.of(500, DaytonaServerException.class)
        );
    }

    private static <T> T argThat(org.mockito.ArgumentMatcher<T> matcher) {
        return org.mockito.ArgumentMatchers.argThat(matcher);
    }
}
