# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import json
from typing import Optional, Dict, Callable
from websockets.sync.client import connect
from websockets.exceptions import ConnectionClosed
import re


class CodeInterpreter:
    def __init__(self, sandbox_base_url: str, headers: Optional[Dict[str, str]] = None):
        self.sandbox_base_url = sandbox_base_url
        self.headers = headers

    def execute(
        self,
        code: str,
        envs: Optional[Dict[str, str]] = None,
        timeout: Optional[int] = None,
        on_stdout: Optional[Callable[[str], None]] = None,
        on_stderr: Optional[Callable[[str], None]] = None,
        on_error: Optional[Callable[[str, str, str], None]] = None,
        on_artifact: Optional[Callable[[Dict], None]] = None,
        on_control: Optional[Callable[[str], None]] = None,
    ) -> None:
        """
        Execute Python code in the remote interpreter.
        
        Args:
            code: Python code to execute
            envs: Optional environment variables for this execution
            timeout: Optional timeout in seconds (0 or None means no timeout)
            on_stdout: Callback for stdout chunks (receives text)
            on_stderr: Callback for stderr chunks (receives text)
            on_error: Callback for error chunks (receives name, value, traceback)
            on_artifact: Callback for artifact chunks (receives artifact dict)
            on_control: Callback for control chunks (receives text)
        """
        url = re.sub(r"^http", "ws", self.sandbox_base_url) + f"/process/interpreter/execute"
        
        try:
            with connect(url, additional_headers=self.headers) as websocket:
                # Send execution request as first message
                request = {"code": code}
                if envs is not None:
                    request["envs"] = envs
                if timeout is not None and timeout > 0:
                    request["timeout"] = timeout
                
                websocket.send(json.dumps(request))
                
                # Process streaming chunks
                while True:
                    try:
                        message = websocket.recv()
                        data = json.loads(message)
                        chunk_type = data.get("type")
                        
                        if chunk_type == "stdout" and on_stdout:
                            on_stdout(data.get("text", ""))
                        elif chunk_type == "stderr" and on_stderr:
                            on_stderr(data.get("text", ""))
                        elif chunk_type == "error" and on_error:
                            on_error(
                                data.get("name", ""),
                                data.get("value", ""),
                                data.get("traceback", "")
                            )
                        elif chunk_type == "artifact" and on_artifact:
                            on_artifact(data.get("artifact", {}))
                        elif chunk_type == "control":
                            control_text = data.get("text", "")
                            if on_control:
                                on_control(control_text)
                            # Check for completion signals
                            if control_text in ["completed", "interrupted", "error_completed", "exit"]:
                                break  # Exit the loop to close WebSocket
                            
                    except ConnectionClosed:
                        # WebSocket closed - execution completed
                        break
                    except json.JSONDecodeError:
                        # Invalid JSON - skip this message
                        continue
                        
        except Exception as e:
            # Connection failed or other error
            if on_error:
                on_error("ConnectionError", str(e), "")