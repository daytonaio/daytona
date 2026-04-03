// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.api.SnapshotsApi;
import io.daytona.api.client.model.CreateBuildInfo;
import io.daytona.api.client.model.CreateSnapshot;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.model.PaginatedSnapshots;
import io.daytona.sdk.model.Snapshot;
import okhttp3.OkHttpClient;

import java.math.BigDecimal;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.function.Consumer;

/**
 * Service for managing Daytona Snapshots.
 *
 * <p>Provides operations to create, list, retrieve, and delete snapshots.
 */
public class SnapshotService {
    private final SnapshotsApi snapshotsApi;
    private final OkHttpClient httpClient;
    private final String apiKey;

    SnapshotService(SnapshotsApi snapshotsApi, OkHttpClient httpClient, String apiKey) {
        this.snapshotsApi = snapshotsApi;
        this.httpClient = httpClient;
        this.apiKey = apiKey;
    }

    /**
     * Creates a snapshot from an existing image reference.
     *
     * @param name snapshot name
     * @param imageName source image name or tag
     * @return created {@link Snapshot}
     * @throws io.daytona.sdk.exception.DaytonaException if the API request fails
     */
    public Snapshot create(String name, String imageName) {
        io.daytona.api.client.model.SnapshotDto snapshotDto = ExceptionMapper.callMain(
                () -> snapshotsApi.createSnapshot(new CreateSnapshot().name(name).imageName(imageName), null)
        );
        return toSnapshot(snapshotDto);
    }

    /**
     * Creates a snapshot from a declarative {@link Image} with optional build log streaming.
     *
     * @param name snapshot name
     * @param image declarative image definition
     * @param onLogs callback for build log lines; {@code null} to skip streaming
     * @return created {@link Snapshot} in active or error state
     * @throws DaytonaException if the API request fails or the build fails
     */
    public Snapshot create(String name, Image image, Consumer<String> onLogs) {
        return create(name, image, null, onLogs);
    }

    /**
     * Creates a snapshot from a declarative {@link Image} with resources and optional build log streaming.
     *
     * @param name snapshot name
     * @param image declarative image definition
     * @param resources CPU/GPU/memory/disk resources; {@code null} for defaults
     * @param onLogs callback for build log lines; {@code null} to skip streaming
     * @return created {@link Snapshot} in active or error state
     * @throws DaytonaException if the API request fails or the build fails
     */
    public Snapshot create(String name, Image image, io.daytona.sdk.model.Resources resources, Consumer<String> onLogs) {
        CreateSnapshot req = new CreateSnapshot().name(name)
                .buildInfo(new CreateBuildInfo().dockerfileContent(image.getDockerfile()))
                .entrypoint(null);

        if (resources != null) {
            if (resources.getCpu() != null) req.setCpu(resources.getCpu());
            if (resources.getGpu() != null) req.setGpu(resources.getGpu());
            if (resources.getMemory() != null) req.setMemory(resources.getMemory());
            if (resources.getDisk() != null) req.setDisk(resources.getDisk());
        }

        final io.daytona.api.client.model.SnapshotDto[] ref = { ExceptionMapper.callMain(
                () -> snapshotsApi.createSnapshot(req, null)
        )};

        if (ref[0] == null) {
            throw new DaytonaException("Failed to create snapshot — no response from API");
        }

        List<String> terminalStates = Arrays.asList("active", "error", "build_failed");
        final String snapshotId = ref[0].getId();
        final String snapshotName = ref[0].getName();

        if (onLogs != null) {
            onLogs.accept("Creating snapshot " + snapshotName + " (" + stateString(ref[0]) + ")");
        }

        boolean logStreamStarted = false;
        while (!terminalStates.contains(stateString(ref[0]))) {
            if (onLogs != null && !logStreamStarted && !"pending".equals(stateString(ref[0]))) {
                logStreamStarted = true;
                io.daytona.api.client.model.Url logsUrl = ExceptionMapper.callMain(
                        () -> snapshotsApi.getSnapshotBuildLogsUrl(snapshotId, null));
                new BuildLogStreamer(httpClient, apiKey).streamLogs(
                        logsUrl.getUrl(), onLogs,
                        () -> terminalStates.contains(stateString(ref[0])));
            }
            try { Thread.sleep(1000); } catch (InterruptedException e) { Thread.currentThread().interrupt(); break; }
            ref[0] = ExceptionMapper.callMain(() -> snapshotsApi.getSnapshot(snapshotName, null));
        }

        if (onLogs != null && "active".equals(stateString(ref[0]))) {
            onLogs.accept("Created snapshot " + snapshotName + " (" + stateString(ref[0]) + ")");
        }

        if ("error".equals(stateString(ref[0])) || "build_failed".equals(stateString(ref[0]))) {
            throw new DaytonaException("Snapshot build failed: " + snapshotName + " (" + stateString(ref[0]) + ")");
        }

        return toSnapshot(ref[0]);
    }

    /**
     * Lists snapshots with pagination.
     *
     * @param page page number starting from 1; defaults to 1 when {@code null}
     * @param limit maximum number of items per page; defaults to 10 when {@code null}
     * @return paginated snapshot result
     * @throws io.daytona.sdk.exception.DaytonaException if the API request fails
     */
    public PaginatedSnapshots list(Integer page, Integer limit) {
        int p = page == null ? 1 : page;
        int l = limit == null ? 10 : limit;
        io.daytona.api.client.model.PaginatedSnapshots result = ExceptionMapper.callMain(
                () -> snapshotsApi.getAllSnapshots(null, BigDecimal.valueOf(p), BigDecimal.valueOf(l), null, null, null)
        );

        PaginatedSnapshots output = new PaginatedSnapshots();
        List<Snapshot> items = new ArrayList<Snapshot>();
        if (result != null && result.getItems() != null) {
            for (io.daytona.api.client.model.SnapshotDto snapshot : result.getItems()) {
                items.add(toSnapshot(snapshot));
            }
        }
        output.setItems(items);
        output.setTotal(result != null && result.getTotal() != null ? result.getTotal().intValue() : 0);
        output.setPage(result != null && result.getPage() != null ? result.getPage().intValue() : 0);
        output.setTotalPages(result != null && result.getTotalPages() != null ? result.getTotalPages().intValue() : 0);
        return output;
    }

    /**
     * Retrieves a snapshot by name or ID.
     *
     * @param nameOrId snapshot name or identifier
     * @return matching {@link Snapshot}
     * @throws io.daytona.sdk.exception.DaytonaException if no snapshot is found or request fails
     */
    public Snapshot get(String nameOrId) {
        io.daytona.api.client.model.SnapshotDto snapshotDto = ExceptionMapper.callMain(() -> snapshotsApi.getSnapshot(nameOrId, null));
        return toSnapshot(snapshotDto);
    }

    /**
     * Deletes a snapshot by ID.
     *
     * @param id snapshot identifier
     * @throws io.daytona.sdk.exception.DaytonaException if deletion fails
     */
    public void delete(String id) {
        ExceptionMapper.runMain(() -> snapshotsApi.removeSnapshot(id, null));
    }

    private String stateString(io.daytona.api.client.model.SnapshotDto dto) {
        return dto.getState() == null ? "" : dto.getState().getValue();
    }

    private Snapshot toSnapshot(io.daytona.api.client.model.SnapshotDto source) {
        Snapshot snapshot = new Snapshot();
        if (source != null) {
            snapshot.setId(source.getId());
            snapshot.setName(source.getName());
            snapshot.setImageName(source.getImageName());
            snapshot.setState(source.getState() == null ? null : source.getState().getValue());
        }
        return snapshot;
    }
}
