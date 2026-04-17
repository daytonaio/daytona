# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import io
import threading
from collections.abc import Callable
from unittest.mock import MagicMock

import pytest

from daytona._sync.filesystem import FileSystem
from daytona.common.errors import DaytonaError, DaytonaValidationError, DownloadAbortedError, UploadAbortedError
from daytona.common.filesystem import DownloadFileOptions, FileDownloadResponse, ProgressCountingIO, UploadFileOptions


def make_filesystem() -> FileSystem:
    """Create a FileSystem instance with mocked internals."""
    fs = object.__new__(FileSystem)
    fs._api_client = MagicMock()  # pylint: disable=protected-access
    return fs


def make_download_response(data: bytes, source: str = "/remote/file.bin") -> FileDownloadResponse:
    return FileDownloadResponse(source=source, result=data, error=None, error_details=None)


def collect_progress(calls: list[int]) -> Callable[[int], None]:
    """Return a progress callback that appends to calls list."""
    return calls.append


# ---- upload_file positional dispatch ----


def test_upload_file_positional_bytes_calls_upload_files() -> None:
    fs = make_filesystem()
    fs.upload_files = MagicMock(return_value=None)
    fs.upload_file(b"hello", "/remote/hello.txt")
    fs.upload_files.assert_called_once()
    call_args = fs.upload_files.call_args
    files = call_args[0][0]
    assert len(files) == 1
    assert files[0].destination == "/remote/hello.txt"


def test_upload_file_positional_str_calls_upload_files() -> None:
    fs = make_filesystem()
    fs.upload_files = MagicMock(return_value=None)
    fs.upload_file("/local/file.txt", "/remote/file.txt")
    fs.upload_files.assert_called_once()
    call_args = fs.upload_files.call_args
    files = call_args[0][0]
    assert files[0].source == "/local/file.txt"
    assert files[0].destination == "/remote/file.txt"


def test_upload_file_positional_with_timeout() -> None:
    fs = make_filesystem()
    fs.upload_files = MagicMock(return_value=None)
    fs.upload_file(b"data", "/remote/file.bin", 300)
    fs.upload_files.assert_called_once()
    call_args = fs.upload_files.call_args
    timeout = call_args[0][1]
    assert timeout == 300


# ---- upload_file options-bag dispatch ----


def test_upload_file_options_with_bytes_source_calls_upload_files() -> None:
    fs = make_filesystem()
    fs.upload_files = MagicMock(return_value=None)
    opts = UploadFileOptions(source=b"hello world", destination="/remote/hello.txt")
    fs.upload_file(options=opts)
    fs.upload_files.assert_called_once()


def test_upload_file_options_with_io_source_and_progress() -> None:
    fs = make_filesystem()
    fs.upload_files = MagicMock(return_value=None)
    buf = io.BytesIO(b"chunk1chunk2chunk3")
    progress_calls: list[int] = []
    opts = UploadFileOptions(
        source=buf,
        destination="/remote/file.bin",
        on_progress=progress_calls.append,
    )
    fs.upload_file(options=opts)
    fs.upload_files.assert_called_once()
    # The ProgressCountingIO wraps the source; progress is called during upload_files
    # Since upload_files is mocked, progress won't be triggered here — just verify dispatch
    assert True


def test_upload_file_options_io_source_progress_counted() -> None:
    """Verify ProgressCountingIO counts bytes correctly when read() is called."""
    data = b"a" * 1000
    buf = io.BytesIO(data)
    progress_calls: list[int] = []
    wrapper = ProgressCountingIO(buf, progress_calls.append, None)
    chunk = wrapper.read(500)
    assert len(chunk) == 500
    assert progress_calls == [500]
    chunk2 = wrapper.read(500)
    assert len(chunk2) == 500
    assert progress_calls == [500, 1000]


# ---- upload_file abort ----


def test_upload_file_abort_signal_pre_set_raises() -> None:
    fs = make_filesystem()
    fs.upload_files = MagicMock(return_value=None)
    signal = threading.Event()
    signal.set()
    opts = UploadFileOptions(source=b"data", destination="/remote/file.bin", signal=signal)
    with pytest.raises(UploadAbortedError):
        fs.upload_file(options=opts)


def test_upload_file_abort_signal_during_io_read() -> None:
    """Signal set before read triggers UploadAbortedError from ProgressCountingIO."""
    signal = threading.Event()
    signal.set()
    buf = io.BytesIO(b"some data")
    wrapper = ProgressCountingIO(buf, None, signal)
    with pytest.raises(UploadAbortedError):
        _ = wrapper.read(1024)


def test_progress_counting_io_zero_bytes_no_progress_callback() -> None:
    """Progress callback must not fire on an EOF (zero-byte) read."""
    buf = io.BytesIO(b"")
    progress_calls: list[int] = []
    wrapper = ProgressCountingIO(buf, progress_calls.append, None)
    chunk = wrapper.read(65536)
    assert chunk == b""
    assert not progress_calls


def test_progress_counting_io_source_read_error_propagates() -> None:
    """IOError from the inner IO source propagates unchanged through ProgressCountingIO."""
    inner = MagicMock()
    inner.read.side_effect = IOError("disk error")
    wrapper = ProgressCountingIO(inner, None, None)
    with pytest.raises(IOError, match="disk error"):
        _ = wrapper.read(1024)


def test_upload_file_on_progress_with_bytes_source_raises_validation_error() -> None:
    fs = make_filesystem()
    fs.upload_files = MagicMock(return_value=None)
    progress_calls: list[int] = []
    opts = UploadFileOptions(
        source=b"data",
        destination="/remote/file.bin",
        on_progress=collect_progress(progress_calls),
    )
    with pytest.raises(DaytonaValidationError):
        fs.upload_file(options=opts)


def test_upload_file_on_progress_with_str_source_raises_validation_error() -> None:
    fs = make_filesystem()
    fs.upload_files = MagicMock(return_value=None)
    progress_calls: list[int] = []
    opts = UploadFileOptions(
        source="/local/file.txt",
        destination="/remote/file.bin",
        on_progress=collect_progress(progress_calls),
    )
    with pytest.raises(DaytonaValidationError):
        fs.upload_file(options=opts)


# ---- download_file positional dispatch ----


def test_download_file_positional_buffer_return() -> None:
    fs = make_filesystem()
    expected = b"file content"
    fs.download_files = MagicMock(return_value=[make_download_response(expected, "/remote/file.bin")])
    result = fs.download_file("/remote/file.bin")
    assert result == expected
    fs.download_files.assert_called_once()


def test_download_file_positional_with_timeout() -> None:
    fs = make_filesystem()
    expected = b"content"
    fs.download_files = MagicMock(return_value=[make_download_response(expected, "/remote/file.bin")])
    result = fs.download_file("/remote/file.bin", 300)
    assert result == expected


def test_download_file_positional_file_path() -> None:
    fs = make_filesystem()
    fs.download_files = MagicMock(
        return_value=[
            FileDownloadResponse(
                source="/remote/file.bin",
                result="/local/out.bin",
                error=None,
                error_details=None,
            )
        ]
    )
    _ = fs.download_file("/remote/file.bin", "/local/out.bin")
    fs.download_files.assert_called_once()


# ---- download_file options-bag dispatch ----


def test_download_file_options_buffer_return() -> None:
    fs = make_filesystem()
    expected = b"buffer data"
    fs.download_files = MagicMock(return_value=[make_download_response(expected, "/remote/file.bin")])
    opts = DownloadFileOptions(remote_path="/remote/file.bin")
    result = fs.download_file(options=opts)
    assert result == expected


def test_download_file_options_abort_signal_pre_set_raises() -> None:
    fs = make_filesystem()
    fs.download_files = MagicMock(return_value=[make_download_response(b"data", "/remote/file.bin")])
    signal = threading.Event()
    signal.set()
    opts = DownloadFileOptions(remote_path="/remote/file.bin", signal=signal)
    with pytest.raises(DownloadAbortedError):
        _ = fs.download_file(options=opts)


def test_download_file_options_on_progress_no_stream_dest_raises() -> None:
    fs = make_filesystem()
    fs.download_files = MagicMock(return_value=[make_download_response(b"data", "/remote/file.bin")])
    progress_calls: list[int] = []
    opts = DownloadFileOptions(
        remote_path="/remote/file.bin",
        on_progress=collect_progress(progress_calls),
        # no destination — buffer return mode
    )
    with pytest.raises(DaytonaValidationError):
        _ = fs.download_file(options=opts)


def test_download_file_options_on_progress_str_dest_raises() -> None:
    fs = make_filesystem()
    fs.download_files = MagicMock(
        return_value=[
            FileDownloadResponse(
                source="/remote/file.bin",
                result="/local/out.bin",
                error=None,
                error_details=None,
            )
        ]
    )
    progress_calls: list[int] = []
    opts = DownloadFileOptions(
        remote_path="/remote/file.bin",
        destination="/local/out.bin",
        on_progress=collect_progress(progress_calls),
    )
    with pytest.raises(DaytonaValidationError):
        _ = fs.download_file(options=opts)


def test_download_file_options_io_destination_streams_data() -> None:
    fs = make_filesystem()
    data = b"streamed content"
    fs.download_files = MagicMock(return_value=[make_download_response(data, "/remote/file.bin")])
    dest = io.BytesIO()
    opts = DownloadFileOptions(remote_path="/remote/file.bin", destination=dest)
    _ = fs.download_file(options=opts)
    assert dest.getvalue() == data


def test_download_file_options_io_destination_with_progress() -> None:
    fs = make_filesystem()
    data = b"x" * 200000  # 200KB — crosses two 65536-byte chunks
    fs.download_files = MagicMock(return_value=[make_download_response(data, "/remote/file.bin")])
    dest = io.BytesIO()
    progress_calls: list[int] = []
    opts = DownloadFileOptions(
        remote_path="/remote/file.bin",
        destination=dest,
        on_progress=progress_calls.append,
    )
    _ = fs.download_file(options=opts)
    assert dest.getvalue() == data
    # Progress should have been called at least once
    assert len(progress_calls) >= 1
    # Final progress value should equal total bytes
    assert progress_calls[-1] == len(data)


def test_download_file_options_io_destination_abort_mid_stream() -> None:
    fs = make_filesystem()
    data = b"y" * 200000
    fs.download_files = MagicMock(return_value=[make_download_response(data, "/remote/file.bin")])
    dest = io.BytesIO()
    signal = threading.Event()

    call_count = [0]

    def on_progress(_bytes_written: int) -> None:
        call_count[0] += 1
        if call_count[0] >= 1:
            signal.set()

    opts = DownloadFileOptions(
        remote_path="/remote/file.bin",
        destination=dest,
        signal=signal,
        on_progress=on_progress,
    )
    with pytest.raises(DownloadAbortedError):
        _ = fs.download_file(options=opts)


# ---- Error class hierarchy ----


def test_upload_aborted_error_is_daytona_error() -> None:
    err = UploadAbortedError()
    assert isinstance(err, DaytonaError)
    assert str(err) == "Upload aborted"


def test_download_aborted_error_is_daytona_error() -> None:
    err = DownloadAbortedError()
    assert isinstance(err, DaytonaError)
    assert str(err) == "Download aborted"


def test_upload_aborted_error_custom_message() -> None:
    err = UploadAbortedError("custom message")
    assert str(err) == "custom message"


def test_download_aborted_error_custom_message() -> None:
    err = DownloadAbortedError("custom message")
    assert str(err) == "custom message"


# ---- destination write error (F3) ----


def test_download_file_io_destination_write_error_propagates() -> None:
    """IOError raised by IO destination is surfaced as DaytonaError (wrapped by @intercept_errors)."""
    fs = make_filesystem()
    data = b"some file content"
    fs.download_files = MagicMock(return_value=[make_download_response(data, "/remote/file.bin")])
    dest = MagicMock()
    dest.write.side_effect = IOError("disk full")
    opts = DownloadFileOptions(remote_path="/remote/file.bin", destination=dest)
    with pytest.raises(DaytonaError, match="disk full"):
        _ = fs.download_file(options=opts)
