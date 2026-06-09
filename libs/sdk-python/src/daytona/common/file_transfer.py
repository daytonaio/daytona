# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from collections.abc import Callable, Mapping
from typing import Any

from python_multipart.multipart import MultipartParser, MultipartState, parse_options_header

from daytona_toolbox_api_client import FilesDownloadRequest

from .errors import DaytonaError


def serialize_download_request(api_client: Any, remote_path: str) -> tuple[str, str, dict[str, str], Any]:
    method, url, headers, body, *_ = api_client._download_files_serialize(
        download_files=FilesDownloadRequest(paths=[remote_path]),
        _request_auth=None,
        _content_type=None,
        _headers=None,
        _host_index=None,
    )
    return method, url, headers, body


def parse_content_type_boundary(headers: Mapping[str, str]) -> bytes:
    content_type_raw, options = parse_options_header(headers.get("Content-Type", ""))
    if not (content_type_raw == b"multipart/form-data" and b"boundary" in options):
        raise DaytonaError(f"Unexpected Content-Type: {content_type_raw!r}")
    return options[b"boundary"]


def create_multipart_parser(
    boundary: bytes,
    on_part_begin: Callable[[], None],
    on_header_field: Callable[[bytes, int, int], None],
    on_header_value: Callable[[bytes, int, int], None],
    on_header_end: Callable[[], None],
    on_headers_finished: Callable[[], None],
    on_part_data: Callable[[bytes, int, int], None],
    on_part_end: Callable[[], None],
) -> MultipartParser:
    return MultipartParser(
        boundary,
        callbacks={
            "on_part_begin": on_part_begin,
            "on_header_field": on_header_field,
            "on_header_value": on_header_value,
            "on_header_end": on_header_end,
            "on_headers_finished": on_headers_finished,
            "on_part_data": on_part_data,
            "on_part_end": on_part_end,
        },
    )


def raise_if_multipart_truncated(parser: MultipartParser, remote_path: str) -> None:
    """Raise if the multipart stream ended before the closing boundary.

    python_multipart can silently drop boundary look-ahead bytes at finalize().
    Call after parser.finalize() so short downloads become retryable errors.
    """
    if parser.state != MultipartState.END:
        msg = (
            f"Truncated multipart response for {remote_path!r}: "
            f"closing boundary not received (parser state={parser.state.name})"
        )
        raise DaytonaError(msg)
