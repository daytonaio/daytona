# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

"""
Stateful Python REPL Worker for Daytona
- JSON line protocol (stdout)
- Persistent globals across exec calls
- Clean user-only tracebacks
- Graceful SIGINT
"""

import io
import json
import os
import signal
import sys
import traceback
from contextlib import redirect_stderr, redirect_stdout


class REPLWorker:
    def __init__(self):
        self.globals = {
            "__name__": "__main__",
            "__doc__": None,
            "__package__": None,
            "__builtins__": __builtins__,
        }
        self.should_shutdown = False
        self._setup_signals()

    # ---------- IO ----------
    def _emit(self, chunk: dict):
        try:
            json.dump(chunk, sys.__stdout__)
            sys.__stdout__.write("\n")
            sys.__stdout__.flush()
        except Exception as e:
            sys.__stderr__.write(f"Failed to send chunk: {e}\n")

    class _StreamEmitter(io.TextIOBase):
        def __init__(self, worker: "REPLWorker", stream_type: str):
            self.worker = worker
            self.stream_type = stream_type
            self._buffer: list[str] = []

        def writable(self):
            return True

        def write(self, data):  # type: ignore[override]
            if not data:
                return 0
            if not isinstance(data, str):
                data = str(data)
            self._buffer.append(data)
            if "\n" in data or sum(len(chunk) for chunk in self._buffer) >= 1024:
                self.flush()
            return len(data)

        def flush(self):  # type: ignore[override]
            if not self._buffer:
                return
            payload = "".join(self._buffer)
            self._buffer.clear()
            # pylint: disable=protected-access
            self.worker._emit({"type": self.stream_type, "text": payload})

        def isatty(self):
            return False

        @property
        def encoding(self):
            return "utf-8"

        def close(self):  # type: ignore[override]
            self.flush()
            return super().close()

    # ---------- Signals ----------
    def _setup_signals(self):
        def sigint_handler(_signum, _frame):
            raise KeyboardInterrupt("Execution interrupted")

        def sigterm_handler(_signum, _frame):
            self.should_shutdown = True
            raise KeyboardInterrupt("Shutting down")

        signal.signal(signal.SIGINT, sigint_handler)
        signal.signal(signal.SIGTERM, sigterm_handler)

    # ---------- Tracebacks ----------
    def _clean_tb(self, etype, evalue, tb):
        frames = [f for f in traceback.extract_tb(tb) if f.filename == "<string>"]
        if not frames:
            return f"{etype.__name__}: {evalue}\n"
        parts = ["Traceback (most recent call last):\n"]
        for f in frames:
            parts.append(f'  File "<stdin>", line {f.lineno}\n')
            if f.line:
                parts.append(f"    {f.line}\n")
        parts.append(f"{etype.__name__}: {evalue}\n")
        return "".join(parts)

    # ---------- Exec ----------
    def execute_code(self, code: str, envs=None) -> None:
        stdout_emitter = self._StreamEmitter(self, "stdout")
        stderr_emitter = self._StreamEmitter(self, "stderr")
        control_text, error_chunk = "completed", None
        env_snapshot = None

        if envs:
            env_snapshot = {}
            for key, value in envs.items():
                if not isinstance(key, str):
                    key = str(key)
                previous = os.environ.get(key)
                env_snapshot[key] = previous

                if value is None:
                    os.environ.pop(key, None)
                else:
                    os.environ[key] = str(value)

        try:
            with redirect_stdout(stdout_emitter), redirect_stderr(stderr_emitter):
                compiled = compile(code, "<string>", "exec")
                exec(compiled, self.globals)  # pylint: disable=exec-used

        except KeyboardInterrupt:
            control_text = "interrupted"

        except (SystemExit, Exception) as e:
            # SystemExit completes normally
            # Errors are indicated by the error chunk, not control type
            if not isinstance(e, SystemExit):
                error_chunk = {
                    "type": "error",
                    "name": type(e).__name__,
                    "value": str(e),
                    "traceback": self._clean_tb(type(e), e, e.__traceback__),
                }
        finally:
            stdout_emitter.flush()
            stderr_emitter.flush()
            if env_snapshot is not None:
                for key, previous in env_snapshot.items():
                    if previous is None:
                        os.environ.pop(key, None)
                    else:
                        os.environ[key] = previous
            if error_chunk:
                self._emit(error_chunk)
            self._emit({"type": "control", "text": control_text})

    # ---------- Protocol ----------
    def handle_command(self, line: str) -> None:
        try:
            msg = json.loads(line)
            envs = msg.get("envs")
            if envs is not None and not isinstance(envs, dict):
                raise ValueError("envs must be an object")
            self.execute_code(msg.get("code", ""), envs)
        except json.JSONDecodeError as e:
            self._emit({"type": "error", "name": "JSONDecodeError", "value": str(e), "traceback": ""})
        except Exception as e:
            self._emit({"type": "error", "name": type(e).__name__, "value": str(e), "traceback": ""})

    # ---------- Main loop ----------
    def run(self):
        while True:
            if self.should_shutdown:
                break

            try:
                line = sys.stdin.readline()
                if not line:
                    break  # EOF
                line = line.strip()
                if not line:
                    continue
                self.handle_command(line)
            except KeyboardInterrupt:
                continue
            except Exception as e:
                sys.__stderr__.write(f"Fatal error in main loop: {e}\n")
                break


if __name__ == "__main__":
    REPLWorker().run()
