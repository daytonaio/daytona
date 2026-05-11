"""Stateful Python REPL Worker for Daytona sessions.

JSON line protocol:
- Stdin: one JSON object per line, of the form
  {"id": "...", "code": "...", "envs": {...}}
- Stdout: one JSON object per line, of types:
  {"type": "stdout"|"stderr", "text": "..."}
  {"type": "error", "name": "...", "value": "...", "traceback": "..."}
  {"type": "display", "formats": [...], "data": {"<mime>": "..."}}
  {"type": "control", "text": "completed"|"interrupted"}

State persists across exec calls (variables, imports, top-level declarations).
"""

# This file is a REPL host. exec/eval and on-demand imports are part of the
# design (user code is the *input*, not a static body to lint), so we silence
# the related pylint checks file-wide rather than per-call.
# pylint: disable=exec-used,eval-used,import-outside-toplevel

import base64
import io
import json
import os
import signal
import sys
import traceback
import uuid
from contextlib import redirect_stderr, redirect_stdout

# Configure matplotlib up-front so the Agg backend is selected before any
# user code can import matplotlib and force-pick a different backend.
try:  # pragma: no cover - import availability depends on the runtime image
    import matplotlib  # type: ignore[import]

    matplotlib.use("Agg", force=True)
    import matplotlib.pyplot as _plt  # noqa: F401  # pylint: disable=unused-import
except Exception:  # noqa: BLE001
    matplotlib = None  # type: ignore[assignment]


# Max bytes for any single text payload written into an output frame. Capping
# here (the emit side) means no single stdout/stderr/error/display frame can blow
# past the reader's bounded line size on the Go side, regardless of how much the
# user code prints in one shot. Kept in sync with MAX_CHUNK_BYTES in repl_host.js.
MAX_CHUNK_BYTES = 1 << 20  # 1 MiB

# Max bytes for a whole serialized output line (one JSON frame + newline). Kept
# safely below the Go reader's maxWorkerLineBytes (8 MiB) cap so even a frame with
# many text/data fields — each individually within MAX_CHUNK_BYTES — can never sum
# past the reader limit and trip its oversized-line recovery. Kept in sync with
# MAX_LINE_BYTES in repl_host.js.
MAX_LINE_BYTES = 4 << 20  # 4 MiB


def _truncate_text(s: str) -> str:
    """Cap a text payload at MAX_CHUNK_BYTES, appending an omitted-bytes marker.

    Slices on a UTF-8 byte boundary so we never emit a split multibyte sequence;
    returns the input untouched when already within budget (the common case).
    """
    if not isinstance(s, str):
        return s
    encoded = s.encode("utf-8")
    if len(encoded) <= MAX_CHUNK_BYTES:
        return s
    kept = encoded[:MAX_CHUNK_BYTES].decode("utf-8", "ignore")
    omitted = len(encoded) - len(kept.encode("utf-8"))
    return f"{kept}…[output truncated: {omitted} bytes omitted]"


def _shrink_frame(chunk: dict) -> dict:
    """Replace a frame's text/data fields with a truncation marker.

    Last-resort fallback for _emit when an entire serialized frame would exceed
    MAX_LINE_BYTES (e.g. many fields each within MAX_CHUNK_BYTES but summing past
    the reader cap). Drops the bulky payloads so the line is small while keeping
    the frame's routing fields (type/name/control text) intact. Mutates and
    returns the same dict.
    """
    marker = "…[output truncated: frame exceeded line limit]"
    # `name` is user-controlled (e.g. type(e).__name__), so a pathological
    # exception class name must not survive into the shrunk frame.
    for key in ("text", "value", "traceback", "name"):
        if isinstance(chunk.get(key), str):
            chunk[key] = marker
    if isinstance(chunk.get("data"), dict):
        chunk["data"] = {mime: marker for mime in chunk["data"]}
    return chunk


class REPLWorker:
    def __init__(self) -> None:
        self.globals = self._fresh_globals()
        self.should_shutdown = False
        self._setup_signals()
        self._patch_pyplot_show()

    def _fresh_globals(self) -> dict:
        # The trio of dunders + __builtins__ is what `exec()` expects to find
        # in the globals dict; rebuilding from scratch gives transient-context
        # callers true one-shot semantics without re-spawning the interpreter.
        # `bash` is injected as a builtin so user code can shell out to the
        # daemon's virtual just-bash shell (see _bash / the hostcall bridge).
        return {
            "__name__": "__main__",
            "__doc__": None,
            "__package__": None,
            "__builtins__": __builtins__,
            "bash": self._bash,
        }

    # ---------- IO ----------
    def _emit(self, chunk: dict) -> None:
        # Cap oversized text payloads (stdout/stderr/error/display) per field so
        # no single frame can be unbounded — see MAX_CHUNK_BYTES. Applied per
        # text-bearing field rather than to the whole JSON line so the structure
        # stays intact and the omitted-bytes marker lands inside the field.
        for key in ("text", "value", "traceback"):
            if isinstance(chunk.get(key), str):
                chunk[key] = _truncate_text(chunk[key])
        data = chunk.get("data")
        if isinstance(data, dict):
            for mime, payload in data.items():
                if isinstance(payload, str):
                    data[mime] = _truncate_text(payload)
        try:
            line = json.dumps(chunk)
            # Whole-line guard: per-field truncation above bounds each text/data
            # field, but a frame carrying many such fields could still serialize
            # past the Go reader's line cap. If so, drop the oversized text/data
            # fields (replacing them with a marker) so the emitted line is always
            # under the reader limit and the host's oversized-line recovery stays
            # unreachable for well-behaved output.
            if len(line.encode("utf-8")) > MAX_LINE_BYTES:
                line = json.dumps(_shrink_frame(chunk))
                # _shrink_frame only drops the known bulky fields; a frame with
                # other large fields could still exceed the cap. Re-check and, if
                # still oversized, fall back to a minimal control frame that
                # preserves only the routing fields so the emitted line is always
                # under the reader cap.
                if len(line.encode("utf-8")) > MAX_LINE_BYTES:
                    line = json.dumps(
                        {
                            "type": chunk.get("type", "control"),
                            "text": "…[output truncated: frame exceeded line limit]",
                        }
                    )
            sys.__stdout__.write(line)
            sys.__stdout__.write("\n")
            sys.__stdout__.flush()
        except Exception as e:  # noqa: BLE001
            sys.__stderr__.write(f"Failed to send chunk: {e}\n")

    class _StreamEmitter(io.TextIOBase):
        def __init__(self, worker: "REPLWorker", stream_type: str) -> None:
            self.worker = worker
            self.stream_type = stream_type
            self._buffer: list[str] = []

        def writable(self) -> bool:  # type: ignore[override]
            return True

        def write(self, data) -> int:  # type: ignore[override]
            if not data:
                return 0
            if not isinstance(data, str):
                data = str(data)
            self._buffer.append(data)
            if "\n" in data or sum(len(c) for c in self._buffer) >= 1024:
                self.flush()
            return len(data)

        def flush(self) -> None:  # type: ignore[override]
            if not self._buffer:
                return
            payload = "".join(self._buffer)
            self._buffer.clear()
            self.worker._emit({"type": self.stream_type, "text": payload})

        def isatty(self) -> bool:  # type: ignore[override]
            return False

        @property
        def encoding(self) -> str:
            return "utf-8"

        def close(self) -> None:  # type: ignore[override]
            self.flush()
            return super().close()

    # ---------- bash() bridge ----------
    def _bash(self, cmd, env=None) -> "BashResult":
        """Run a command in the daemon's virtual just-bash shell.

        Emits a hostcall frame on stdout and blocks reading stdin for the
        correlated result. This is safe because the daemon never interleaves a
        new exec command while one is in flight, so during a bash() call stdin
        carries only this call's reply. Raises RuntimeError if the bridge is
        unavailable or the command could not be dispatched; a non-zero command
        exit is NOT an error — inspect the returned exit_code.
        """
        call_id = uuid.uuid4().hex
        payload = {"type": "hostcall", "id": call_id, "method": "bash", "cmd": str(cmd)}
        if env:
            payload["env"] = {str(k): str(v) for k, v in env.items()}
        # Write directly (not via _emit, which truncates text fields) so the
        # command and correlation id are never mangled.
        sys.__stdout__.write(json.dumps(payload))
        sys.__stdout__.write("\n")
        sys.__stdout__.flush()

        while True:
            line = sys.stdin.readline()
            if not line:
                raise RuntimeError("bash bridge: connection closed by daemon")
            line = line.strip()
            if not line:
                continue
            try:
                msg = json.loads(line)
            except json.JSONDecodeError:
                continue
            if not isinstance(msg, dict) or msg.get("id") != call_id:
                continue
            mtype = msg.get("type")
            if mtype == "hostcall_error":
                raise RuntimeError("bash: " + str(msg.get("message", "bridge error")))
            if mtype == "hostcall_result":
                return BashResult(
                    str(msg.get("stdout", "")),
                    str(msg.get("stderr", "")),
                    int(msg.get("exitCode", 0)),
                )

    # ---------- Signals ----------
    def _setup_signals(self) -> None:
        def sigint_handler(_signum, _frame):
            raise KeyboardInterrupt("Execution interrupted")

        def sigterm_handler(_signum, _frame):
            self.should_shutdown = True
            raise KeyboardInterrupt("Shutting down")

        signal.signal(signal.SIGINT, sigint_handler)
        signal.signal(signal.SIGTERM, sigterm_handler)

    # ---------- Tracebacks ----------
    def _clean_tb(self, etype, evalue, tb) -> str:
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

    # ---------- Display hook ----------
    def _patch_pyplot_show(self) -> None:
        if matplotlib is None:
            return
        try:
            import matplotlib.pyplot as plt  # type: ignore[import]
        except Exception:  # noqa: BLE001
            return

        worker = self

        def show(*_args, **_kwargs):
            try:
                fig = plt.gcf()
                if not fig.axes:
                    return
                buf = io.BytesIO()
                fig.savefig(buf, format="png", bbox_inches="tight")
                payload = base64.b64encode(buf.getvalue()).decode("ascii")
                worker._emit(
                    {
                        "type": "display",
                        "formats": ["image/png"],
                        "data": {"image/png": payload},
                    }
                )
                plt.close(fig)
            except Exception as e:  # noqa: BLE001
                sys.__stderr__.write(f"matplotlib show hook failed: {e}\n")

        plt.show = show  # type: ignore[assignment]

    def _emit_display_for_value(self, value) -> None:
        if value is None:
            return

        formats: list[str] = []
        data: dict[str, str] = {}

        # Capture richest representation first.
        for attr, mime, encoder in (
            ("_repr_png_", "image/png", _b64_or_none),
            ("_repr_jpeg_", "image/jpeg", _b64_or_none),
            ("_repr_svg_", "image/svg+xml", _str_or_none),
            ("_repr_html_", "text/html", _str_or_none),
            ("_repr_json_", "application/json", _json_or_none),
        ):
            if hasattr(value, attr):
                try:
                    raw = getattr(value, attr)()
                except Exception:  # noqa: BLE001
                    continue
                payload = encoder(raw)
                if payload is not None:
                    formats.append(mime)
                    data[mime] = payload

        # Fall back to JSON for plain dicts/lists (matches the TS engine), then
        # text/plain repr for everything else.
        if not formats and isinstance(value, (dict, list, tuple)):
            try:
                json_payload = json.dumps(value, default=str)
                formats.append("application/json")
                data["application/json"] = json_payload
            except Exception:  # noqa: BLE001
                pass
        if not formats:
            formats.append("text/plain")
            data["text/plain"] = repr(value)

        self._emit({"type": "display", "formats": formats, "data": data})

    # ---------- Exec ----------
    def execute_code(self, code: str, envs=None, reset: bool = False) -> None:
        stdout_emitter = self._StreamEmitter(self, "stdout")
        stderr_emitter = self._StreamEmitter(self, "stderr")
        control_text, error_chunk = "completed", None
        env_snapshot = None

        if reset:
            # Drop user-imported modules + variables, keep the interpreter alive.
            self.globals = self._fresh_globals()

        if envs:
            env_snapshot = {}
            for key, value in envs.items():
                if not isinstance(key, str):
                    key = str(key)
                env_snapshot[key] = os.environ.get(key)
                if value is None:
                    os.environ.pop(key, None)
                else:
                    os.environ[key] = str(value)

        last_value = None
        try:
            with redirect_stdout(stdout_emitter), redirect_stderr(stderr_emitter):
                last_value = _exec_with_last_expr(code, self.globals)
        except KeyboardInterrupt:
            control_text = "interrupted"
        except (SystemExit, Exception) as e:  # noqa: BLE001
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
            elif last_value is not None:
                self._emit_display_for_value(last_value)
            self._emit({"type": "control", "text": control_text})

    # ---------- Protocol ----------
    def handle_command(self, line: str) -> None:
        try:
            msg = json.loads(line)
            # A bash() hostcall reply can land in the main loop if the exec that
            # issued it was interrupted before consuming it. It's not a command —
            # discard it rather than running it as empty code.
            if isinstance(msg, dict) and str(msg.get("type", "")).startswith("hostcall"):
                return
            envs = msg.get("envs")
            if envs is not None and not isinstance(envs, dict):
                raise ValueError("envs must be an object")
            self.execute_code(msg.get("code", ""), envs, bool(msg.get("reset")))
        except json.JSONDecodeError as e:
            self._emit({"type": "error", "name": "JSONDecodeError", "value": str(e), "traceback": ""})
        except Exception as e:  # noqa: BLE001
            self._emit({"type": "error", "name": type(e).__name__, "value": str(e), "traceback": ""})

    # ---------- Main loop ----------
    def run(self) -> None:
        while not self.should_shutdown:
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
            except Exception as e:  # noqa: BLE001
                sys.__stderr__.write(f"Fatal error in main loop: {e}\n")
                break


class BashResult:
    """Result of a bash() bridge call: captured stdout/stderr and the exit code.

    repr() is kept compact so a trailing `bash(...)` expression in a cell renders
    a readable one-liner via the daemon's display hook.
    """

    __slots__ = ("stdout", "stderr", "exit_code")

    def __init__(self, stdout: str, stderr: str, exit_code: int) -> None:
        self.stdout = stdout
        self.stderr = stderr
        self.exit_code = exit_code

    def __repr__(self) -> str:
        return f"BashResult(exit_code={self.exit_code}, stdout={self.stdout!r}, stderr={self.stderr!r})"


def _exec_with_last_expr(code: str, globals_dict: dict):
    """Compile + exec ``code``; return the value of the trailing expression if any.

    Mirrors the Pyodide convention of "last top-level expression becomes the cell
    value", which the daemon then turns into a display chunk for rich-output rendering.
    """
    import ast

    try:
        tree = ast.parse(code, mode="exec")
    except SyntaxError:
        # Let exec() raise the same SyntaxError so the worker's except branch handles it.
        compiled = compile(code, "<string>", "exec")
        exec(compiled, globals_dict)  # noqa: S102
        return None

    if not tree.body:
        return None

    last = tree.body[-1]
    if isinstance(last, ast.Expr):
        body = ast.Module(body=tree.body[:-1], type_ignores=[])
        tail = ast.Expression(body=last.value)
        compiled_body = compile(body, "<string>", "exec")
        compiled_tail = compile(tail, "<string>", "eval")
        exec(compiled_body, globals_dict)  # noqa: S102
        return eval(compiled_tail, globals_dict)  # noqa: S307

    compiled = compile(tree, "<string>", "exec")
    exec(compiled, globals_dict)  # noqa: S102
    return None


def _b64_or_none(value):
    if value is None:
        return None
    if isinstance(value, str):
        return value
    if isinstance(value, (bytes, bytearray)):
        return base64.b64encode(value).decode("ascii")
    return None


def _str_or_none(value):
    if value is None:
        return None
    if isinstance(value, (bytes, bytearray)):
        return value.decode("utf-8", "replace")
    return str(value)


def _json_or_none(value):
    if value is None:
        return None
    if isinstance(value, str):
        return value
    try:
        return json.dumps(value, default=str)
    except Exception:  # noqa: BLE001
        return None


if __name__ == "__main__":
    REPLWorker().run()
