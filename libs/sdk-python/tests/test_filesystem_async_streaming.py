# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import asyncio
import io
import threading
from collections.abc import Callable
from unittest.mock import AsyncMock, MagicMock

import pytest

from daytona._async.filesystem import AsyncFileSystem
from daytona.common.errors import DaytonaError, DaytonaValidationError, DownloadAbortedError, UploadAbortedError
from daytona.common.filesystem import DownloadFileOptions, FileDownloadResponse, ProgressCountingIO, UploadFileOptions


def make_async_filesystem() -> AsyncFileSystem:
    """Create an AsyncFileSystem instance with mocked internals."""
    fs = object.__new__(AsyncFileSystem)
    fs._api_client = MagicMock()  # pylint: disable=protected-access
    return fs


def make_download_response(data: bytes, source: str = "/remote/file.bin") -> FileDownloadResponse:
    return FileDownloadResponse(source=source, result=data, error=None, error_details=None)


def collect_progress(calls: list[int]) -> Callable[[int], None]:
    """Return a progress callback that appends to calls list."""
    return calls.append


# ---- upload_file positional dispatch ----


@pytest.mark.anyio
async def test_async_upload_file_positional_bytes_calls_upload_files() -> None:
    fs = make_async_filesystem()
    fs.upload_files = AsyncMock(return_value=None)
    await fs.upload_file(b"hello", "/remote/hello.txt")
    fs.upload_files.assert_called_once()
    call_args = fs.upload_files.call_args
    files = call_args[0][0]
    assert len(files) == 1
    assert files[0].destination == "/remote/hello.txt"


@pytest.mark.anyio
async def test_async_upload_file_positional_str_calls_upload_files() -> None:
    fs = make_async_filesystem()
    fs.upload_files = AsyncMock(return_value=None)
    await fs.upload_file("/local/file.txt", "/remote/file.txt")
    fs.upload_files.assert_called_once()
    call_args = fs.upload_files.call_args
    files = call_args[0][0]
    assert files[0].source == "/local/file.txt"
    assert files[0].destination == "/remote/file.txt"


@pytest.mark.anyio
async def test_async_upload_file_positional_with_timeout() -> None:
    fs = make_async_filesystem()
    fs.upload_files = AsyncMock(return_value=None)
    await fs.upload_file(b"data", "/remote/file.bin", 300)
    fs.upload_files.assert_called_once()
    call_args = fs.upload_files.call_args
    timeout = call_args[0][1]
    assert timeout == 300


# ---- upload_file options-bag dispatch ----


@pytest.mark.anyio
async def test_async_upload_file_options_with_bytes_source_calls_upload_files() -> None:
    fs = make_async_filesystem()
    fs.upload_files = AsyncMock(return_value=None)
    opts = UploadFileOptions(source=b"hello world", destination="/remote/hello.txt")
    await fs.upload_file(options=opts)
    fs.upload_files.assert_called_once()


@pytest.mark.anyio
async def test_async_upload_file_options_with_io_source() -> None:
    fs = make_async_filesystem()
    fs.upload_files = AsyncMock(return_value=None)
    buf = io.BytesIO(b"chunk1chunk2chunk3")
    opts = UploadFileOptions(source=buf, destination="/remote/file.bin")
    await fs.upload_file(options=opts)
    fs.upload_files.assert_called_once()


@pytest.mark.anyio
async def test_async_upload_file_abort_signal_pre_set_raises() -> None:
    fs = make_async_filesystem()
    fs.upload_files = AsyncMock(return_value=None)
    signal = threading.Event()
    signal.set()
    opts = UploadFileOptions(source=b"data", destination="/remote/file.bin", signal=signal)
    with pytest.raises(UploadAbortedError):
        await fs.upload_file(options=opts)


@pytest.mark.anyio
async def test_async_upload_file_on_progress_with_bytes_source_raises_validation_error() -> None:
    fs = make_async_filesystem()
    fs.upload_files = AsyncMock(return_value=None)
    progress_calls: list[int] = []
    opts = UploadFileOptions(
        source=b"data",
        destination="/remote/file.bin",
        on_progress=collect_progress(progress_calls),
    )
    with pytest.raises(DaytonaValidationError):
        await fs.upload_file(options=opts)


@pytest.mark.anyio
async def test_async_upload_file_on_progress_with_str_source_raises_validation_error() -> None:
    fs = make_async_filesystem()
    fs.upload_files = AsyncMock(return_value=None)
    progress_calls: list[int] = []
    opts = UploadFileOptions(
        source="/local/file.txt",
        destination="/remote/file.bin",
        on_progress=collect_progress(progress_calls),
    )
    with pytest.raises(DaytonaValidationError):
        await fs.upload_file(options=opts)


# ---- download_file positional dispatch ----


@pytest.mark.anyio
async def test_async_download_file_positional_buffer_return() -> None:
    fs = make_async_filesystem()
    expected = b"file content"
    fs.download_files = AsyncMock(return_value=[make_download_response(expected, "/remote/file.bin")])
    result = await fs.download_file("/remote/file.bin")
    assert result == expected
    fs.download_files.assert_called_once()


@pytest.mark.anyio
async def test_async_download_file_positional_with_timeout() -> None:
    fs = make_async_filesystem()
    expected = b"content"
    fs.download_files = AsyncMock(return_value=[make_download_response(expected, "/remote/file.bin")])
    result = await fs.download_file("/remote/file.bin", 300)
    assert result == expected


@pytest.mark.anyio
async def test_async_download_file_positional_file_path() -> None:
    fs = make_async_filesystem()
    fs.download_files = AsyncMock(
        return_value=[
            FileDownloadResponse(
                source="/remote/file.bin",
                result="/local/out.bin",
                error=None,
                error_details=None,
            )
        ]
    )
    _ = await fs.download_file("/remote/file.bin", "/local/out.bin")
    fs.download_files.assert_called_once()


# ---- download_file options-bag dispatch ----


@pytest.mark.anyio
async def test_async_download_file_options_buffer_return() -> None:
    fs = make_async_filesystem()
    expected = b"buffer data"
    fs.download_files = AsyncMock(return_value=[make_download_response(expected, "/remote/file.bin")])
    opts = DownloadFileOptions(remote_path="/remote/file.bin")
    result = await fs.download_file(options=opts)
    assert result == expected


@pytest.mark.anyio
async def test_async_download_file_options_abort_signal_pre_set_raises() -> None:
    fs = make_async_filesystem()
    fs.download_files = AsyncMock(return_value=[make_download_response(b"data", "/remote/file.bin")])
    signal = threading.Event()
    signal.set()
    opts = DownloadFileOptions(remote_path="/remote/file.bin", signal=signal)
    with pytest.raises(DownloadAbortedError):
        _ = await fs.download_file(options=opts)


@pytest.mark.anyio
async def test_async_download_file_options_on_progress_no_stream_dest_raises() -> None:
    fs = make_async_filesystem()
    fs.download_files = AsyncMock(return_value=[make_download_response(b"data", "/remote/file.bin")])
    progress_calls: list[int] = []
    opts = DownloadFileOptions(
        remote_path="/remote/file.bin",
        on_progress=collect_progress(progress_calls),
    )
    with pytest.raises(DaytonaValidationError):
        _ = await fs.download_file(options=opts)


@pytest.mark.anyio
async def test_async_download_file_options_on_progress_str_dest_raises() -> None:
    fs = make_async_filesystem()
    fs.download_files = AsyncMock(
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
        _ = await fs.download_file(options=opts)


@pytest.mark.anyio
async def test_async_download_file_options_io_destination_streams_data() -> None:
    fs = make_async_filesystem()
    data = b"streamed content"
    fs.download_files = AsyncMock(return_value=[make_download_response(data, "/remote/file.bin")])
    dest = io.BytesIO()
    opts = DownloadFileOptions(remote_path="/remote/file.bin", destination=dest)
    _ = await fs.download_file(options=opts)
    assert dest.getvalue() == data


@pytest.mark.anyio
async def test_async_download_file_options_io_destination_with_progress() -> None:
    fs = make_async_filesystem()
    data = b"x" * 200000  # 200KB — crosses two 65536-byte chunks
    fs.download_files = AsyncMock(return_value=[make_download_response(data, "/remote/file.bin")])
    dest = io.BytesIO()
    progress_calls: list[int] = []
    opts = DownloadFileOptions(
        remote_path="/remote/file.bin",
        destination=dest,
        on_progress=progress_calls.append,
    )
    _ = await fs.download_file(options=opts)
    assert dest.getvalue() == data
    assert len(progress_calls) >= 1
    assert progress_calls[-1] == len(data)


@pytest.mark.anyio
async def test_async_download_file_options_io_destination_abort_mid_stream() -> None:
    fs = make_async_filesystem()
    data = b"y" * 200000
    fs.download_files = AsyncMock(return_value=[make_download_response(data, "/remote/file.bin")])
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
        _ = await fs.download_file(options=opts)


# ---- asyncio.CancelledError handling (design §5.2) ----


@pytest.mark.anyio
async def test_async_upload_file_cancelled_error_becomes_upload_aborted_error() -> None:
    """asyncio.Task cancellation during upload must raise UploadAbortedError, not CancelledError."""
    fs = make_async_filesystem()
    fs.upload_files = AsyncMock(side_effect=asyncio.CancelledError())
    buf = io.BytesIO(b"data to upload")
    opts = UploadFileOptions(source=buf, destination="/remote/file.bin")
    with pytest.raises(UploadAbortedError):
        await fs.upload_file(options=opts)


@pytest.mark.anyio
async def test_async_download_file_cancelled_error_becomes_download_aborted_error() -> None:
    """asyncio.Task cancellation during download must raise DownloadAbortedError, not CancelledError."""
    fs = make_async_filesystem()
    fs.download_files = AsyncMock(side_effect=asyncio.CancelledError())
    opts = DownloadFileOptions(remote_path="/remote/file.bin")
    with pytest.raises(DownloadAbortedError):
        _ = await fs.download_file(options=opts)


# ---- error class hierarchy (mirrors sync equivalents) ----


def test_upload_aborted_error_is_daytona_error() -> None:
    err = UploadAbortedError()
    assert isinstance(err, DaytonaError)
    assert str(err) == "Upload aborted"


def test_download_aborted_error_is_daytona_error() -> None:
    err = DownloadAbortedError()
    assert isinstance(err, DaytonaError)
    assert str(err) == "Download aborted"


def test_upload_aborted_error_custom_message() -> None:
    err = UploadAbortedError("transfer interrupted")
    assert str(err) == "transfer interrupted"


def test_download_aborted_error_custom_message() -> None:
    err = DownloadAbortedError("transfer interrupted")
    assert str(err) == "transfer interrupted"


# ---- streaming edge cases (F3) ----


def test_async_progress_counting_io_zero_bytes_no_progress_callback() -> None:
    """Progress callback must not fire on an EOF (zero-byte) read."""
    buf = io.BytesIO(b"")
    progress_calls: list[int] = []
    wrapper = ProgressCountingIO(buf, progress_calls.append, None)
    chunk = wrapper.read(65536)
    assert chunk == b""
    assert not progress_calls


def test_async_progress_counting_io_source_read_error_propagates() -> None:
    """IOError from the inner IO source propagates unchanged through ProgressCountingIO."""
    inner = MagicMock()
    inner.read.side_effect = IOError("disk error")
    wrapper = ProgressCountingIO(inner, None, None)
    with pytest.raises(IOError, match="disk error"):
        _ = wrapper.read(1024)


@pytest.mark.anyio
async def test_async_download_file_io_destination_write_error_propagates() -> None:
    """IOError raised by IO destination is surfaced as DaytonaError (wrapped by @intercept_errors)."""
    fs = make_async_filesystem()
    data = b"some file content"
    fs.download_files = AsyncMock(return_value=[make_download_response(data, "/remote/file.bin")])
    dest = MagicMock()
    dest.write.side_effect = IOError("disk full")
    opts = DownloadFileOptions(remote_path="/remote/file.bin", destination=dest)
    with pytest.raises(DaytonaError, match="disk full"):
        _ = await fs.download_file(options=opts)
