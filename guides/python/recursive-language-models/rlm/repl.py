"""REPL execution for deeper-rlm agents.

Uses a broker server pattern for blocking rlm_query() calls:
- Flask broker server runs inside the sandbox on port 8080
- Sandbox code calls rlm_query() which POSTs to broker and BLOCKS
- External poller thread polls broker for pending requests
- Poller spawns sub-agent, gets result, POSTs back to broker
- Sandbox code unblocks with actual result
"""

import base64
import json
import logging
import re
import textwrap
import time
from dataclasses import dataclass, field
from threading import Event, Lock, Thread
from typing import TYPE_CHECKING, Callable

import requests

from rlm.types import CodeBlockResult

if TYPE_CHECKING:
    from daytona import Sandbox

from daytona import SessionExecuteRequest


def _retry_file_op(op_func, max_retries: int = 3, operation_name: str = "file operation"):
    """Retry a file operation on transient errors (307 redirects, 502, etc)."""
    last_error = None
    for attempt in range(max_retries):
        try:
            return op_func()
        except Exception as e:
            last_error = e
            error_str = str(e)
            # Retry on redirect or gateway errors
            if any(
                x in error_str for x in ["307", "302", "502", "503", "504", "Redirect", "Gateway"]
            ):
                if attempt < max_retries - 1:
                    logger.warning(
                        f"{operation_name} failed, retrying ({attempt + 1}/{max_retries}): {e}"
                    )
                    time.sleep(2 * (attempt + 1))
                    continue
            raise
    raise last_error


logger = logging.getLogger(__name__)


# Regex pattern to find ```python code blocks
CODE_BLOCK_PATTERN = re.compile(
    r"```python\s*\n(.*?)```",
    re.DOTALL | re.IGNORECASE,
)

# =============================================================================
# Broker Server Script (runs inside sandbox, handles rlm_query request queue)
# =============================================================================

_BROKER_SCRIPT = textwrap.dedent('''
import json
import threading
import time
import uuid
from flask import Flask, request, jsonify

app = Flask(__name__)

# Request queue: {request_id: {"request": {...}, "response": None, "event": Event}}
pending_requests = {}
lock = threading.Lock()
condition = threading.Condition(lock)

@app.route("/health")
def health():
    return jsonify({"status": "ok"})

@app.route("/enqueue", methods=["POST"])
def enqueue():
    """Called by sandbox code to submit an rlm_query request and wait for response."""
    data = request.json
    request_id = str(uuid.uuid4())
    event = threading.Event()

    with condition:
        pending_requests[request_id] = {
            "request": data,
            "response": None,
            "event": event,
        }
        condition.notify_all()  # Wake up any waiting /pending calls

    # Wait for response (with timeout)
    event.wait(timeout=1800)

    with lock:
        entry = pending_requests.pop(request_id, None)

    if entry and entry["response"] is not None:
        return jsonify(entry["response"])
    else:
        return jsonify({"error": "Request timed out"}), 504

@app.route("/pending")
def get_pending():
    """Called by parent process poller to long-poll for pending requests."""
    timeout_s = float(request.args.get("timeout_s", 0))
    deadline = time.time() + timeout_s

    with condition:
        while True:
            pending = [
                {"id": rid, "request": entry["request"]}
                for rid, entry in pending_requests.items()
                if entry["response"] is None
            ]
            if pending or time.time() >= deadline:
                return jsonify({"pending": pending})

            remaining = deadline - time.time()
            if remaining <= 0:
                return jsonify({"pending": []})
            condition.wait(timeout=min(remaining, 1.0))

@app.route("/respond", methods=["POST"])
def respond():
    """Called by parent process poller to submit a response."""
    data = request.json
    request_id = data.get("id")
    response = data.get("response")

    with lock:
        if request_id in pending_requests:
            pending_requests[request_id]["response"] = response
            pending_requests[request_id]["event"].set()
            return jsonify({"status": "ok"})

    return jsonify({"error": "Request not found"}), 404

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080, threaded=True)
''')


@dataclass
class REPLResult:
    """Result of executing code in the REPL."""

    code_blocks: list[CodeBlockResult] = field(default_factory=list)
    final_answer: str | None = None
    final_var_name: str | None = None


def find_code_blocks(response: str) -> list[str]:
    """Extract Python code blocks from a model response."""
    matches = CODE_BLOCK_PATTERN.findall(response)
    return [match.strip() for match in matches if match.strip()]


def find_final_answer(response: str) -> tuple[str | None, str | None]:
    """Find FINAL() or FINAL_VAR() call in response."""
    final_match = re.search(r'FINAL\s*\(\s*["\'](.+?)["\']\s*\)', response, re.DOTALL)
    if final_match:
        return final_match.group(1), "FINAL"

    final_var_match = re.search(r"FINAL\s*\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)", response)
    if final_var_match:
        return final_var_match.group(1), "FINAL_VAR"

    final_var_match2 = re.search(r'FINAL_VAR\s*\(\s*["\'](.+?)["\']\s*\)', response)
    if final_var_match2:
        return final_var_match2.group(1), "FINAL_VAR"

    return None, None


class DaytonaREPL:
    """
    REPL environment that executes Python code in a Daytona sandbox.

    Uses a broker server pattern for blocking rlm_query() calls:
    - Flask broker runs inside sandbox on port 8080
    - rlm_query() blocks until sub-agent returns result
    - External poller thread handles sub-agent spawning
    """

    BROKER_PORT = 8080
    LONG_POLL_TIMEOUT = 30  # seconds

    def __init__(
        self,
        sandbox: "Sandbox",
        rlm_query_handler: Callable[[str], str],
        rlm_query_batched_handler: Callable[[list[str]], list[str]],
        cwd: str = "/workspace",
        conda_env: str = "testbed",
        initial_variables: dict[str, str] | None = None,
    ):
        """
        Initialize the REPL.

        Args:
            sandbox: Daytona sandbox instance
            rlm_query_handler: Callback for rlm_query() - spawns sub-agent
            rlm_query_batched_handler: Callback for rlm_query_batched()
            cwd: Working directory
            conda_env: Conda environment
            initial_variables: Variables to inject into the REPL namespace (e.g., {"task": "..."})
        """
        self.sandbox = sandbox
        self.cwd = cwd
        self.conda_env = conda_env
        self.rlm_query_handler = rlm_query_handler
        self.rlm_query_batched_handler = rlm_query_batched_handler
        self.initial_variables = initial_variables or {}

        # Broker state
        self.broker_session_id: str | None = None
        self.broker_url: str | None = None
        self.broker_token: str | None = None
        self.poller_thread: Thread | None = None
        self.poller_stop: Event = Event()
        self._poller_lock: Lock = Lock()

        # REPL state
        self._final_answer: str | None = None
        self._final_var_name: str | None = None

        # Start broker
        self._start_broker()

    def _start_broker(self):
        """Start the broker server in the sandbox and begin polling."""
        logger.info("Starting broker server in sandbox...")

        # Upload broker script with retry
        _retry_file_op(
            lambda: self.sandbox.fs.upload_file(
                _BROKER_SCRIPT.encode("utf-8"), "/tmp/rlm_broker.py"
            ),
            operation_name="broker script upload",
        )

        # Install flask in the conda environment (may already be installed)
        logger.info("Installing flask in sandbox...")
        if self.conda_env:
            install_cmd = f"conda run -n {self.conda_env} pip install flask requests dill -q"
        else:
            install_cmd = "pip install flask requests dill -q"
        self.sandbox.process.exec(install_cmd, timeout=120)

        # Start broker in background session with conda env
        self.broker_session_id = "rlm-broker"
        self.sandbox.process.create_session(self.broker_session_id)

        if self.conda_env:
            broker_cmd = f"conda run -n {self.conda_env} python /tmp/rlm_broker.py"
        else:
            broker_cmd = "python /tmp/rlm_broker.py"

        self.sandbox.process.execute_session_command(
            self.broker_session_id,
            SessionExecuteRequest(
                command=broker_cmd,
                run_async=True,
            ),
        )

        # Get preview URL for HTTP communication
        preview = self.sandbox.get_preview_link(self.BROKER_PORT)
        self.broker_url = preview.url
        self.broker_token = preview.token
        logger.info(f"Broker URL: {self.broker_url}")

        # Wait for broker to be ready
        self._wait_for_broker()

        # Start polling thread
        self.poller_stop.clear()
        self.poller_thread = Thread(target=self._poll_broker, daemon=True)
        self.poller_thread.start()
        logger.info("Broker poller thread started")

    def _preview_headers(self) -> dict[str, str]:
        """Return headers required for preview URL authentication."""
        return {
            "X-Daytona-Preview-Token": self.broker_token or "",
            "X-Daytona-Skip-Preview-Warning": "true",
        }

    def _wait_for_broker(self, max_wait: int = 30):
        """Wait for broker to be ready by probing /health endpoint."""
        logger.info("Waiting for broker to be ready...")
        for i in range(max_wait * 2):
            try:
                resp = requests.get(
                    f"{self.broker_url}/health",
                    headers=self._preview_headers(),
                    timeout=2,
                )
                if resp.ok:
                    logger.info(f"Broker ready after {i * 0.5:.1f}s")
                    return
            except requests.RequestException:
                pass
            time.sleep(0.5)
        raise TimeoutError("Broker failed to start within timeout")

    def _poll_broker(self):
        """Poll the broker for pending rlm_query requests using long-polling."""
        session = requests.Session()
        logger.info("Broker poller started")

        while not self.poller_stop.is_set():
            try:
                resp = session.get(
                    f"{self.broker_url}/pending",
                    params={"timeout_s": self.LONG_POLL_TIMEOUT},
                    headers=self._preview_headers(),
                    timeout=self.LONG_POLL_TIMEOUT + 5,
                )

                for item in resp.json().get("pending", []):
                    request_id = item["id"]
                    req_data = item["request"]

                    # Handle the request
                    response = self._handle_request(req_data)

                    # Post response back to broker
                    session.post(
                        f"{self.broker_url}/respond",
                        headers=self._preview_headers(),
                        json={"id": request_id, "response": response},
                        timeout=10,
                    )

            except requests.RequestException as e:
                if not self.poller_stop.is_set():
                    logger.debug(f"Poller request failed: {e}")
                    time.sleep(1)

        logger.info("Broker poller stopped")

    def _handle_request(self, req_data: dict) -> dict:
        """Handle an rlm_query request from the sandbox."""
        req_type = req_data.get("type")

        if req_type == "single":
            task = req_data.get("task", "")
            logger.info(f"Handling rlm_query: {task[:60]}...")
            try:
                result = self.rlm_query_handler(task)
                return {"result": result}
            except Exception as e:
                logger.exception(f"Error handling rlm_query: {e}")
                return {"error": str(e)}

        elif req_type == "batched":
            tasks = req_data.get("tasks", [])
            logger.info(f"Handling rlm_query_batched with {len(tasks)} tasks")
            try:
                results = self.rlm_query_batched_handler(tasks)
                return {"results": results}
            except Exception as e:
                logger.exception(f"Error handling rlm_query_batched: {e}")
                return {"error": str(e)}

        return {"error": f"Unknown request type: {req_type}"}

    def _build_execution_script(self, code: str) -> str:
        """
        Build a Python script that executes code with blocking rlm_query().

        The script:
        - Sets up rlm_query() that blocks via HTTP to broker
        - Sets up FINAL/FINAL_VAR functions
        - Injects initial variables (e.g., task) into the namespace
        - Executes the user code
        - Captures stdout/stderr
        - Uses dill for state persistence (variables AND imports persist)
        """
        code_b64 = base64.b64encode(code.encode()).decode()

        # Encode initial variables as base64 to avoid escaping issues
        initial_vars_b64 = base64.b64encode(json.dumps(self.initial_variables).encode()).decode()

        return textwrap.dedent(f'''
import sys
import io
import json
import base64
import traceback
import os
import requests

try:
    import dill
except ImportError:
    import pickle as dill

# =============================================================================
# File Editing Function (Claude Code style)
# =============================================================================

def edit_file(file_path: str, old_string: str, new_string: str, replace_all: bool = False) -> str:
    """
    Edit a file by replacing old_string with new_string.

    Args:
        file_path: Path to the file to edit
        old_string: The exact text to find and replace (must be unique unless replace_all=True)
        new_string: The text to replace it with
        replace_all: If True, replace all occurrences. If False, old_string must be unique.

    Returns:
        Success message or error description

    Example:
        edit_file("/workspace/src/module.py", "def old_func():", "def new_func():")
    """
    import os

    # Check file exists
    if not os.path.exists(file_path):
        return "Error: File not found: {{}}".format(file_path)

    # Read file
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception as e:
        return "Error reading file: {{}}".format(e)

    # Check old_string exists
    if old_string not in content:
        # Provide helpful debug info
        lines = old_string.split('\\n')
        first_line = lines[0][:50] if lines else old_string[:50]
        return "Error: old_string not found in {{}}. First line of search: '{{}}...'".format(file_path, first_line)

    # Check uniqueness (unless replace_all)
    count = content.count(old_string)
    if count > 1 and not replace_all:
        return "Error: old_string appears {{}} times in {{}}. Use replace_all=True or provide more context to make it unique.".format(count, file_path)

    # Perform replacement
    if replace_all:
        new_content = content.replace(old_string, new_string)
        replaced_count = count
    else:
        new_content = content.replace(old_string, new_string, 1)
        replaced_count = 1

    # Verify the edit was valid Python if it's a .py file
    if file_path.endswith('.py'):
        try:
            compile(new_content, file_path, 'exec')
        except SyntaxError as e:
            return "Error: Edit would create invalid Python syntax: {{}}".format(e)

    # Write file
    try:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(new_content)
    except Exception as e:
        return "Error writing file: {{}}".format(e)

    suffix = 's' if replaced_count > 1 else ''
    return "Successfully edited {{}} ({{}} replacement{{}})".format(file_path, replaced_count, suffix)


# =============================================================================
# LLM Query Functions (via local broker - BLOCKING)
# =============================================================================

BROKER_URL = "http://127.0.0.1:{self.BROKER_PORT}"

def rlm_query(task: str) -> str:
    """Query a sub-agent via the broker. BLOCKS until result is ready."""
    try:
        response = requests.post(
            BROKER_URL + "/enqueue",
            json={{"type": "single", "task": task}},
            timeout=1800,
        )
        data = response.json()
        if data.get("error"):
            return "Error: {{}}".format(data['error'])
        return data.get("result", "Error: No result")
    except Exception as e:
        return "Error: rlm_query failed - {{}}".format(e)


def rlm_query_batched(tasks: list) -> list:
    """Query multiple sub-agents in parallel. BLOCKS until all results ready."""
    try:
        response = requests.post(
            BROKER_URL + "/enqueue",
            json={{"type": "batched", "tasks": tasks}},
            timeout=1800,
        )
        data = response.json()
        if data.get("error"):
            return ["Error: {{}}".format(data['error'])] * len(tasks)
        return data.get("results", ["Error: No result"] * len(tasks))
    except Exception as e:
        return ["Error: rlm_query_batched failed - {{}}".format(e)] * len(tasks)


# =============================================================================
# State Management (using dill for complex objects including modules)
# =============================================================================

STATE_FILE = "/tmp/rlm_state.dill"

def load_state():
    if os.path.exists(STATE_FILE):
        try:
            with open(STATE_FILE, "rb") as f:
                return dill.load(f)
        except:
            pass
    return {{}}

def save_state(state):
    clean_state = {{}}
    for k, v in state.items():
        if k.startswith("_"):
            continue
        try:
            dill.dumps(v)
            clean_state[k] = v
        except:
            pass
    with open(STATE_FILE, "wb") as f:
        dill.dump(clean_state, f)

def serialize_locals(state):
    result = {{}}
    for k, v in state.items():
        if k.startswith("_"):
            continue
        try:
            result[k] = repr(v)
        except:
            result[k] = "<{{}}>".format(type(v).__name__)
    return result

# =============================================================================
# FINAL functions
# =============================================================================

_final_answer = None
_final_var_name = None

def FINAL(answer):
    """Return final answer to parent."""
    global _final_answer
    _final_answer = str(answer)

def FINAL_VAR(variable_name):
    """Return a variable as final answer."""
    global _final_var_name
    _final_var_name = variable_name.strip().strip("\\"\\'")

# =============================================================================
# Execution
# =============================================================================

_locals = load_state()

_globals = {{
    "__builtins__": __builtins__,
    "__name__": "__main__",
    "rlm_query": rlm_query,
    "rlm_query_batched": rlm_query_batched,
    "FINAL": FINAL,
    "FINAL_VAR": FINAL_VAR,
    "edit_file": edit_file,
}}

# Inject initial variables (e.g., task)
_initial_vars = json.loads(base64.b64decode("{initial_vars_b64}").decode())
_globals.update(_initial_vars)

code = base64.b64decode("{code_b64}").decode()

stdout_buf = io.StringIO()
stderr_buf = io.StringIO()
old_stdout, old_stderr = sys.stdout, sys.stderr

try:
    sys.stdout = stdout_buf
    sys.stderr = stderr_buf
    combined = {{**_globals, **_locals}}
    exec(code, combined, combined)
    # Update locals with new variables
    for key, value in combined.items():
        if key not in _globals and not key.startswith("_"):
            _locals[key] = value
except Exception as e:
    traceback.print_exc(file=stderr_buf)
finally:
    sys.stdout = old_stdout
    sys.stderr = old_stderr

save_state(_locals)

# Handle FINAL_VAR
_final_answer_resolved = _final_answer
if _final_var_name and _final_var_name in _locals:
    _final_answer_resolved = str(_locals[_final_var_name])

# Print stdout first
print(stdout_buf.getvalue(), end="")

# Print JSON result marker
print("___REPL_RESULT___", end="")
print(json.dumps({{
    "stderr": stderr_buf.getvalue(),
    "final_answer": _final_answer_resolved,
    "final_var_name": _final_var_name,
    "locals": serialize_locals(_locals),
}}))
''')

    def execute_code(self, code: str) -> CodeBlockResult:
        """Execute a Python code block in the sandbox."""
        start_time = time.time()

        # Build execution script
        script = self._build_execution_script(code)

        # Write script to file to avoid shell escaping issues (with retry)
        _retry_file_op(
            lambda: self.sandbox.fs.upload_file(script.encode("utf-8"), "/tmp/exec_script.py"),
            operation_name="exec script upload",
        )

        # Execute in sandbox with conda environment
        if self.conda_env:
            cmd = f"conda run -n {self.conda_env} --no-capture-output python /tmp/exec_script.py"
        else:
            cmd = "python /tmp/exec_script.py"

        # Execute with retry for transient errors (502, etc)
        max_retries = 3
        for attempt in range(max_retries):
            try:
                result = self.sandbox.process.exec(cmd, cwd=self.cwd, timeout=1800)
                break
            except Exception as e:
                error_str = str(e)
                # Retry on 502/503/504 gateway errors
                if any(code in error_str for code in ["502", "503", "504", "Bad Gateway"]):
                    if attempt < max_retries - 1:
                        logger.warning(
                            f"Command execution failed with gateway error, retrying ({attempt + 1}/{max_retries})"
                        )
                        time.sleep(2 * (attempt + 1))
                        continue
                raise

        execution_time = time.time() - start_time
        output = result.result or ""

        # Parse output
        stdout = ""
        stderr = ""
        error = None

        try:
            if "___REPL_RESULT___" in output:
                parts = output.split("___REPL_RESULT___")
                stdout = parts[0]
                if len(parts) > 1:
                    result_json = json.loads(parts[1])
                    stderr = result_json.get("stderr", "")
                    self._final_answer = result_json.get("final_answer")
                    self._final_var_name = result_json.get("final_var_name")
            else:
                stdout = output
                if result.exit_code != 0:
                    error = output
        except (json.JSONDecodeError, IndexError) as e:
            stdout = output
            logger.warning(f"Failed to parse REPL result: {e}")

        return CodeBlockResult(
            code=code,
            stdout=stdout,
            stderr=stderr,
            execution_time=execution_time,
            error=error,
        )

    def execute_response(self, response: str) -> REPLResult:
        """Execute all Python code blocks in a model response."""
        result = REPLResult()

        # Find and execute Python code blocks
        code_blocks = find_code_blocks(response)

        for code in code_blocks:
            block_result = self.execute_code(code)
            result.code_blocks.append(block_result)

            # Check for final answer
            if self._final_answer is not None:
                result.final_answer = self._final_answer
                break
            if self._final_var_name is not None:
                result.final_var_name = self._final_var_name
                break

        # Also check response text for FINAL() that might not be in code
        if result.final_answer is None and result.final_var_name is None:
            answer, answer_type = find_final_answer(response)
            if answer_type == "FINAL":
                result.final_answer = answer
            elif answer_type == "FINAL_VAR":
                result.final_var_name = answer

        return result

    def cleanup(self):
        """Stop the broker and cleanup resources."""
        logger.info("Cleaning up REPL...")

        # Stop poller thread
        self.poller_stop.set()
        if self.poller_thread is not None:
            self.poller_thread.join(timeout=2)
            self.poller_thread = None

        # Delete broker session
        if self.broker_session_id:
            try:
                self.sandbox.process.delete_session(self.broker_session_id)
            except Exception as e:
                logger.debug(f"Failed to delete broker session: {e}")
            self.broker_session_id = None

    def __del__(self):
        self.cleanup()
