#!/usr/bin/env python3
"""Self-contained reproducer: spins up a hostile multipart upstream that
RSTs the connection after writing the body (closing boundary staged but
unflushed). Then runs Python SDK-style download_file_stream against it
and prints how many requests silently truncated.

Run:
    poetry run python hack/stream-flake-repro/sdk_repro.py
"""

from __future__ import annotations

import socket
import threading
import time
from concurrent.futures import ThreadPoolExecutor
from contextlib import closing

import httpx
from python_multipart.multipart import MultipartParser, parse_options_header


UNIT = b"progress-check-8993686a88ec4238b758d71cd6077b01"
PER_FILE = (24 * 1024 // len(UNIT)) * len(UNIT)
BOUNDARY = b"DAYTONA-FILE-BOUNDARY"


def serve_conn(conn: socket.socket) -> None:
    try:
        conn.settimeout(5.0)
        # Read request line + headers
        buf = b""
        while b"\r\n\r\n" not in buf:
            data = conn.recv(4096)
            if not data:
                return
            buf += data

        send = conn.sendall
        send(
            b"HTTP/1.1 200 OK\r\n"
            b"Content-Type: multipart/form-data; boundary=" + BOUNDARY + b"\r\n"
            b"Transfer-Encoding: chunked\r\n"
            b"\r\n"
        )

        def chunk(p: bytes) -> None:
            send(f"{len(p):x}\r\n".encode() + p + b"\r\n")

        hdr = (
            b"--" + BOUNDARY + b"\r\n"
            b"Content-Type: application/octet-stream\r\n"
            b'Content-Disposition: form-data; name="file"; filename="f.bin"\r\n'
            b"Content-Length: " + str(PER_FILE).encode() + b"\r\n\r\n"
        )
        chunk(hdr)

        body = (UNIT * ((PER_FILE // len(UNIT)) + 1))[:PER_FILE]
        # Stream in 4KB chunks like the daemon.
        for i in range(0, PER_FILE, 4096):
            chunk(body[i : i + 4096])

        # Stage the closing boundary + chunked terminator into the kernel
        # buffer, then RST: set SO_LINGER to {1, 0} which forces RST on close.
        closing = b"\r\n--" + BOUNDARY + b"--\r\n"
        # Send the closing boundary header for chunked, but don't fully flush.
        try:
            send(f"{len(closing):x}\r\n".encode() + closing + b"\r\n")
            send(b"0\r\n\r\n")
        except OSError:
            pass
        # Force RST instead of FIN so kernel discards send buffer.
        import struct as _s
        conn.setsockopt(
            socket.SOL_SOCKET, socket.SO_LINGER, _s.pack("ii", 1, 0)
        )
    finally:
        conn.close()


def run_server(stop_evt: threading.Event) -> int:
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.bind(("127.0.0.1", 0))
    s.listen(64)
    port = s.getsockname()[1]

    def loop() -> None:
        while not stop_evt.is_set():
            try:
                s.settimeout(0.2)
                conn, _ = s.accept()
            except socket.timeout:
                continue
            except OSError:
                return
            threading.Thread(target=serve_conn, args=(conn,), daemon=True).start()
        s.close()

    t = threading.Thread(target=loop, daemon=True)
    t.start()
    return port


def download_once(url: str) -> tuple[int, str]:
    try:
        with httpx.Client(timeout=5.0) as c:
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

                parser = MultipartParser(
                    boundary,
                    callbacks={
                        "on_part_begin": lambda: None,
                        "on_header_field": lambda d, s, e: None,
                        "on_header_value": lambda d, s, e: None,
                        "on_header_end": lambda: None,
                        "on_headers_finished": on_headers_finished,
                        "on_part_data": on_part_data,
                        "on_part_end": lambda: None,
                    },
                )
                try:
                    for chunk in resp.iter_bytes(64 * 1024):
                        parser.write(chunk)
                    parser.finalize()
                except Exception as e:
                    return len(file_buf), f"stream-error:{type(e).__name__}"
                return len(file_buf), "ok"
    except Exception as e:
        return 0, f"client-error:{type(e).__name__}"


def main() -> int:
    import sys

    if len(sys.argv) >= 2:
        url = sys.argv[1]
        stop = None
        print(f"using external upstream: {url}")
    else:
        stop = threading.Event()
        port = run_server(stop)
        time.sleep(0.1)
        url = f"http://127.0.0.1:{port}"
        print(f"hostile upstream (built-in python server): {url}")
    print(f"expected file size: {PER_FILE} bytes")

    total = 200
    silent_truncations = 0
    exact = 0
    client_errors = 0
    sizes = []
    with ThreadPoolExecutor(max_workers=32) as ex:
        for got, status in ex.map(lambda _: download_once(url), range(total)):
            sizes.append(got)
            if status == "ok" and got == PER_FILE:
                exact += 1
            elif status == "ok":
                # Stream parsed cleanly, but yielded fewer bytes than expected.
                # This is the silent truncation we are trying to demonstrate.
                silent_truncations += 1
            else:
                client_errors += 1

    if stop is not None:
        stop.set()
    print(f"results over {total} requests:")
    print(f"  exact (got {PER_FILE} bytes, no error): {exact}")
    print(f"  silent truncation (parser fine, got < expected): {silent_truncations}")
    print(f"  client errors / exceptions:            {client_errors}")
    if silent_truncations:
        unique = sorted(set(s for s in sizes if 0 < s < PER_FILE))
        print(f"  silent-truncated sizes seen (up to 5): {unique[:5]}")
    return 0 if silent_truncations == 0 and client_errors == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
