# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import json
import requests
from typing import Optional, Dict, Callable, List
from websockets.sync.client import connect
from websockets.exceptions import ConnectionClosed
import re


class CodeInterpreter:
    def __init__(self, sandbox_base_url: str, headers: Optional[Dict[str, str]] = None):
        self.sandbox_base_url = sandbox_base_url
        self.headers = headers

    def create_context(
        self,
        cwd: Optional[str] = None,
        language: Optional[str] = None,
    ) -> Dict:
        """
        Create a new isolated interpreter context.
        
        Args:
            cwd: Working directory for the context (defaults to sandbox workDir)
            language: Language for the context (currently only "python", default: "python")
            
        Returns:
            Dict with context information (id, cwd, language, createdAt, active)
            
        Raises:
            requests.HTTPError: If context creation fails
        """
        url = f"{self.sandbox_base_url}/process/interpreter/context"
        
        payload = {}
        if cwd is not None:
            payload["cwd"] = cwd
        if language is not None:
            payload["language"] = language
            
        response = requests.post(url, json=payload, headers=self.headers)
        response.raise_for_status()
        
        return response.json()

    def list_contexts(self) -> List[Dict]:
        """
        List all user-created interpreter contexts (excludes default context).
        
        Returns:
            List of context dictionaries with (id, cwd, language, createdAt, active)
            
        Raises:
            requests.HTTPError: If request fails
        """
        url = f"{self.sandbox_base_url}/process/interpreter/context"
        
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        
        return response.json().get("contexts", [])

    def delete_context(self, context_id: str) -> None:
        """
        Delete an interpreter context and shut down its worker process.
        
        Args:
            context_id: ID of the context to delete
            
        Raises:
            requests.HTTPError: If deletion fails or context not found
        """
        url = f"{self.sandbox_base_url}/process/interpreter/context/{context_id}"
        
        response = requests.delete(url, headers=self.headers)
        response.raise_for_status()

    def execute(
        self,
        code: str,
        context_id: Optional[str] = None,
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
            context_id: Optional context ID (if not provided, uses default context)
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
                if context_id is not None:
                    request["contextId"] = context_id
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
                            # Check for completion signals (but not "interrupted" - wait for close frame)
                            if control_text in ["completed", "error_completed", "exit"]:
                                break  # Exit the loop to close WebSocket
                            # For "interrupted", don't break - wait for server to close with error code
                            
                    except ConnectionClosed as e:
                        # Check if closed with timeout error code
                        if e.rcvd and e.rcvd.code == 4008:
                            if on_error:
                                on_error("TimeoutError", "Execution timeout - code took too long to execute", "")
                            raise TimeoutError("Execution timeout - code took too long to execute")
                        # Normal close - break out
                        break
                    except json.JSONDecodeError:
                        # Invalid JSON - skip this message
                        continue
                        
        except ConnectionClosed as e:
            # Check if closed with error code (e.g., timeout)
            if e.rcvd and e.rcvd.code == 4008:
                # Timeout error
                if on_error:
                    on_error("TimeoutError", "Execution timeout - code took too long to execute", "")
                raise TimeoutError("Execution timeout - code took too long to execute")
            # Other abnormal close codes
            elif e.rcvd and e.rcvd.code not in [1000, 1001]:
                # Abnormal close
                reason = e.rcvd.reason if e.rcvd and e.rcvd.reason else "Connection closed abnormally"
                if on_error:
                    on_error("ConnectionError", reason, "")
                raise ConnectionError(reason)
            # Normal close (1000, 1001) - just return
        except Exception as e:
            # Connection failed or other error
            if on_error:
                on_error("ConnectionError", str(e), "")
            raise