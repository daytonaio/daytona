# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import json
import os
import threading
import time
import uuid

import pytest

from daytona import (
    CreateSandboxFromImageParams,
    CreateSandboxFromSnapshotParams,
    Daytona,
    DownloadProgress,
    FileDownloadRequest,
    FileUpload,
    Image,
    PtySize,
    Sandbox,
    SessionExecuteRequest,
    UploadProgress,
)
from daytona.common.errors import DaytonaError

if not os.getenv("DAYTONA_API_KEY"):
    raise RuntimeError("DAYTONA_API_KEY environment variable is required for E2E tests")

pytestmark = [pytest.mark.e2e]


lsp_server = None
pty_session_id = ""


def _error_message(exc: Exception) -> str:
    return str(exc).lower()


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


@pytest.fixture(scope="module")
def daytona_client() -> Daytona:
    return Daytona()


@pytest.fixture(scope="module")
def sandbox(daytona_client: Daytona):
    params = CreateSandboxFromSnapshotParams(language="python")
    sb = daytona_client.create(params, timeout=120)
    yield sb
    try:
        daytona_client.delete(sb)
    except Exception:
        pass


# ===========================================================================
# Sandbox Lifecycle
# ===========================================================================


def test_sandbox_has_valid_id(sandbox):
    assert sandbox.id, "Sandbox should have a non-empty ID"


def test_sandbox_has_valid_name(sandbox):
    assert sandbox.name, "Sandbox should have a non-empty name"


def test_sandbox_has_organization_id(sandbox):
    assert sandbox.organization_id, "Sandbox should have an organization_id"


def test_sandbox_state_is_started(sandbox):
    state = str(getattr(sandbox.state, "value", sandbox.state)).lower()
    assert state == "started", f"Expected 'started', got {state!r}"


def test_sandbox_has_resource_properties(sandbox):
    assert sandbox.cpu > 0, f"Expected cpu > 0, got {sandbox.cpu}"
    assert sandbox.memory > 0, f"Expected memory > 0, got {sandbox.memory}"
    assert sandbox.disk > 0, f"Expected disk > 0, got {sandbox.disk}"


def test_sandbox_has_timestamps(sandbox):
    assert sandbox.created_at, "created_at should be set"
    assert sandbox.updated_at, "updated_at should be set"


def test_get_user_home_dir_returns_path(sandbox):
    home = sandbox.get_user_home_dir()
    assert isinstance(home, str) and home.startswith("/"), f"Expected absolute path, got {home!r}"


def test_get_work_dir_returns_path(sandbox):
    work = sandbox.get_work_dir()
    assert isinstance(work, str) and work.startswith("/"), f"Expected absolute path, got {work!r}"


def test_set_labels_returns_updated_labels(sandbox):
    labels = sandbox.set_labels({"e2e": "true", "suite": "python"})
    assert isinstance(labels, dict), "set_labels should return a dict"
    assert labels.get("e2e") == "true"
    assert labels.get("suite") == "python"


def test_set_autostop_interval_updates_value(sandbox):
    sandbox.set_autostop_interval(30)
    assert sandbox.auto_stop_interval == 30, f"Expected 30, got {sandbox.auto_stop_interval}"


def test_set_auto_archive_interval_updates_value(sandbox):
    sandbox.set_auto_archive_interval(60)
    assert sandbox.auto_archive_interval == 60, f"Expected 60, got {sandbox.auto_archive_interval}"


def test_set_auto_delete_interval_set_and_disable(sandbox):
    sandbox.set_auto_delete_interval(120)
    assert sandbox.auto_delete_interval == 120, f"Expected 120, got {sandbox.auto_delete_interval}"
    sandbox.set_auto_delete_interval(-1)
    assert sandbox.auto_delete_interval == -1, f"Expected -1 (disabled), got {sandbox.auto_delete_interval}"


def test_refresh_data_updates_sandbox(sandbox):
    old_id = sandbox.id
    sandbox.refresh_data()
    assert sandbox.id == old_id, "Sandbox ID should not change after refresh_data"
    assert sandbox.state is not None, "State should still be present after refresh"


# ===========================================================================
# Daytona Client Operations
# ===========================================================================


def test_list_sandboxes(daytona_client, sandbox):
    result = daytona_client.list()
    assert result.total > 0, f"Expected total > 0, got {result.total}"
    assert len(result.items) > 0, "Expected at least one sandbox in items"


def test_list_with_pagination(daytona_client, sandbox):
    result = daytona_client.list(page=1, limit=1)
    assert result.total >= 1, "Expected total >= 1"
    assert len(result.items) <= 1, f"Expected at most 1 item, got {len(result.items)}"
    assert result.page == 1, f"Expected page 1, got {result.page}"


def test_get_sandbox_by_id(daytona_client, sandbox):
    fetched = daytona_client.get(sandbox.id)
    assert fetched.id == sandbox.id, f"Expected id {sandbox.id}, got {fetched.id}"
    assert fetched.name == sandbox.name


# ===========================================================================
# File System Operations
# ===========================================================================

FS_TEST_DIR = "e2e-fs-test"


def test_create_folder(sandbox):
    try:
        sandbox.fs.delete_file(FS_TEST_DIR, recursive=True)
    except Exception:
        pass
    sandbox.fs.create_folder(FS_TEST_DIR, "755")
    listed = sandbox.fs.list_files(FS_TEST_DIR)
    assert isinstance(listed, list), "list_files should return a list"


def test_upload_file_bytes(sandbox):
    sandbox.fs.upload_file(b"hello e2e", f"{FS_TEST_DIR}/hello.txt")
    content = sandbox.fs.download_file(f"{FS_TEST_DIR}/hello.txt")
    assert content == b"hello e2e", f"Expected exact bytes, got {content!r}"


def test_upload_files_batch(sandbox):
    sandbox.fs.upload_files(
        [
            FileUpload(source=b"file one", destination=f"{FS_TEST_DIR}/batch1.txt"),
            FileUpload(source=b"file two", destination=f"{FS_TEST_DIR}/batch2.txt"),
        ]
    )
    c1 = sandbox.fs.download_file(f"{FS_TEST_DIR}/batch1.txt")
    c2 = sandbox.fs.download_file(f"{FS_TEST_DIR}/batch2.txt")
    assert c1 == b"file one"
    assert c2 == b"file two"


def test_list_files_contains_uploaded(sandbox):
    files = sandbox.fs.list_files(FS_TEST_DIR)
    names = [f.name for f in files]
    assert "hello.txt" in names, f"Expected hello.txt in {names}"
    assert "batch1.txt" in names, f"Expected batch1.txt in {names}"


def test_get_file_info_returns_size(sandbox):
    info = sandbox.fs.get_file_info(f"{FS_TEST_DIR}/hello.txt")
    assert info.name == "hello.txt"
    assert info.size is not None and info.size > 0, f"Expected size > 0, got {info.size}"
    assert info.is_dir is False


def test_download_file_returns_exact_content(sandbox):
    sandbox.fs.upload_file(b"exact content check", f"{FS_TEST_DIR}/exact.txt")
    content = sandbox.fs.download_file(f"{FS_TEST_DIR}/exact.txt")
    assert content == b"exact content check"


def test_download_files_batch(sandbox):
    results = sandbox.fs.download_files(
        [
            FileDownloadRequest(source=f"{FS_TEST_DIR}/batch1.txt"),
            FileDownloadRequest(source=f"{FS_TEST_DIR}/batch2.txt"),
        ]
    )
    assert len(results) == 2, f"Expected 2 results, got {len(results)}"
    assert results[0].result == b"file one"
    assert results[1].result == b"file two"


def test_download_file_stream_returns_exact_content(sandbox):
    expected = b"streamed content\nsecond line\n"
    path = f"{FS_TEST_DIR}/streamed-{uuid.uuid4().hex}.txt"
    sandbox.fs.upload_file(expected, path)

    content = b"".join(sandbox.fs.download_file_stream(path))

    assert content == expected


def test_download_file_stream_on_progress(sandbox: Sandbox):
    run_id = uuid.uuid4().hex
    expected = ("progress-check-" + run_id).encode("utf-8") * 512
    # Use a run-unique path to avoid interference when multiple E2E suites
    # run concurrently against the same sandbox (e.g. nx run-many in CI).
    path = f"{FS_TEST_DIR}/streamed-progress-{run_id}.txt"
    sandbox.fs.upload_file(expected, path)
    progress_updates: list[DownloadProgress] = []

    content = b"".join(sandbox.fs.download_file_stream(path, on_progress=progress_updates.append))

    assert content == expected
    assert progress_updates
    assert progress_updates[-1].bytes_received == len(expected)
    assert progress_updates[-1].total_bytes == len(expected)
    assert [progress.bytes_received for progress in progress_updates] == sorted(
        progress.bytes_received for progress in progress_updates
    )
    assert all(progress.total_bytes == len(expected) for progress in progress_updates)


def test_download_file_stream_nonexistent_file_raises(sandbox):
    with pytest.raises(DaytonaError):
        b"".join(sandbox.fs.download_file_stream(f"{FS_TEST_DIR}/stream-does-not-exist.txt"))


def test_upload_file_stream_with_progress(sandbox: Sandbox):
    import io as _io

    run_id = uuid.uuid4().hex
    payload = ("upload-stream-content-" + run_id).encode("utf-8") * 512
    path = f"{FS_TEST_DIR}/upload-stream-{run_id}.bin"
    progress: list[UploadProgress] = []

    sandbox.fs.upload_file_stream(
        _io.BytesIO(payload),
        path,
        on_progress=progress.append,
    )

    downloaded = b"".join(sandbox.fs.download_file_stream(path))
    assert downloaded == payload
    assert progress, "expected at least one progress event"
    assert progress[-1].bytes_sent == len(payload)
    assert [p.bytes_sent for p in progress] == sorted(p.bytes_sent for p in progress)


def test_find_files_finds_text_content(sandbox):
    matches = sandbox.fs.find_files(FS_TEST_DIR, "hello")
    assert len(matches) > 0, "Expected find_files to return at least one match for 'hello'"


def test_search_files_by_glob(sandbox):
    result = sandbox.fs.search_files(FS_TEST_DIR, "*.txt")
    assert len(result.files) >= 1, f"Expected at least 1 .txt file, got {result.files}"
    assert any("hello.txt" in f for f in result.files), f"Expected hello.txt in {result.files}"


def test_replace_in_files_modifies_content(sandbox):
    sandbox.fs.upload_file(b"old_value here", f"{FS_TEST_DIR}/replace_test.txt")
    results = sandbox.fs.replace_in_files([f"{FS_TEST_DIR}/replace_test.txt"], "old_value", "new_value")
    assert len(results) == 1
    assert results[0].success, f"Replace failed: {results[0]}"
    content = sandbox.fs.download_file(f"{FS_TEST_DIR}/replace_test.txt")
    assert b"new_value" in content, f"Expected 'new_value' in content, got {content!r}"
    assert b"old_value" not in content


def test_set_file_permissions(sandbox):
    sandbox.fs.set_file_permissions(f"{FS_TEST_DIR}/hello.txt", mode="644")
    info = sandbox.fs.get_file_info(f"{FS_TEST_DIR}/hello.txt")
    # Verify file still accessible after permission change
    assert info.name == "hello.txt"


def test_move_files(sandbox):
    sandbox.fs.upload_file(b"move me", f"{FS_TEST_DIR}/to_move.txt")
    sandbox.fs.move_files(f"{FS_TEST_DIR}/to_move.txt", f"{FS_TEST_DIR}/moved.txt")
    info = sandbox.fs.get_file_info(f"{FS_TEST_DIR}/moved.txt")
    assert info.name == "moved.txt"
    content = sandbox.fs.download_file(f"{FS_TEST_DIR}/moved.txt")
    assert content == b"move me"


def test_delete_file(sandbox):
    sandbox.fs.upload_file(b"delete me", f"{FS_TEST_DIR}/deletable.txt")
    sandbox.fs.delete_file(f"{FS_TEST_DIR}/deletable.txt")
    search = sandbox.fs.search_files(FS_TEST_DIR, "deletable.txt")
    assert not any("deletable.txt" in f for f in search.files), "File should have been deleted"


def test_nested_directory_operations(sandbox):
    nested = f"{FS_TEST_DIR}/nested/deep/dir"
    sandbox.fs.create_folder(nested, "755")
    sandbox.fs.upload_file(b"deep file", f"{nested}/deep.txt")
    content = sandbox.fs.download_file(f"{nested}/deep.txt")
    assert content == b"deep file"


def test_upload_binary_content(sandbox):
    binary_data = bytes(range(256))
    sandbox.fs.upload_file(binary_data, f"{FS_TEST_DIR}/binary.bin")
    content = sandbox.fs.download_file(f"{FS_TEST_DIR}/binary.bin")
    assert content == binary_data, "Binary content should round-trip exactly"


# ===========================================================================
# Process Execution
# ===========================================================================


def test_exec_basic_echo(sandbox):
    resp = sandbox.process.exec("echo hello")
    assert resp.exit_code == 0
    assert "hello" in resp.result


def test_exec_with_cwd(sandbox):
    resp = sandbox.process.exec("pwd", cwd="/tmp")
    assert resp.exit_code == 0
    assert "/tmp" in resp.result


def test_exec_with_env_vars(sandbox):
    resp = sandbox.process.exec("echo $MY_VAR", env={"MY_VAR": "test_value"})
    assert resp.exit_code == 0
    assert "test_value" in resp.result


def test_exec_with_multiple_env_vars(sandbox):
    resp = sandbox.process.exec(
        'echo "$A $B"',
        env={"A": "first", "B": "second"},
    )
    assert resp.exit_code == 0
    assert "first" in resp.result
    assert "second" in resp.result


def test_exec_nonzero_exit_code(sandbox):
    resp = sandbox.process.exec("exit 42")
    assert resp.exit_code == 42, f"Expected exit code 42, got {resp.exit_code}"


def test_exec_captures_stderr(sandbox):
    resp = sandbox.process.exec("echo err >&2")
    assert resp.exit_code == 0
    # stderr goes to result in the unified output
    assert "err" in resp.result


def test_code_run_python_print(sandbox):
    resp = sandbox.process.code_run('print("hello from python")')
    assert resp.exit_code == 0
    assert "hello from python" in resp.result


def test_code_run_multiline_python(sandbox):
    code = """
x = 10
y = 20
print(f"sum={x+y}")
"""
    resp = sandbox.process.code_run(code)
    assert resp.exit_code == 0
    assert "sum=30" in resp.result


def test_code_run_stderr_output(sandbox):
    code = 'import sys; sys.stderr.write("stderr_msg\\n")'
    resp = sandbox.process.code_run(code)
    assert resp.exit_code == 0
    assert "stderr_msg" in resp.result


def test_code_run_syntax_error(sandbox):
    resp = sandbox.process.code_run("def incomplete(")
    assert resp.exit_code != 0, "Syntax error should produce non-zero exit code"


# ===========================================================================
# Session Management
# ===========================================================================


def test_create_session(sandbox):
    sid = "e2e-session"
    try:
        sandbox.process.delete_session(sid)
    except Exception:
        pass
    sandbox.process.create_session(sid)
    session = sandbox.process.get_session(sid)
    assert session.session_id == sid


def test_get_session_details(sandbox):
    session = sandbox.process.get_session("e2e-session")
    assert session.session_id == "e2e-session"
    assert isinstance(session.commands, list)


def test_execute_session_command(sandbox):
    out = sandbox.process.execute_session_command(
        "e2e-session",
        SessionExecuteRequest(command="echo session_test"),
    )
    assert "session_test" in (out.stdout or out.output or ""), f"Expected 'session_test' in output, got {out}"


def test_session_state_persistence(sandbox):
    sandbox.process.execute_session_command(
        "e2e-session",
        SessionExecuteRequest(command="export E2E_VAR=persisted"),
    )
    out = sandbox.process.execute_session_command(
        "e2e-session",
        SessionExecuteRequest(command="echo $E2E_VAR"),
    )
    assert "persisted" in (
        out.stdout or out.output or ""
    ), f"Expected 'persisted' in output, got stdout={out.stdout!r} output={out.output!r}"


def test_get_session_command_logs(sandbox):
    out = sandbox.process.execute_session_command(
        "e2e-session",
        SessionExecuteRequest(command="echo log_check"),
    )
    # Give a brief moment for logs to be available
    time.sleep(0.5)
    logs = sandbox.process.get_session_command_logs("e2e-session", out.cmd_id)
    assert logs.stdout is not None or logs.output is not None, "Expected some log output"


def test_list_sessions_includes_ours(sandbox):
    sessions = sandbox.process.list_sessions()
    sids = [s.session_id for s in sessions]
    assert "e2e-session" in sids, f"Expected 'e2e-session' in {sids}"


def test_delete_session(sandbox):
    sandbox.process.delete_session("e2e-session")
    sessions = sandbox.process.list_sessions()
    sids = [s.session_id for s in sessions]
    assert "e2e-session" not in sids, f"Session should have been deleted, still in {sids}"


# ===========================================================================
# Git Operations
# ===========================================================================

GIT_REPO_PATH = "e2e-git-repo"


@pytest.mark.timeout(120)
def test_clone_public_repo(sandbox):
    try:
        sandbox.fs.delete_file(GIT_REPO_PATH, recursive=True)
    except Exception:
        pass
    sandbox.git.clone("https://github.com/octocat/Hello-World.git", GIT_REPO_PATH)
    info = sandbox.fs.get_file_info(GIT_REPO_PATH)
    assert info.is_dir is True


def test_git_status_has_current_branch(sandbox):
    status = sandbox.git.status(GIT_REPO_PATH)
    assert status.current_branch, "Expected current_branch to be set"


def test_git_branches_returns_list(sandbox):
    branches = sandbox.git.branches(GIT_REPO_PATH)
    assert len(branches.branches) > 0, "Expected at least one branch"


def test_create_branch(sandbox):
    sandbox.git.create_branch(GIT_REPO_PATH, "e2e-test-branch")
    branches = sandbox.git.branches(GIT_REPO_PATH)
    branch_names = branches.branches
    assert "e2e-test-branch" in branch_names, f"Expected 'e2e-test-branch' in {branch_names}"


def test_checkout_branch(sandbox):
    sandbox.git.checkout_branch(GIT_REPO_PATH, "e2e-test-branch")
    status = sandbox.git.status(GIT_REPO_PATH)
    assert status.current_branch == "e2e-test-branch", f"Expected 'e2e-test-branch', got {status.current_branch!r}"


def test_add_files(sandbox):
    sandbox.fs.upload_file(b"git test file", f"{GIT_REPO_PATH}/e2e_new.txt")
    sandbox.git.add(GIT_REPO_PATH, ["e2e_new.txt"])
    status = sandbox.git.status(GIT_REPO_PATH)
    # After add, there should be staged changes — file_status list non-empty
    assert status.file_status is not None


def test_commit_returns_sha(sandbox):
    result = sandbox.git.commit(
        path=GIT_REPO_PATH,
        message="e2e test commit",
        author="E2E Test",
        email="e2e@test.com",
    )
    assert result.sha, f"Expected a commit SHA, got {result.sha!r}"
    assert len(result.sha) >= 7, "SHA should be at least 7 characters"


def test_delete_branch(sandbox):
    sandbox.git.checkout_branch(GIT_REPO_PATH, "master")
    sandbox.git.delete_branch(GIT_REPO_PATH, "e2e-test-branch")
    branches = sandbox.git.branches(GIT_REPO_PATH)
    branch_names = branches.branches
    assert "e2e-test-branch" not in branch_names, f"Branch should be deleted, still in {branch_names}"


@pytest.mark.timeout(120)
def test_clone_specific_branch(sandbox):
    clone_path = "e2e-git-branch-clone"
    try:
        sandbox.fs.delete_file(clone_path, recursive=True)
    except Exception:
        pass
    sandbox.git.clone(
        "https://github.com/octocat/Hello-World.git",
        clone_path,
        branch="test",
    )
    status = sandbox.git.status(clone_path)
    assert status.current_branch == "test", f"Expected branch 'test', got {status.current_branch!r}"
    # Cleanup
    sandbox.fs.delete_file(clone_path, recursive=True)


# ===========================================================================
# Code Interpreter
# ===========================================================================


def test_run_code_simple_print(sandbox):
    result = sandbox.code_interpreter.run_code('print("ci_hello")')
    assert "ci_hello" in result.stdout, f"Expected 'ci_hello' in stdout, got {result.stdout!r}"
    assert result.error is None


def test_run_code_state_persistence(sandbox):
    sandbox.code_interpreter.run_code("ci_var = 42")
    result = sandbox.code_interpreter.run_code("print(ci_var)")
    assert "42" in result.stdout, f"Expected '42' in stdout, got {result.stdout!r}"


def test_create_context(sandbox):
    ctx = sandbox.code_interpreter.create_context()
    assert ctx.id, "Context should have a non-empty ID"
    sandbox.code_interpreter.delete_context(ctx)


def test_run_code_in_isolated_context(sandbox):
    ctx = sandbox.code_interpreter.create_context()
    sandbox.code_interpreter.run_code("isolated_var = 99", context=ctx)
    result = sandbox.code_interpreter.run_code("print(isolated_var)", context=ctx)
    assert "99" in result.stdout
    sandbox.code_interpreter.delete_context(ctx)


def test_list_contexts(sandbox):
    ctx = sandbox.code_interpreter.create_context()
    contexts = sandbox.code_interpreter.list_contexts()
    ctx_ids = [c.id for c in contexts]
    assert ctx.id in ctx_ids, f"Expected {ctx.id} in {ctx_ids}"
    sandbox.code_interpreter.delete_context(ctx)


def test_delete_context(sandbox):
    ctx = sandbox.code_interpreter.create_context()
    sandbox.code_interpreter.delete_context(ctx)
    contexts = sandbox.code_interpreter.list_contexts()
    ctx_ids = [c.id for c in contexts]
    assert ctx.id not in ctx_ids, f"Context {ctx.id} should have been deleted"


def test_run_code_with_error(sandbox):
    result = sandbox.code_interpreter.run_code("raise ValueError('boom')")
    assert result.error is not None, "Expected an error for raised exception"
    assert "ValueError" in result.error.name or "ValueError" in result.error.value


# ===========================================================================
# LSP Server Operations
# ===========================================================================


LSP_PROJECT_DIR = "e2e-lsp-project"
LSP_FILE_PATH = f"{LSP_PROJECT_DIR}/sample.py"


def test_lsp_create_and_start_server(sandbox):
    global lsp_server

    sandbox.fs.create_folder(LSP_PROJECT_DIR, "755")
    sandbox.fs.upload_file(
        b'class Greeter:\n    def greet(self) -> str:\n        return "hello"\n\ngreeter = Greeter()\ngreeter.\n',
        LSP_FILE_PATH,
    )
    lsp_server = sandbox.create_lsp_server("python", LSP_PROJECT_DIR)
    lsp_server.start()


def test_lsp_did_open(sandbox):
    assert lsp_server is not None, "LSP server should be started first"
    lsp_server.did_open(LSP_FILE_PATH)
    time.sleep(5)


def test_lsp_document_symbols(sandbox):
    assert lsp_server is not None, "LSP server should be started first"
    symbols = lsp_server.document_symbols(LSP_FILE_PATH)
    assert len(symbols) > 0, "Expected document symbols"
    names = [symbol.name for symbol in symbols]
    assert "Greeter" in names


def test_lsp_workspace_symbols(sandbox):
    assert lsp_server is not None, "LSP server should be started first"
    symbols = lsp_server.sandbox_symbols("Greeter")
    assert len(symbols) > 0, "Expected workspace symbols"


def test_lsp_did_close(sandbox):
    assert lsp_server is not None, "LSP server should be started first"
    lsp_server.did_close(LSP_FILE_PATH)


def test_lsp_stop_server(sandbox):
    global lsp_server

    assert lsp_server is not None, "LSP server should be started first"
    lsp_server.stop()
    lsp_server = None


# ===========================================================================
# PTY Operations
# ===========================================================================


def test_create_pty_session_and_list_it(sandbox):
    global pty_session_id

    handle = sandbox.process.create_pty_session(
        id=f"e2e-pty-{uuid.uuid4().hex[:8]}",
        pty_size=PtySize(rows=24, cols=80),
    )

    pty_session_id = handle.session_id
    handle.disconnect()

    sessions = sandbox.process.list_pty_sessions()
    assert any(session.id == pty_session_id for session in sessions)


def test_get_pty_session_info(sandbox):
    assert pty_session_id, "PTY session was not created in previous test"

    info = sandbox.process.get_pty_session_info(pty_session_id)
    assert info.id == pty_session_id


def test_resize_pty_session(sandbox):
    assert pty_session_id, "PTY session was not created in previous test"

    info = sandbox.process.resize_pty_session(pty_session_id, PtySize(rows=30, cols=100))
    assert info.rows == 30
    assert info.cols == 100


def test_connect_pty_session_write_read_and_close(sandbox):
    assert pty_session_id, "PTY session was not created in previous test"

    output_chunks: list[str] = []
    handle = sandbox.process.connect_pty_session(pty_session_id)

    try:
        wait_thread = threading.Thread(
            target=lambda: output_chunks.append(
                handle.wait(
                    on_data=lambda data: output_chunks.append(data.decode("utf-8", errors="ignore")),
                    timeout=15,
                )
            )
        )
        wait_thread.start()
        time.sleep(1)
        handle.send_input('printf "pty-output\\n"\n')
        time.sleep(2)
        handle.send_input("exit\n")
        wait_thread.join(timeout=20)
        assert "pty-output" in "".join(str(c) for c in output_chunks)
    finally:
        handle.disconnect()


# ===========================================================================
# Additional Process and Error Handling
# ===========================================================================


def test_exec_nonexistent_path_returns_failure(sandbox):
    resp = sandbox.process.exec("ls /definitely-missing-e2e-path")
    assert resp.exit_code != 0


def test_download_nonexistent_file_raises(sandbox):
    with pytest.raises(Exception):
        sandbox.fs.download_file(f"{FS_TEST_DIR}/does-not-exist.txt")


def test_duplicate_session_creation_rejects_or_is_idempotent(sandbox):
    sid = f"duplicate-session-{uuid.uuid4().hex[:8]}"
    sandbox.process.create_session(sid)

    duplicate_succeeded = False
    try:
        sandbox.process.create_session(sid)
        duplicate_succeeded = True
    except Exception as exc:
        assert str(exc), "Expected non-empty error message for duplicate session rejection"

    if duplicate_succeeded:
        sessions = sandbox.process.list_sessions()
        matching = [s for s in sessions if s.session_id == sid]
        assert len(matching) == 1, f"Idempotent duplicate should yield exactly 1 session, got {len(matching)}"

    sandbox.process.delete_session(sid)


def test_exec_timeout_code_path(sandbox):
    try:
        resp = sandbox.process.exec("sleep 2", timeout=1)
        assert resp.exit_code != 0
    except Exception as exc:
        assert "timeout" in _error_message(exc)


def test_code_run_returns_list_and_dict_data(sandbox):
    resp = sandbox.process.code_run('import json\nprint(json.dumps({"items": [1, 2, 3], "meta": {"ok": True}}))')
    assert resp.exit_code == 0
    assert json.loads(resp.result.strip()) == {"items": [1, 2, 3], "meta": {"ok": True}}


def test_long_running_command_execution(sandbox):
    resp = sandbox.process.exec(
        "python - <<'PY'\nimport time\ntime.sleep(1)\nprint('long-run-complete')\nPY",
        timeout=10,
    )
    assert resp.exit_code == 0
    assert "long-run-complete" in resp.result


# ===========================================================================
# Additional Sandbox Operations
# ===========================================================================


def test_declarative_image_build(daytona_client):
    cache_key = f"e2e-build-{uuid.uuid4().hex}"
    build_logs: list[str] = []

    def on_build_log(chunk: str) -> None:
        if chunk.strip():
            build_logs.append(chunk)

    image = Image.debian_slim("3.12").pip_install(["numpy"]).env({"CACHE_BUSTER": cache_key})
    params = CreateSandboxFromImageParams(
        image=image,
        language="python",
        name=f"sdk-py-e2e-build-{uuid.uuid4().hex[:8]}",
    )
    sb = daytona_client.create(params, timeout=300, on_snapshot_create_logs=on_build_log)

    try:
        state = str(getattr(sb.state, "value", sb.state)).lower()
        assert build_logs, "Expected streamed build logs for declarative image build"
        assert state == "started"

        result = sb.process.exec('python3 -c "import numpy; print(numpy.__version__)"')
        assert result.exit_code == 0
        assert result.result.strip()
    finally:
        try:
            daytona_client.delete(sb)
        except Exception:
            pass


def test_archive_and_unarchive_lifecycle_if_supported(sandbox):
    try:
        sandbox.stop(timeout=120)
        sandbox.archive()
        state = str(getattr(sandbox.state, "value", sandbox.state)).lower()
        assert state in {"archived", "archiving", "stopped"}
        sandbox.start(timeout=120)
        state = str(getattr(sandbox.state, "value", sandbox.state)).lower()
        assert state == "started"
    except Exception as exc:
        if any(part in _error_message(exc) for part in ["archive", "not supported"]):
            sandbox.start(timeout=120)
            return
        raise


def test_sandbox_remains_usable_after_archive_cycle(sandbox):
    sandbox.refresh_data()
    resp = sandbox.process.exec("echo post-archive-check")
    assert resp.exit_code == 0
    assert "post-archive-check" in resp.result


# ===========================================================================
# Volume Operations
# ===========================================================================

VOLUME_NAME = f"e2e-vol-{uuid.uuid4().hex[:8]}"


def test_volume_create(daytona_client):
    vol = daytona_client.volume.create(VOLUME_NAME)
    assert vol.name == VOLUME_NAME, f"Expected name {VOLUME_NAME}, got {vol.name}"
    assert vol.id, "Volume should have an ID"


def test_volume_list_includes_created(daytona_client):
    volumes = daytona_client.volume.list()
    names = [v.name for v in volumes]
    assert VOLUME_NAME in names, f"Expected {VOLUME_NAME} in {names}"


def test_volume_get_by_name(daytona_client):
    vol = daytona_client.volume.get(VOLUME_NAME)
    assert vol.name == VOLUME_NAME


def _vol_state(vol) -> str:
    raw = getattr(vol, "state", "")
    return str(getattr(raw, "value", raw)).lower()


def _wait_volume_ready(client, name, max_wait=30):
    for _ in range(max_wait):
        vol = client.volume.get(name)
        if _vol_state(vol) in ("ready", "error"):
            return vol
        time.sleep(1)
    return client.volume.get(name)


def test_volume_delete(daytona_client):
    vol = _wait_volume_ready(daytona_client, VOLUME_NAME)
    daytona_client.volume.delete(vol)
    for _ in range(15):
        volumes = daytona_client.volume.list()
        names = [v.name for v in volumes]
        if VOLUME_NAME not in names:
            break
        time.sleep(1)
    else:
        volumes = daytona_client.volume.list()
        names = [v.name for v in volumes]
        assert VOLUME_NAME not in names, f"Volume should have been deleted, still in {names}"


def test_volume_get_with_create_flag(daytona_client):
    vol = daytona_client.volume.get(VOLUME_NAME, create=True)
    assert vol.name == VOLUME_NAME
    vol = _wait_volume_ready(daytona_client, VOLUME_NAME)
    daytona_client.volume.delete(vol)
    for _ in range(15):
        volumes = daytona_client.volume.list()
        if VOLUME_NAME not in [v.name for v in volumes]:
            break
        time.sleep(1)


# ===========================================================================
# Snapshot Operations
# ===========================================================================


def test_snapshot_list_returns_results(daytona_client):
    result = daytona_client.snapshot.list()
    assert result.total > 0, f"Expected at least one snapshot, got total={result.total}"
    assert len(result.items) > 0


def test_snapshot_list_with_pagination(daytona_client):
    result = daytona_client.snapshot.list(page=1, limit=2)
    assert len(result.items) <= 2, f"Expected at most 2 items, got {len(result.items)}"
    assert result.page == 1


def test_snapshot_get_by_name(daytona_client):
    listing = daytona_client.snapshot.list(page=1, limit=1)
    assert len(listing.items) > 0, "Need at least one snapshot to test get"
    name = listing.items[0].name
    snapshot = daytona_client.snapshot.get(name)
    assert snapshot.name == name, f"Expected name {name!r}, got {snapshot.name!r}"
