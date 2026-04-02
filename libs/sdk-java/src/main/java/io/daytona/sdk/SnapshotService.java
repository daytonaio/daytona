// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.api.SnapshotsApi;
import io.daytona.api.client.model.CreateSnapshot;
import io.daytona.sdk.model.PaginatedSnapshots;
import io.daytona.sdk.model.Snapshot;

import java.math.BigDecimal;
import java.util.ArrayList;
import java.util.List;

/**
 * Service for managing Daytona Snapshots.
 *
 * <p>Provides operations to create, list, retrieve, and delete snapshots.
 */
public class SnapshotService {
    private final SnapshotsApi snapshotsApi;

    SnapshotService(SnapshotsApi snapshotsApi) {
        this.snapshotsApi = snapshotsApi;
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
