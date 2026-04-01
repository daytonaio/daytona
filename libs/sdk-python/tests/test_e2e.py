from __future__ import annotations

# pyright: reportMissingImports=false,reportUnknownMemberType=false,reportUntypedFunctionDecorator=false,reportUnknownParameterType=false,reportMissingParameterType=false,reportUnknownVariableType=false,reportUnknownArgumentType=false

import pytest

from daytona import (
    CreateSandboxFromSnapshotParams,
    Daytona,
    FileDownloadRequest,
    FileUpload,
    SessionExecuteRequest,
)


def _state_value(state: object) -> str:
    return str(getattr(state, "value", state)).lower()


@pytest.fixture(scope="module")
def daytona_client() -> Daytona:
    return Daytona()


@pytest.fixture(scope="module")
def sandbox(daytona_client: Daytona):
    params = CreateSandboxFromSnapshotParams(language="python")
    sandbox = daytona_client.create(params, timeout=120)
    yield sandbox
    try:
        daytona_client.delete(sandbox)
    except Exception:
        pass


@pytest.mark.timeout(120)
def test_sandbox_lifecycle_basics(sandbox):
    assert _state_value(sandbox.state) == "started", f"Expected sandbox state 'started', got {sandbox.state!r}"
    assert sandbox.id, "Sandbox ID should be present"
    assert sandbox.name, "Sandbox name should be present"
    assert sandbox.organization_id, "Sandbox organization_id should be present"

    user_home = sandbox.get_user_home_dir()
    assert isinstance(user_home, str) and user_home, "Expected non-empty home directory path"

    work_dir = sandbox.get_work_dir()
    assert isinstance(work_dir, str) and work_dir, "Expected non-empty working directory path"


def test_set_labels_and_autostop_and_refresh(sandbox):
    labels = sandbox.set_labels({"test": "e2e"})
    assert isinstance(labels, dict), "set_labels should return a dictionary"
    assert labels.get("test") == "e2e", "Expected sandbox label test=e2e"

    sandbox.set_autostop_interval(30)
    assert sandbox.auto_stop_interval == 30, "Expected auto_stop_interval to be updated to 30"

    sandbox.refresh_data()
    assert sandbox.id, "Sandbox should still have an ID after refresh_data"


def test_create_folder(sandbox):
    try:
        sandbox.fs.delete_file("test-dir", recursive=True)
    except Exception:
        pass
    sandbox.fs.create_folder("test-dir", "755")
    listed = sandbox.fs.list_files("test-dir")
    assert isinstance(listed, list), "list_files should return a list"


def test_upload_and_download_file(sandbox):
    sandbox.fs.upload_files([FileUpload(source=b"hello world", destination="test-dir/hello.txt")])

    files = sandbox.fs.list_files("test-dir")
    names = [f.name for f in files]
    assert "hello.txt" in names, f"Expected hello.txt in test-dir, got {names}"

    info = sandbox.fs.get_file_info("test-dir/hello.txt")
    assert info.size is not None and info.size > 0, f"Expected file size > 0, got {info.size}"

    download_response = sandbox.fs.download_files([FileDownloadRequest(source="test-dir/hello.txt")])
    assert len(download_response) == 1, "Expected exactly one downloaded file"
    content = download_response[0].result
    assert content == b"hello world", f"Expected exact file bytes, got {content!r}"


def test_find_search_replace_move_and_delete_file(sandbox):
    matches = sandbox.fs.find_files("test-dir", "hello")
    assert len(matches) > 0, "Expected find_files to return at least one match"

    search_result = sandbox.fs.search_files("test-dir", "*.txt")
    assert any(path.endswith("hello.txt") for path in search_result.files), (
        f"Expected hello.txt in search results, got {search_result.files}"
    )

    replace_results = sandbox.fs.replace_in_files(["test-dir/hello.txt"], "hello", "world")
    assert len(replace_results) == 1, "Expected one replace result"
    assert replace_results[0].success, f"Expected replace success, got {replace_results[0]}"

    replaced_content = sandbox.fs.download_file("test-dir/hello.txt")
    assert b"world" in replaced_content, f"Expected replaced content to contain 'world', got {replaced_content!r}"

    sandbox.fs.move_files("test-dir/hello.txt", "test-dir/moved.txt")
    moved_info = sandbox.fs.get_file_info("test-dir/moved.txt")
    assert moved_info.name == "moved.txt", f"Expected moved file info name to be moved.txt, got {moved_info.name!r}"

    sandbox.fs.delete_file("test-dir/moved.txt")
    remaining = sandbox.fs.search_files("test-dir", "moved.txt")
    assert not any(path.endswith("moved.txt") for path in remaining.files), "Expected moved.txt to be deleted"


def test_process_exec_and_code_run(sandbox):
    hello = sandbox.process.exec("echo hello")
    assert hello.exit_code == 0, f"Expected exit_code 0, got {hello.exit_code}"
    assert "hello" in hello.result, f"Expected output to contain 'hello', got {hello.result!r}"

    ls_out = sandbox.process.exec("ls /", cwd="/tmp")
    assert ls_out.exit_code == 0, f"Expected ls / to succeed, got {ls_out.exit_code}"

    env_out = sandbox.process.exec("echo $MY_VAR", env={"MY_VAR": "test123"})
    assert env_out.exit_code == 0, f"Expected env command success, got {env_out.exit_code}"
    assert "test123" in env_out.result, f"Expected env output to contain test123, got {env_out.result!r}"

    code_out = sandbox.process.code_run('print("hello from python")')
    assert code_out.exit_code == 0, f"Expected code_run to succeed, got {code_out.exit_code}"
    assert "hello from python" in code_out.result, (
        f"Expected code_run output to contain 'hello from python', got {code_out.result!r}"
    )

    fail_out = sandbox.process.exec("exit 1")
    assert fail_out.exit_code != 0, "Expected non-zero exit code for 'exit 1'"


def test_process_session_management(sandbox):
    session_id = "test-session"
    sandbox.process.create_session(session_id)

    session = sandbox.process.get_session(session_id)
    assert session.session_id == session_id, f"Expected session_id {session_id}, got {session.session_id!r}"

    sandbox.process.execute_session_command(session_id, SessionExecuteRequest(command="export FOO=bar"))
    out = sandbox.process.execute_session_command(session_id, SessionExecuteRequest(command="echo $FOO"))
    assert "bar" in out.stdout, f"Expected session stdout to contain 'bar', got {out.stdout!r}"

    sessions = sandbox.process.list_sessions()
    session_ids = [s.session_id for s in sessions]
    assert session_id in session_ids, f"Expected {session_id} in active sessions, got {session_ids}"

    sandbox.process.delete_session(session_id)


@pytest.mark.timeout(120)
def test_git_operations(sandbox):
    repo_path = "hello-world"
    try:
        sandbox.fs.delete_file(repo_path, recursive=True)
    except Exception:
        pass

    sandbox.git.clone("https://github.com/octocat/Hello-World.git", repo_path)
    status = sandbox.git.status(repo_path)
    assert status.current_branch, "Expected git status to include current_branch"

    branches = sandbox.git.branches(repo_path)
    assert len(branches.branches) > 0, "Expected at least one git branch"


def test_daytona_client_list_and_get(daytona_client: Daytona, sandbox):
    sandboxes = daytona_client.list()
    assert sandboxes.total > 0, f"Expected list().total > 0, got {sandboxes.total}"

    fetched = daytona_client.get(sandbox.id)
    assert fetched.id == sandbox.id, f"Expected fetched sandbox id {sandbox.id}, got {fetched.id}"
