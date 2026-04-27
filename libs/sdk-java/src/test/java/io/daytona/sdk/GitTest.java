// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaBadRequestException;
import io.daytona.sdk.exception.DaytonaConflictException;
import io.daytona.sdk.exception.DaytonaForbiddenException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaRateLimitException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.sdk.model.GitCommitResponse;
import io.daytona.sdk.model.GitStatus;
import io.daytona.toolbox.client.api.GitApi;
import io.daytona.toolbox.client.model.FileStatus;
import io.daytona.toolbox.client.model.ListBranchResponse;
import io.daytona.toolbox.client.model.Status;
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
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class GitTest {

    @Mock
    private GitApi gitApi;

    private Git git;

    @BeforeEach
    void setUp() {
        git = new Git(gitApi);
    }

    @Test
    void cloneBuildsMinimalRequest() {
        git.clone("https://example.com/repo.git", "/workspace/repo");

        ArgumentCaptor<io.daytona.toolbox.client.model.GitCloneRequest> captor = ArgumentCaptor.forClass(io.daytona.toolbox.client.model.GitCloneRequest.class);
        verify(gitApi).cloneRepository(captor.capture());
        assertThat(captor.getValue().getUrl()).isEqualTo("https://example.com/repo.git");
        assertThat(captor.getValue().getPath()).isEqualTo("/workspace/repo");
        assertThat(captor.getValue().getBranch()).isNull();
    }

    @Test
    void cloneBuildsFullRequest() {
        git.clone("https://example.com/repo.git", "/workspace/repo", "main", "abc123", "user", "pass");

        ArgumentCaptor<io.daytona.toolbox.client.model.GitCloneRequest> captor = ArgumentCaptor.forClass(io.daytona.toolbox.client.model.GitCloneRequest.class);
        verify(gitApi).cloneRepository(captor.capture());
        assertThat(captor.getValue().getBranch()).isEqualTo("main");
        assertThat(captor.getValue().getCommitId()).isEqualTo("abc123");
        assertThat(captor.getValue().getUsername()).isEqualTo("user");
        assertThat(captor.getValue().getPassword()).isEqualTo("pass");
    }

    @Test
    void branchesReturnsResponseData() {
        when(gitApi.listBranches("/repo")).thenReturn(new ListBranchResponse().branches(Arrays.asList("main", "feature")));

        assertThat(git.branches("/repo")).containsEntry("branches", Arrays.asList("main", "feature"));
    }

    @Test
    void branchesReturnsEmptyListWhenApiReturnsNull() {
        when(gitApi.listBranches("/repo")).thenReturn(null);

        assertThat(git.branches("/repo")).containsEntry("branches", Collections.emptyList());
    }

    @Test
    void addBuildsRequest() {
        git.add("/repo", Arrays.asList("A.java", "B.java"));

        verify(gitApi).addFiles(argThat(request -> "/repo".equals(request.getPath()) && request.getFiles().equals(Arrays.asList("A.java", "B.java"))));
    }

    @Test
    void commitMapsHash() {
        io.daytona.toolbox.client.model.GitCommitResponse response = new io.daytona.toolbox.client.model.GitCommitResponse();
        response.setHash("abc123");
        when(gitApi.commitChanges(any())).thenReturn(response);

        GitCommitResponse commitResponse = git.commit("/repo", "msg", "Author", "a@example.com");

        assertThat(commitResponse.getHash()).isEqualTo("abc123");
    }

    @Test
    void commitReturnsEmptyResponseWhenApiReturnsNull() {
        when(gitApi.commitChanges(any())).thenReturn(null);

        GitCommitResponse commitResponse = git.commit("/repo", "msg", "Author", "a@example.com");

        assertThat(commitResponse.getHash()).isNull();
    }

    @Test
    void statusMapsNestedFileStatuses() {
        io.daytona.toolbox.client.model.GitStatus response = new io.daytona.toolbox.client.model.GitStatus();
        response.setCurrentBranch("main");
        response.setAhead(2);
        response.setBehind(1);
        response.setBranchPublished(true);
        FileStatus fileStatus = new FileStatus();
        fileStatus.setName("README.md");
        fileStatus.setStaging(Status.Modified);
        fileStatus.setWorktree(Status.Untracked);
        response.setFileStatus(Collections.singletonList(fileStatus));
        when(gitApi.getStatus("/repo")).thenReturn(response);

        GitStatus status = git.status("/repo");

        assertThat(status.getCurrentBranch()).isEqualTo("main");
        assertThat(status.getAhead()).isEqualTo(2);
        assertThat(status.getBehind()).isEqualTo(1);
        assertThat(status.isBranchPublished()).isTrue();
        assertThat(status.getFileStatus()).singleElement().satisfies(item -> {
            assertThat(item.getPath()).isEqualTo("README.md");
            assertThat(item.getStatus()).isEqualTo("Modified/Untracked");
        });
    }

    @Test
    void statusUsesDefaultsForNullResponse() {
        when(gitApi.getStatus("/repo")).thenReturn(null);

        GitStatus status = git.status("/repo");

        assertThat(status.getCurrentBranch()).isNull();
        assertThat(status.getAhead()).isZero();
        assertThat(status.getBehind()).isZero();
        assertThat(status.getFileStatus()).isEmpty();
    }

    @Test
    void pushAndPullDelegate() {
        git.push("/repo");
        git.pull("/repo");

        verify(gitApi).pushChanges(argThat(request -> "/repo".equals(request.getPath())));
        verify(gitApi).pullChanges(argThat(request -> "/repo".equals(request.getPath())));
    }

    @ParameterizedTest
    @MethodSource("mappedToolboxExceptions")
    void cloneMapsToolboxErrors(int status, Class<? extends RuntimeException> type) {
        org.mockito.Mockito.doThrow(new io.daytona.toolbox.client.ApiException(status, "boom", null, "{\"message\":\"mapped\"}"))
                .when(gitApi).cloneRepository(any());

        assertThatThrownBy(() -> git.clone("https://example.com/repo.git", "/repo"))
                .isInstanceOf(type)
                .hasMessage("mapped");
    }

    private static Stream<Arguments> mappedToolboxExceptions() {
        return Stream.of(
                Arguments.of(400, DaytonaBadRequestException.class),
                Arguments.of(403, DaytonaForbiddenException.class),
                Arguments.of(404, DaytonaNotFoundException.class),
                Arguments.of(409, DaytonaConflictException.class),
                Arguments.of(429, DaytonaRateLimitException.class),
                Arguments.of(500, DaytonaServerException.class)
        );
    }

    private static <T> T argThat(org.mockito.ArgumentMatcher<T> matcher) {
        return org.mockito.ArgumentMatchers.argThat(matcher);
    }
}
