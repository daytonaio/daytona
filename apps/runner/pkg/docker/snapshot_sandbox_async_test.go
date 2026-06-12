// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const testImageHash = "a7be6198544f09a75b26e6376459b47c5b9972e7351d440e092c4faa9ea064ff"

// fakeAPIClient implements just the docker API surface CreateSnapshotFromSandbox
// touches. The embedded client.APIClient panics on anything else.
type fakeAPIClient struct {
	client.APIClient

	inspectErr error
	commitErr  error

	// When pushRelease is non-nil, ImagePush blocks until it is closed or the
	// push context is canceled. pushStarted receives one value per push begin;
	// pushCanceled receives the context error when a blocked push is canceled.
	pushRelease  chan struct{}
	pushStarted  chan struct{}
	pushCanceled chan error

	// When commitRelease is non-nil, ContainerCommit blocks until it is closed
	// or the commit context is canceled. commitStarted receives one value per
	// commit begin.
	commitRelease chan struct{}
	commitStarted chan struct{}

	// mu guards the container state reported by ContainerInspect (mutated by
	// ContainerPause/ContainerUnpause) and the pause bookkeeping below.
	mu                   sync.Mutex
	running              bool
	paused               bool
	pauseCalls           int
	unpauseCalls         int
	commitsInFlight      int
	unpausedDuringCommit bool
}

func (f *fakeAPIClient) ContainerInspect(_ context.Context, containerID string) (container.InspectResponse, error) {
	if f.inspectErr != nil {
		return container.InspectResponse{}, f.inspectErr
	}
	f.mu.Lock()
	state := container.State{Running: f.running, Paused: f.paused}
	f.mu.Unlock()
	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:    containerID,
			State: &state,
		},
	}, nil
}

func (f *fakeAPIClient) ContainerPause(_ context.Context, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pauseCalls++
	f.paused = true
	return nil
}

func (f *fakeAPIClient) ContainerUnpause(_ context.Context, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unpauseCalls++
	if f.commitsInFlight > 0 {
		f.unpausedDuringCommit = true
	}
	f.paused = false
	return nil
}

func (f *fakeAPIClient) ContainerCommit(ctx context.Context, _ string, _ container.CommitOptions) (container.CommitResponse, error) {
	f.mu.Lock()
	f.commitsInFlight++
	f.mu.Unlock()
	defer func() {
		f.mu.Lock()
		f.commitsInFlight--
		f.mu.Unlock()
	}()

	if f.commitStarted != nil {
		f.commitStarted <- struct{}{}
	}
	if f.commitRelease != nil {
		select {
		case <-f.commitRelease:
		case <-ctx.Done():
			return container.CommitResponse{}, ctx.Err()
		}
	}
	if f.commitErr != nil {
		return container.CommitResponse{}, f.commitErr
	}
	return container.CommitResponse{ID: "sha256:" + testImageHash}, nil
}

func (f *fakeAPIClient) ImageInspect(_ context.Context, _ string, _ ...client.ImageInspectOption) (image.InspectResponse, error) {
	return image.InspectResponse{
		ID:   "sha256:" + testImageHash,
		Size: 2 * 1024 * 1024 * 1024,
		Config: &dockerspec.DockerOCIImageConfig{
			ImageConfig: ocispec.ImageConfig{
				Entrypoint: []string{"/entry"},
				Cmd:        []string{"serve"},
			},
		},
	}, nil
}

func (f *fakeAPIClient) ImageTag(_ context.Context, _, _ string) error {
	return nil
}

func (f *fakeAPIClient) ImagePush(ctx context.Context, _ string, _ image.PushOptions) (io.ReadCloser, error) {
	if f.pushStarted != nil {
		f.pushStarted <- struct{}{}
	}
	if f.pushRelease != nil {
		select {
		case <-f.pushRelease:
		case <-ctx.Done():
			if f.pushCanceled != nil {
				f.pushCanceled <- ctx.Err()
			}
			return nil, ctx.Err()
		}
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (f *fakeAPIClient) ImageRemove(_ context.Context, _ string, _ image.RemoveOptions) ([]image.DeleteResponse, error) {
	return nil, nil
}

func testRegistry() *dto.RegistryDTO {
	project := "proj"
	return &dto.RegistryDTO{
		Url:     "registry.example.com",
		Project: &project,
	}
}

func newTestDockerClient(t *testing.T, fake *fakeAPIClient) *DockerClient {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return &DockerClient{
		apiClient:                    fake,
		snapshotFromSandboxInfoCache: cache.NewSnapshotFromSandboxInfoCache(ctx, time.Hour),
		logger:                       slog.New(slog.NewTextHandler(io.Discard, nil)),
		backupTimeoutMin:             1,
	}
}

func getCaptureInfo(t *testing.T, d *DockerClient, sandboxID string) *models.SnapshotFromSandboxInfo {
	t.Helper()
	entry, err := d.snapshotFromSandboxInfoCache.Get(context.Background(), sandboxID)
	if err != nil {
		return nil
	}
	return entry
}

func waitForCaptureInfo(t *testing.T, d *DockerClient, sandboxID string, pred func(models.SnapshotFromSandboxInfo) bool) models.SnapshotFromSandboxInfo {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if entry := getCaptureInfo(t, d, sandboxID); entry != nil && pred(*entry) {
			return *entry
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for capture info for sandbox %s", sandboxID)
	return models.SnapshotFromSandboxInfo{}
}

func TestCreateSnapshotFromSandboxAsyncHappyPath(t *testing.T) {
	fake := &fakeAPIClient{
		pushRelease: make(chan struct{}),
		pushStarted: make(chan struct{}, 1),
	}
	d := newTestDockerClient(t, fake)

	request := dto.CreateSnapshotFromSandboxRequestDTO{
		Name:     "my-snap",
		Registry: testRegistry(),
		Async:    true,
	}

	if err := d.CreateSnapshotFromSandboxAsync(context.Background(), "sandbox-1", request); err != nil {
		t.Fatalf("CreateSnapshotFromSandboxAsync: %v", err)
	}

	// IN_PROGRESS must be visible synchronously, before the capture finishes.
	entry := getCaptureInfo(t, d, "sandbox-1")
	if entry == nil {
		t.Fatal("expected an IN_PROGRESS cache entry immediately after initiation")
	}
	if entry.State != enums.SnapshotFromSandboxStateInProgress {
		t.Fatalf("expected IN_PROGRESS, got %s", entry.State)
	}
	if entry.Name != "my-snap" {
		t.Fatalf("expected entry name my-snap, got %q", entry.Name)
	}

	<-fake.pushStarted
	close(fake.pushRelease)

	done := waitForCaptureInfo(t, d, "sandbox-1", func(e models.SnapshotFromSandboxInfo) bool {
		return e.State == enums.SnapshotFromSandboxStateCompleted
	})

	if done.Name != "my-snap" {
		t.Fatalf("expected completed entry name my-snap, got %q", done.Name)
	}
	if done.Info == nil {
		t.Fatal("expected completed entry to carry snapshot info")
	}
	wantRef := "registry.example.com/proj/daytona-" + testImageHash + ":daytona"
	if done.Info.Name != wantRef {
		t.Fatalf("expected canonical ref %q, got %q", wantRef, done.Info.Name)
	}
	if done.Info.Hash != testImageHash {
		t.Fatalf("expected hash %q, got %q", testImageHash, done.Info.Hash)
	}
	if done.Info.SizeGB != 2.0 {
		t.Fatalf("expected sizeGB 2.0, got %v", done.Info.SizeGB)
	}
	if done.Error != nil {
		t.Fatalf("expected no error on completed capture, got %v", done.Error)
	}
}

func TestCreateSnapshotFromSandboxAsyncCommitFailure(t *testing.T) {
	fake := &fakeAPIClient{commitErr: errors.New("commit exploded")}
	d := newTestDockerClient(t, fake)

	request := dto.CreateSnapshotFromSandboxRequestDTO{
		Name:     "my-snap",
		Registry: testRegistry(),
		Async:    true,
	}

	if err := d.CreateSnapshotFromSandboxAsync(context.Background(), "sandbox-2", request); err != nil {
		t.Fatalf("CreateSnapshotFromSandboxAsync: %v", err)
	}

	failed := waitForCaptureInfo(t, d, "sandbox-2", func(e models.SnapshotFromSandboxInfo) bool {
		return e.State == enums.SnapshotFromSandboxStateFailed
	})

	if failed.Error == nil {
		t.Fatal("expected failed capture to carry an error")
	}
	if !strings.Contains(failed.Error.Error(), "commit exploded") {
		t.Fatalf("expected underlying commit error to be preserved, got %q", failed.Error.Error())
	}
	if failed.Info != nil {
		t.Fatalf("expected no snapshot info on failure, got %+v", failed.Info)
	}
}

func TestCreateSnapshotFromSandboxAsyncMissingContainer(t *testing.T) {
	fake := &fakeAPIClient{inspectErr: errors.New("No such container: sandbox-3")}
	d := newTestDockerClient(t, fake)

	request := dto.CreateSnapshotFromSandboxRequestDTO{
		Name:     "my-snap",
		Registry: testRegistry(),
		Async:    true,
	}

	err := d.CreateSnapshotFromSandboxAsync(context.Background(), "sandbox-3", request)
	if err == nil {
		t.Fatal("expected a synchronous error for a missing container")
	}
	if !strings.Contains(err.Error(), "No such container") {
		t.Fatalf("expected inspect error to surface, got %q", err.Error())
	}

	if entry := getCaptureInfo(t, d, "sandbox-3"); entry != nil {
		t.Fatalf("expected no cache entry after validation failure, got %+v", entry)
	}
}

func TestCreateSnapshotFromSandboxAsyncSupersedesPriorCapture(t *testing.T) {
	fake := &fakeAPIClient{
		pushRelease:  make(chan struct{}),
		pushStarted:  make(chan struct{}, 2),
		pushCanceled: make(chan error, 2),
	}
	d := newTestDockerClient(t, fake)

	first := dto.CreateSnapshotFromSandboxRequestDTO{
		Name:     "snap-a",
		Registry: testRegistry(),
		Async:    true,
	}
	if err := d.CreateSnapshotFromSandboxAsync(context.Background(), "sandbox-4", first); err != nil {
		t.Fatalf("first CreateSnapshotFromSandboxAsync: %v", err)
	}

	// Wait until the first capture is blocked mid-push.
	<-fake.pushStarted

	second := dto.CreateSnapshotFromSandboxRequestDTO{
		Name:     "snap-b",
		Registry: testRegistry(),
		Async:    true,
	}
	if err := d.CreateSnapshotFromSandboxAsync(context.Background(), "sandbox-4", second); err != nil {
		t.Fatalf("second CreateSnapshotFromSandboxAsync: %v", err)
	}

	// The second capture must cancel the first one's in-flight push.
	select {
	case pushErr := <-fake.pushCanceled:
		if !errors.Is(pushErr, context.Canceled) {
			t.Fatalf("expected first push to observe context.Canceled, got %v", pushErr)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("first capture was not canceled by the superseding capture")
	}

	// The superseded capture must terminate without writing anything to the
	// cache: the successor's IN_PROGRESS entry has to survive. Wait until the
	// second capture is blocked mid-push (the first capture's unwind is pure
	// compute, long finished by then), then hold the assertion for a window.
	<-fake.pushStarted
	deadline := time.Now().Add(150 * time.Millisecond)
	for time.Now().Before(deadline) {
		entry := getCaptureInfo(t, d, "sandbox-4")
		if entry == nil {
			t.Fatal("expected the successor's cache entry to survive the superseded capture")
		}
		if entry.Name != "snap-b" || entry.State != enums.SnapshotFromSandboxStateInProgress {
			t.Fatalf("superseded capture clobbered the successor's cache entry: %+v", entry)
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Release the second capture's push and let it complete.
	close(fake.pushRelease)

	done := waitForCaptureInfo(t, d, "sandbox-4", func(e models.SnapshotFromSandboxInfo) bool {
		return e.Name == "snap-b" && e.State == enums.SnapshotFromSandboxStateCompleted
	})
	if done.Info == nil {
		t.Fatal("expected superseding capture to complete with snapshot info")
	}
}

// TestSnapshotFromSandboxPausesAndUnpausesOwnedCapture covers the normal
// owner lifecycle: a running container is paused for commit, unpaused exactly
// once afterwards (never while a commit is in flight), and the pause record
// is released.
func TestSnapshotFromSandboxPausesAndUnpausesOwnedCapture(t *testing.T) {
	fake := &fakeAPIClient{running: true}
	d := newTestDockerClient(t, fake)

	owner, ownerCancel := context.WithCancel(context.Background())
	defer ownerCancel()

	if _, err := d.createSnapshotFromSandbox(context.Background(), "sandbox-5", testRegistry(), owner); err != nil {
		t.Fatalf("createSnapshotFromSandbox: %v", err)
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.pauseCalls != 1 {
		t.Fatalf("expected exactly one pause, got %d", fake.pauseCalls)
	}
	if fake.unpauseCalls != 1 {
		t.Fatalf("expected exactly one unpause, got %d", fake.unpauseCalls)
	}
	if fake.unpausedDuringCommit {
		t.Fatal("container was unpaused while a commit was in flight")
	}
	if fake.paused {
		t.Fatal("expected the container to end up unpaused")
	}
	if _, ok := capture_pause_map.Get("sandbox-5"); ok {
		t.Fatal("expected the pause record to be released after the capture")
	}
}

// TestSnapshotFromSandboxAdoptsPauseOfSupersededCapture: a capture that finds
// the container paused by the capture it superseded adopts the unpause — it
// commits without pausing again and unpauses exactly once, while the
// superseded owner can no longer claim the pause.
func TestSnapshotFromSandboxAdoptsPauseOfSupersededCapture(t *testing.T) {
	fake := &fakeAPIClient{running: true, paused: true}
	d := newTestDockerClient(t, fake)

	predecessor, predecessorCancel := context.WithCancel(context.Background())
	defer predecessorCancel()
	capture_pause_map.Set("sandbox-6", predecessor)

	successor, successorCancel := context.WithCancel(context.Background())
	defer successorCancel()

	if _, err := d.createSnapshotFromSandbox(context.Background(), "sandbox-6", testRegistry(), successor); err != nil {
		t.Fatalf("createSnapshotFromSandbox: %v", err)
	}

	if claimCapturePause("sandbox-6", predecessor) {
		t.Fatal("expected the superseded capture to have lost its pause claim")
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.pauseCalls != 0 {
		t.Fatalf("expected the adopting capture not to pause again, got %d pauses", fake.pauseCalls)
	}
	if fake.unpauseCalls != 1 {
		t.Fatalf("expected exactly one unpause by the adopting capture, got %d", fake.unpauseCalls)
	}
	if fake.unpausedDuringCommit {
		t.Fatal("container was unpaused while a commit was in flight")
	}
	if fake.paused {
		t.Fatal("expected the container to end up unpaused")
	}
}

// TestSnapshotFromSandboxSupersededCaptureSkipsUnpause: once a successor has
// adopted the pause record, the superseded capture's cleanup must not unpause
// the container underneath it.
func TestSnapshotFromSandboxSupersededCaptureSkipsUnpause(t *testing.T) {
	fake := &fakeAPIClient{
		running:       true,
		commitRelease: make(chan struct{}),
		commitStarted: make(chan struct{}, 4),
	}
	d := newTestDockerClient(t, fake)

	predecessor, predecessorCancel := context.WithCancel(context.Background())
	defer predecessorCancel()

	captureErr := make(chan error, 1)
	go func() {
		_, err := d.createSnapshotFromSandbox(predecessor, "sandbox-7", testRegistry(), predecessor)
		captureErr <- err
	}()

	// The predecessor has paused the container and is now blocked mid-commit.
	<-fake.commitStarted

	// Adopt the pause record exactly the way a superseding capture does, then
	// cancel the predecessor.
	if !capture_pause_map.RemoveCb("sandbox-7", func(_ string, _ context.Context, exists bool) bool { return exists }) {
		t.Fatal("expected the predecessor to hold the pause record while committing")
	}
	successor, successorCancel := context.WithCancel(context.Background())
	defer successorCancel()
	capture_pause_map.Set("sandbox-7", successor)
	predecessorCancel()

	select {
	case err := <-captureErr:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected the canceled capture to fail with context.Canceled, got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("canceled capture did not return")
	}

	if !claimCapturePause("sandbox-7", successor) {
		t.Fatal("expected the successor to still own the pause record")
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.unpauseCalls != 0 {
		t.Fatalf("expected the superseded capture to skip its unpause, got %d", fake.unpauseCalls)
	}
	if !fake.paused {
		t.Fatal("expected the container to stay paused for the successor")
	}
}
