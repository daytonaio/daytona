#!/usr/bin/env python3
"""End-to-end check: pointing the Python SDK's download_file_stream at the
hostile upstream from this directory should observe the exact CI failure
pattern (silent truncation: returned bytes < expected, no exception).

Run after starting the hostile upstream:
    cd hack/stream-flake-repro
    GOWORK=off go test -run Hostile -timeout=30s ./...   # NOT needed; just unit tests
    # OR: write a tiny standalone harness — easier:
    GOWORK=off go run ./harness.go &   # see harness.go below

Then:
    python3 sdk_test.py http://127.0.0.1:<port>
"""

from __future__ import annotations
import sys
import io
from python_multipart.multipart import MultipartParser, parse_options_header
import httpx


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: sdk_test.py <hostile_upstream_url>", file=sys.stderr)
        return 2
    url = sys.argv[1]

    truncated = 0
    short = 0
    total = 100
    for _ in range(total):
        with httpx.Client(timeout=5.0) as c:
            try:
                with c.stream("GET", url) as resp:
                    ct = resp.headers.get("Content-Type", "")
                    _, params = parse_options_header(ct)
                    boundary = params.get(b"boundary", b"")
                    file_buf = bytearray()
                    mode = {"current": None}

                    def on_headers_finished():
                        mode["current"] = "file"

                    def on_part_data(data, start, end):
                        if mode["current"] == "file":
                            file_buf.extend(data[start:end])

                    parser = MultipartParser(boundary, callbacks={
                        "on_part_begin": lambda: None,
                        "on_header_field": lambda d, s, e: None,
                        "on_header_value": lambda d, s, e: None,
                        "on_header_end": lambda: None,
                        "on_headers_finished": on_headers_finished,
                        "on_part_data": on_part_data,
                        "on_part_end": lambda: None,
                    })
                    for chunk in resp.iter_bytes(64 * 1024):
                        parser.write(chunk)
                    parser.finalize()
                    # File content-length is 24534 bytes.
                    if len(file_buf) < 24534:
                        truncated += 1
                        short += 24534 - len(file_buf)
            except Exception:
                truncated += 1

    print(f"truncated: {truncated}/{total}, total missing bytes: {short}")
    return 0 if truncated == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
