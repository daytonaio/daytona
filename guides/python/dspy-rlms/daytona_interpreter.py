# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import inspect
import json
import keyword
import logging
import math
import threading
import time
import urllib.error
import urllib.request
import uuid
from concurrent.futures import Future, ThreadPoolExecutor
from typing import Any, Callable

from dspy.primitives.code_interpreter import CodeInterpreterError, FinalOutput

logger = logging.getLogger(__name__)


# =============================================================================
# DaytonaInterpreter - Implements CodeInterpreter protocol using Daytona
# =============================================================================

# Python setup code injected into the Daytona context.
# Defines FinalOutput exception and SUBMIT function for early termination.
_SETUP_CODE = '''
import json as _json

class FinalOutput(BaseException):
    """Control-flow exception to signal completion (like StopIteration)."""
    pass

_FINAL_OUTPUT_MARKER = "__DSPY_FINAL_OUTPUT__"

def SUBMIT(output=None, **kwargs):
    """Submit final output and terminate execution."""
    if kwargs:
        result = kwargs
    elif output is not None:
        result = {"output": output}
    else:
        result = {}
    # Encode the result so we can detect it in stdout
    print(f"{_FINAL_OUTPUT_MARKER}{_json.dumps(result)}{_FINAL_OUTPUT_MARKER}")
    raise FinalOutput(result)
'''


# Flask broker server code to be uploaded and run in the sandbox.
# This server mediates tool calls between sandbox code and the host.
_BROKER_SERVER_CODE = '''
"""Broker server for mediating tool calls between sandbox and host."""
import json
import threading
import time
import uuid
from flask import Flask, request, jsonify

app = Flask(__name__)

# Thread-safe storage for pending requests and results
_lock = threading.Lock()
_pending_requests = {}  # id -> {tool_name, args, kwargs, claimed, claimed_at, lease_token}
_results = {}  # id -> result

@app.route("/health", methods=["GET"])
def health():
    """Health check endpoint."""
    return jsonify({"status": "ok"})

@app.route("/tool_call", methods=["POST"])
def tool_call():
    """Sandbox code calls this to request a tool execution.

    Blocks until the host provides a result (polling with timeout).
    """
    data = request.json
    call_id = data.get("id")
    tool_name = data.get("tool_name")
    args = data.get("args", [])
    kwargs = data.get("kwargs", {})

    # Register the pending request
    with _lock:
        _pending_requests[call_id] = {
            "tool_name": tool_name,
            "args": args,
            "kwargs": kwargs,
            "claimed": False,
            "claimed_at": None,
            "lease_token": None,
        }

    # Poll for result (50ms interval, 2 minute timeout)
    timeout = 120.0
    interval = 0.05
    elapsed = 0.0

    while elapsed < timeout:
        with _lock:
            if call_id in _results:
                result = _results.pop(call_id)
                _pending_requests.pop(call_id, None)
                return jsonify({"result": result})
        time.sleep(interval)
        elapsed += interval

    # Timeout - clean up and return error
    with _lock:
        _pending_requests.pop(call_id, None)

    return jsonify({"error": "Tool call timeout"}), 504

@app.route("/pending", methods=["GET"])
def get_pending():
    """Host polls this to get pending tool requests.

    Returns up to `max` requests. Claims use short leases so abandoned claims
    can be recovered if a worker crashes before posting a result.
    """
    try:
        max_items = int(request.args.get("max", "1"))
    except ValueError:
        max_items = 1
    if max_items < 1:
        max_items = 1
    try:
        lease_seconds = float(request.args.get("lease_seconds", "60"))
    except ValueError:
        lease_seconds = 30.0
    if lease_seconds < 1:
        lease_seconds = 1.0

    requests_out = []
    with _lock:
        now = time.time()
        # Return up to `max` pending requests that don't have a result yet.
        for call_id, req in _pending_requests.items():
            if len(requests_out) >= max_items:
                break
            if call_id in _results:
                continue
            claimed_at = req.get("claimed_at")
            if req.get("claimed") and isinstance(claimed_at, (int, float)):
                if now - claimed_at < lease_seconds:
                    continue
            claim_token = str(uuid.uuid4())
            req["claimed"] = True
            req["claimed_at"] = now
            req["lease_token"] = claim_token
            requests_out.append({
                "id": call_id,
                "tool_name": req["tool_name"],
                "args": req["args"],
                "kwargs": req["kwargs"],
                "claim_token": claim_token,
            })
    return jsonify({"requests": requests_out})

@app.route("/result/<call_id>", methods=["POST"])
def post_result(call_id):
    """Host calls this to provide tool execution result for a valid claim."""
    data = request.json
    result = data.get("result")
    claim_token = data.get("claim_token")

    with _lock:
        req = _pending_requests.get(call_id)
        if req is None:
            return jsonify({"error": "Unknown or expired call_id"}), 404
        expected_token = req.get("lease_token")
        if not expected_token or claim_token != expected_token:
            return jsonify({"error": "Stale or invalid claim token"}), 409
        _results[call_id] = result
        # Invalidate the lease token so a late duplicate post is rejected.
        req["lease_token"] = None

    return jsonify({"status": "ok"})

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=3000, threaded=True)
'''


# Template for generating tool wrapper functions in the sandbox
_TOOL_WRAPPER_TEMPLATE = '''
def {tool_name}({signature}):
    """Wrapper for {tool_name} tool - calls host via broker."""
    import json as _json
    import urllib.request as _urllib_request
    import uuid as _uuid

    call_id = str(_uuid.uuid4())
    payload = _json.dumps({{
        "id": call_id,
        "tool_name": "{tool_name}",
        "args": [{args_list}],
        "kwargs": {{{kwargs_dict}}},
    }}).encode("utf-8")

    req = _urllib_request.Request(
        "http://localhost:3000/tool_call",
        data=payload,
        headers={{"Content-Type": "application/json"}},
        method="POST",
    )

    with _urllib_request.urlopen(req, timeout=130) as resp:
        data = _json.loads(resp.read().decode("utf-8"))

    if "error" in data:
        raise RuntimeError(f"Tool call failed: {{data['error']}}")

    result = data.get("result")
    # Try to parse JSON result, otherwise return as-is
    if isinstance(result, str):
        try:
            return _json.loads(result)
        except (ValueError, _json.JSONDecodeError):
            return result
    return result
'''


class DaytonaInterpreter:
    """Remote interpreter using Daytona sandbox for Python execution.

    Implements the CodeInterpreter protocol for secure code execution in a
    Daytona cloud sandbox. The sandbox provides isolated execution with
    persistent state across calls.

    Prerequisites:
        - Daytona SDK: pip install daytona
        - DAYTONA_API_KEY environment variable set

    Example:
        ```python
        # Basic execution
        with DaytonaInterpreter() as interp:
            result = interp.execute("print(1 + 2)")  # Returns "3"
        ```
    """

    def __init__(
        self,
        tools: dict[str, Callable[..., str]] | None = None,
        output_fields: list[dict] | None = None,
        sandbox_params: dict[str, Any] | None = None,
        max_concurrent_tool_calls: int = 32,
        tool_claim_lease_seconds: float = 60.0,
        max_retries: int = 3,
        retry_delay: float = 2.0,
    ) -> None:
        """
        Args:
            tools: Dictionary mapping tool names to callable functions.
            output_fields: List of output field definitions for typed SUBMIT signature.
            sandbox_params: Optional parameters to pass to Daytona sandbox creation.
            max_concurrent_tool_calls: Maximum host-side parallel tool workers.
            tool_claim_lease_seconds: Claim lease duration used by broker recovery.
            max_retries: Max attempts for transient errors (e.g. WebSocket drops).
            retry_delay: Seconds to wait between retries.
        """
        self.tools = dict(tools) if tools else {}
        self.output_fields = output_fields
        self.sandbox_params = sandbox_params or {}
        if max_concurrent_tool_calls < 1:
            raise ValueError("max_concurrent_tool_calls must be >= 1")
        if tool_claim_lease_seconds < 1:
            raise ValueError("tool_claim_lease_seconds must be >= 1")
        if max_retries < 1:
            raise ValueError("max_retries must be >= 1")
        self.max_concurrent_tool_calls = max_concurrent_tool_calls
        self.tool_claim_lease_seconds = float(tool_claim_lease_seconds)
        self._max_retries = max_retries
        self._retry_delay = float(retry_delay)

        self._daytona = None
        self._sandbox = None
        self._context = None
        self._initialized = False
        self._session_execute_request = None

        # Broker-related state
        self._broker_url = None  # Preview URL for broker (e.g., https://xxx.preview.daytona.work)
        self._broker_token = None  # Auth token for preview URL
        self._broker_session_id = None  # Session ID for the broker process

        # Track which tools have been injected into the sandbox so we can
        # detect when RLM (or the user) adds new ones after start().
        self._injected_tools: set[str] = set()

        # Track whether typed SUBMIT has been registered, so we can detect
        # when RLM sets output_fields after start().
        self._submit_registered: bool = False

    @property
    def _final_output_marker(self) -> str:
        return "__DSPY_FINAL_OUTPUT__"

    def start(self) -> None:
        """Initialize the Daytona sandbox and interpreter context."""
        if self._initialized:
            return

        try:
            from daytona import Daytona, SessionExecuteRequest  # pylint: disable=import-outside-toplevel

            self._session_execute_request = SessionExecuteRequest
        except ImportError as exc:
            raise CodeInterpreterError(
                "Daytona not installed. Install with: pip install daytona\n"
                "Also ensure DAYTONA_API_KEY environment variable is set."
            ) from exc

        logger.info("Creating Daytona sandbox...")
        self._daytona = Daytona()
        self._sandbox = self._daytona.create(**self.sandbox_params)
        logger.info("Sandbox created: %s", self._sandbox.id)

        # Create isolated interpreter context
        self._context = self._sandbox.code_interpreter.create_context()
        logger.info("Interpreter context created: %s", self._context.id)

        # Run setup code to define SUBMIT and FinalOutput
        self._run_setup_code()

        # Start broker and inject tools if we have any at init time.
        # Tools may also be added later (e.g. RLM injects llm_query after
        # start), so execute() checks for new tools on every call.
        if self.tools:
            self._start_broker()
            self._inject_tool_wrappers()

        self._initialized = True

    def _run_setup_code(self) -> None:
        """Run initialization code in the sandbox context."""
        result = self._sandbox.code_interpreter.run_code(
            _SETUP_CODE,
            context=self._context,
        )
        if result.error:
            raise CodeInterpreterError(f"Failed to initialize sandbox: {result.error.value}")

        # If we have typed output fields, override SUBMIT with typed signature
        self._register_typed_submit()

    def _register_typed_submit(self) -> None:
        """Register typed SUBMIT if output_fields are set and not yet registered."""
        if not self.output_fields or self._submit_registered:
            return
        submit_code = self._generate_typed_submit()
        result = self._sandbox.code_interpreter.run_code(
            submit_code,
            context=self._context,
        )
        if result.error:
            raise CodeInterpreterError(f"Failed to register SUBMIT: {result.error.value}")
        self._submit_registered = True
        logger.info("Registered typed SUBMIT with fields: %s", [f["name"] for f in self.output_fields])

    def _generate_typed_submit(self) -> str:
        """Generate SUBMIT function with typed output field signature."""
        if not self.output_fields:
            return ""

        sig_parts = []
        for field in self.output_fields:
            part = field["name"]
            if "type" in field:
                part += f": {field['type']}"
            sig_parts.append(part)

        dict_parts = [f'"{f["name"]}": {f["name"]}' for f in self.output_fields]

        return f'''
def SUBMIT({", ".join(sig_parts)}):
    """Submit final output and terminate execution."""
    result = {{{", ".join(dict_parts)}}}
    print(f"{{_FINAL_OUTPUT_MARKER}}{{_json.dumps(result)}}{{_FINAL_OUTPUT_MARKER}}")
    raise FinalOutput(result)
'''

    def _start_broker(self) -> None:
        """Upload and start the broker server in the sandbox."""
        if not self.tools:
            return  # No tools, no broker needed

        logger.info("Starting broker server in sandbox...")

        # Write broker server code to sandbox using code_interpreter
        write_code = f"""
import os
broker_code = {repr(_BROKER_SERVER_CODE)}
with open("/home/daytona/broker_server.py", "w") as f:
    f.write(broker_code)
print("Broker server code written")
"""
        result = self._sandbox.code_interpreter.run_code(write_code, context=self._context)
        if result.error:
            raise CodeInterpreterError(f"Failed to write broker server: {result.error.value}")

        # Start broker server as background process
        self._broker_session_id = f"broker-{uuid.uuid4().hex[:8]}"
        self._sandbox.process.create_session(self._broker_session_id)
        self._sandbox.process.execute_session_command(
            self._broker_session_id,
            self._session_execute_request(
                command="cd /home/daytona && python broker_server.py",
                run_async=True,
            ),
        )

        # Get preview link for port 3000
        preview = self._sandbox.get_preview_link(3000)
        self._broker_url = preview.url
        self._broker_token = preview.token
        logger.info("Broker preview URL: %s", self._broker_url)

        # Wait for broker to be ready
        self._wait_for_broker_health()

    def _wait_for_broker_health(self, timeout: float = 30.0) -> None:
        """Poll broker health endpoint until ready."""
        url = f"{self._broker_url}/health"
        headers = {"x-daytona-preview-token": self._broker_token}

        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                req = urllib.request.Request(url, headers=headers, method="GET")
                with urllib.request.urlopen(req, timeout=5) as resp:
                    if resp.status == 200:
                        logger.info("Broker server is ready")
                        return
            except (urllib.error.URLError, urllib.error.HTTPError, OSError):
                pass
            time.sleep(0.5)

        raise CodeInterpreterError("Broker server failed to start within timeout")

    def _inject_tool_wrappers(self) -> None:
        """Inject tool wrapper functions into the sandbox context.

        Only injects tools that haven't been injected yet, so this is safe
        to call repeatedly (e.g. after RLM adds llm_query / llm_query_batched).
        """
        if not self.tools:
            return

        for tool_name, tool_func in self.tools.items():
            if tool_name in self._injected_tools:
                continue
            if not tool_name.isidentifier() or keyword.iskeyword(tool_name):
                raise CodeInterpreterError(f"Invalid tool name: '{tool_name}'")
            wrapper_code = self._generate_tool_wrapper(tool_name, tool_func)
            result = self._sandbox.code_interpreter.run_code(
                wrapper_code,
                context=self._context,
            )
            if result.error:
                raise CodeInterpreterError(f"Failed to inject tool '{tool_name}': {result.error.value}")
            self._injected_tools.add(tool_name)
            logger.info("Injected tool wrapper: %s", tool_name)

    def _generate_tool_wrapper(self, tool_name: str, tool_func: Callable) -> str:
        """Generate a wrapper function for a tool that calls the broker."""
        sig = inspect.signature(tool_func)
        params = list(sig.parameters.values())

        # Build signature parts
        sig_parts = []
        args_list = []
        kwargs_dict = []
        last_positional_only_idx = max(
            (i for i, param in enumerate(params) if param.kind == inspect.Parameter.POSITIONAL_ONLY),
            default=-1,
        )
        added_kw_only_separator = False

        def format_param(param: inspect.Parameter) -> str:
            if param.default is inspect.Parameter.empty:
                return param.name
            return f"{param.name}={repr(param.default)}"

        for i, param in enumerate(params):
            if param.kind == inspect.Parameter.POSITIONAL_ONLY:
                sig_parts.append(format_param(param))
                args_list.append(param.name)
                if i == last_positional_only_idx:
                    sig_parts.append("/")
            elif param.kind == inspect.Parameter.POSITIONAL_OR_KEYWORD:
                sig_parts.append(format_param(param))
                if param.default is inspect.Parameter.empty:
                    args_list.append(param.name)
                else:
                    kwargs_dict.append(f'"{param.name}": {param.name}')
            elif param.kind == inspect.Parameter.VAR_POSITIONAL:
                sig_parts.append(f"*{param.name}")
                # *args gets passed as args list
                args_list.append(f"*{param.name}")
                added_kw_only_separator = True
            elif param.kind == inspect.Parameter.KEYWORD_ONLY:
                if not added_kw_only_separator:
                    sig_parts.append("*")
                    added_kw_only_separator = True
                sig_parts.append(format_param(param))
                kwargs_dict.append(f'"{param.name}": {param.name}')
            elif param.kind == inspect.Parameter.VAR_KEYWORD:
                sig_parts.append(f"**{param.name}")
                # **kwargs gets merged
                kwargs_dict.append(f"**{param.name}")
            else:
                raise CodeInterpreterError(f"Unsupported parameter kind for tool '{tool_name}': {param.kind}")

        signature = ", ".join(sig_parts)
        args_str = ", ".join(args_list)
        kwargs_str = ", ".join(kwargs_dict)

        return _TOOL_WRAPPER_TEMPLATE.format(
            tool_name=tool_name,
            signature=signature,
            args_list=args_str,
            kwargs_dict=kwargs_str,
        )

    def _poll_and_execute_tools(self, code_done_event: threading.Event) -> None:
        """Poll for pending tool requests and execute them on the host."""
        headers = {"x-daytona-preview-token": self._broker_token}
        lease_seconds = self.tool_claim_lease_seconds

        def fetch_pending(max_items: int) -> list[dict[str, Any]]:
            url_pending = f"{self._broker_url}/pending?max={max_items}&lease_seconds={lease_seconds}"
            req = urllib.request.Request(url_pending, headers=headers, method="GET")
            with urllib.request.urlopen(req, timeout=5) as resp:
                data = json.loads(resp.read().decode("utf-8"))
            if isinstance(data, dict) and isinstance(data.get("requests"), list):
                return [r for r in data["requests"] if isinstance(r, dict) and "id" in r]
            if isinstance(data, dict) and "id" in data:
                return [data]
            return []

        def execute_one(pending: dict[str, Any]) -> None:
            call_id = pending["id"]
            tool_name = pending["tool_name"]
            args = pending.get("args", [])
            kwargs = pending.get("kwargs", {})
            claim_token = pending.get("claim_token")

            logger.info("Executing tool on host: %s(%s, %s)", tool_name, args, kwargs)

            try:
                tool_func = self.tools.get(tool_name)
                if tool_func is None:
                    result = json.dumps({"error": f"Unknown tool: {tool_name}"})
                else:
                    result = tool_func(*args, **kwargs)
                    if not isinstance(result, str):
                        result = json.dumps(result)
            except Exception as e:
                result = json.dumps({"error": str(e)})

            logger.info("Tool result for %s: %.500s", tool_name, result)

            url_result = f"{self._broker_url}/result/{call_id}"
            payload = json.dumps({"result": result, "claim_token": claim_token}).encode("utf-8")
            req = urllib.request.Request(
                url_result,
                data=payload,
                headers={**headers, "Content-Type": "application/json"},
                method="POST",
            )
            try:
                with urllib.request.urlopen(req, timeout=5):
                    pass
            except urllib.error.HTTPError as e:
                # Expected when this worker's lease expired and another worker took over.
                if e.code in (404, 409):
                    logger.info("Discarded stale tool result for call %s (HTTP %s)", call_id, e.code)
                    return
                raise

            logger.info("Tool result posted for call %s", call_id)

        inflight: dict[str, Future[None]] = {}

        with ThreadPoolExecutor(max_workers=self.max_concurrent_tool_calls) as executor:
            while not code_done_event.is_set() or inflight:
                for call_id, fut in list(inflight.items()):
                    if not fut.done():
                        continue
                    inflight.pop(call_id, None)
                    try:
                        fut.result()
                    except Exception as e:
                        logger.warning("Tool execution failed for call %s: %s", call_id, e)

                if code_done_event.is_set():
                    time.sleep(0.01)
                    continue

                capacity = self.max_concurrent_tool_calls - len(inflight)
                if capacity > 0:
                    try:
                        for pending in fetch_pending(capacity):
                            call_id = pending.get("id")
                            if not call_id or call_id in inflight:
                                continue
                            inflight[call_id] = executor.submit(execute_one, pending)
                    except (urllib.error.URLError, urllib.error.HTTPError, OSError):
                        pass
                    except Exception as e:
                        logger.warning("Error in tool polling loop: %s", e)

                time.sleep(0.05)

    def _inject_variables(self, code: str, variables: dict[str, Any]) -> str:
        """Insert Python assignments for each variable at the top of the code."""
        if not variables:
            return code

        for key in variables:
            if not key.isidentifier() or keyword.iskeyword(key):
                raise CodeInterpreterError(f"Invalid variable name: '{key}'")

        assignments = [f"{k} = {self._serialize_value(v)}" for k, v in variables.items()]
        return "\n".join(assignments) + "\n" + code

    def _serialize_value(self, value: Any) -> str:
        """Convert a Python value to a string that can be evaluated."""

        # Use Python literals (not JSON) so nested bool/None serialize correctly.
        def to_literal(obj: Any) -> str:
            if obj is None:
                return "None"
            if isinstance(obj, bool):
                return "True" if obj else "False"
            if isinstance(obj, int):
                return repr(obj)
            if isinstance(obj, float):
                if math.isnan(obj):
                    return "float('nan')"
                if math.isinf(obj):
                    return "float('inf')" if obj > 0 else "float('-inf')"
                return repr(obj)
            if isinstance(obj, str):
                return repr(obj)
            if isinstance(obj, list):
                return "[" + ", ".join(to_literal(v) for v in obj) + "]"
            if isinstance(obj, tuple):
                inner = ", ".join(to_literal(v) for v in obj)
                if len(obj) == 1:
                    inner += ","
                return "(" + inner + ")"
            if isinstance(obj, set):
                if not obj:
                    return "set()"
                return "{" + ", ".join(to_literal(v) for v in obj) + "}"
            if isinstance(obj, dict):
                parts = []
                for k, v in obj.items():
                    parts.append(f"{to_literal(k)}: {to_literal(v)}")
                return "{" + ", ".join(parts) + "}"
            raise CodeInterpreterError(f"Unsupported value type: {type(obj).__name__}")

        try:
            return to_literal(value)
        except TypeError as exc:
            raise CodeInterpreterError(f"Unsupported value type: {type(value).__name__}") from exc

    def execute(
        self,
        code: str,
        variables: dict[str, Any] | None = None,
    ) -> Any:
        """Execute Python code in the Daytona sandbox.

        Args:
            code: Python code to execute.
            variables: Variables to inject into the namespace before execution.

        Returns:
            One of:
            - FinalOutput: If SUBMIT() was called in code
            - str: Captured stdout from print() statements
            - None: If no output was produced

        Raises:
            CodeInterpreterError: On runtime errors
            SyntaxError: On invalid Python syntax
        """
        if not self._initialized:
            self.start()

        if variables:
            code = self._inject_variables(code, variables)

        # RLM sets output_fields after start(), so register typed SUBMIT
        # lazily on first execute() that has them.
        self._register_typed_submit()

        # RLM (and users) can add tools to self.tools after start().
        # For example, RLM injects llm_query and llm_query_batched before
        # each execution via interpreter.tools.update(...).  We lazily
        # start the broker on first use and inject any new wrappers.
        if self.tools:
            if not self._broker_url:
                self._start_broker()
            new_tools = set(self.tools) - self._injected_tools
            if new_tools:
                self._inject_tool_wrappers()

        # If we have tools, run with polling loop; otherwise run directly
        last_error: Exception | None = None
        for attempt in range(1, self._max_retries + 1):
            try:
                if self.tools and self._broker_url:
                    return self._execute_with_tools(code)
                return self._execute_direct(code)
            except (CodeInterpreterError, SyntaxError):
                raise
            except Exception as e:
                if attempt < self._max_retries and self._is_transient_error(e):
                    logger.warning(
                        "Transient error on attempt %d/%d, retrying in %.1fs: %s",
                        attempt,
                        self._max_retries,
                        self._retry_delay,
                        e,
                    )
                    last_error = e
                    time.sleep(self._retry_delay)
                    continue
                raise

        raise last_error  # type: ignore[misc]  # unreachable but keeps mypy happy

    @staticmethod
    def _is_transient_error(error: Exception) -> bool:
        """Return True if the error looks like a transient connection issue."""
        msg = str(error).lower()
        return any(
            phrase in msg
            for phrase in [
                "keepalive ping timeout",
                "ping timeout",
                "connection reset",
                "broken pipe",
                "close code 1011",
            ]
        )

    def _execute_direct(self, code: str) -> Any:
        """Execute code directly without tool support."""
        stdout_parts = []

        def on_stdout(msg):
            stdout_parts.append(msg.output)

        stderr_parts = []

        def on_stderr(msg):
            stderr_parts.append(msg.output)

        result = self._sandbox.code_interpreter.run_code(
            code,
            context=self._context,
            on_stdout=on_stdout,
            on_stderr=on_stderr,
            timeout=0,
        )

        stdout = "".join(stdout_parts)
        return self._process_result(result, stdout)

    def _execute_with_tools(self, code: str) -> Any:
        """Execute code with tool polling support."""
        # Storage for results from the code execution thread
        execution_result = {"result": None, "stdout": "", "error": None}
        code_done_event = threading.Event()

        stdout_parts = []

        def on_stdout(msg):
            stdout_parts.append(msg.output)

        stderr_parts = []

        def on_stderr(msg):
            stderr_parts.append(msg.output)

        def run_code():
            try:
                result = self._sandbox.code_interpreter.run_code(
                    code,
                    context=self._context,
                    on_stdout=on_stdout,
                    on_stderr=on_stderr,
                    timeout=0,
                )
                execution_result["result"] = result
                execution_result["stdout"] = "".join(stdout_parts)
            except Exception as e:
                execution_result["error"] = e
            finally:
                code_done_event.set()

        # Start code execution in background thread
        code_thread = threading.Thread(target=run_code)
        code_thread.start()

        # Poll for tool calls while code is running
        self._poll_and_execute_tools(code_done_event)

        # Wait for code thread to finish
        code_thread.join()

        # Check for thread-level errors
        err = execution_result["error"]
        if err is not None:
            raise err

        return self._process_result(
            execution_result["result"],
            execution_result["stdout"],
        )

    def _process_result(self, result, stdout: str) -> Any:
        """Process execution result and return appropriate output."""
        # Check for FinalOutput marker in stdout
        if self._final_output_marker in stdout:
            return self._extract_final_output(stdout)

        # Check for errors
        if result.error:
            error_name = result.error.name
            error_value = result.error.value

            if error_name == "FinalOutput":
                return self._extract_final_output(stdout)

            if error_name == "SyntaxError":
                raise SyntaxError(f"Invalid Python syntax: {error_value}")
            raise CodeInterpreterError(f"{error_name}: {error_value}")

        return stdout.strip() if stdout.strip() else None

    def _extract_final_output(self, stdout: str) -> FinalOutput:
        """Extract FinalOutput from stdout containing the marker."""
        marker = self._final_output_marker
        start = stdout.find(marker)
        if start == -1:
            return FinalOutput({})

        start += len(marker)
        end = stdout.find(marker, start)
        if end == -1:
            return FinalOutput({})

        json_str = stdout[start:end]
        try:
            output = json.loads(json_str)
            return FinalOutput(output)
        except json.JSONDecodeError:
            return FinalOutput({})

    def shutdown(self) -> None:
        """Release resources and delete the Daytona sandbox."""
        # Delete broker session if running
        if self._broker_session_id and self._sandbox:
            try:
                self._sandbox.process.delete_session(self._broker_session_id)
                logger.info("Deleted broker session")
            except Exception as e:
                logger.warning("Failed to delete broker session: %s", e)

        if self._context and self._sandbox:
            try:
                self._sandbox.code_interpreter.delete_context(self._context)
                logger.info("Deleted interpreter context")
            except Exception as e:
                logger.warning("Failed to delete context: %s", e)

        if self._sandbox:
            try:
                self._sandbox.delete()
                logger.info("Deleted sandbox")
            except Exception as e:
                logger.warning("Failed to delete sandbox: %s", e)

        self._context = None
        self._sandbox = None
        self._daytona = None
        self._initialized = False
        self._broker_url = None
        self._broker_token = None
        self._broker_session_id = None
        self._injected_tools = set()
        self._submit_registered = False

    def __enter__(self):
        self.start()
        return self

    def __exit__(self, *_):
        self.shutdown()

    def __call__(self, code: str, variables: dict[str, Any] | None = None) -> Any:
        return self.execute(code, variables)
