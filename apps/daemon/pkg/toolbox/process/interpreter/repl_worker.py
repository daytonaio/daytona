#!/usr/bin/env python3
"""
Stateful Python REPL Worker for Daytona
Communicates via JSON over stdin/stdout
Maintains persistent global context between executions
"""

import sys
import json
import signal
import traceback
import io
import os
import base64
from contextlib import redirect_stdout, redirect_stderr


class REPLWorker:
    def __init__(self):
        self.globals = {"__name__": "__main__", "__doc__": None, "__package__": None}
        self.interrupted = False
        self.setup_signal_handlers()
        # Configure matplotlib for headless rendering if available
        self.matplotlib_enabled = False
        os.environ.setdefault("MPLBACKEND", "Agg")
        try:
            import matplotlib  # noqa: F401
            matplotlib.use("Agg", force=True)
            import matplotlib.pyplot as plt  # noqa: F401
            self.plt = plt
            self.matplotlib_enabled = True
        except Exception:
            self.plt = None
    
    def setup_signal_handlers(self):
        """Set up signal handlers for graceful interruption"""
        def sigint_handler(signum, frame):
            self.interrupted = True
            raise KeyboardInterrupt("Execution interrupted")
        
        signal.signal(signal.SIGINT, sigint_handler)
    
    
    def send_stream(self, msg_id, stream_name, text):
        """Send a stream chunk (stdout or stderr)"""
        if not text:
            return
        chunk = {
            "type": stream_name,  # "stdout" or "stderr"
            "text": text,
        }
        self._send_chunk(chunk)
    
    def _send_chunk(self, chunk):
        """Send a chunk via line-delimited JSON"""
        try:
            json.dump(chunk, sys.__stdout__)
            sys.__stdout__.write("\n")
            sys.__stdout__.flush()
        except Exception as e:
            sys.__stderr__.write(f"Failed to send chunk: {e}\n")
    
    def execute_code(self, msg_id, code):
        """Execute code and capture output, maintaining persistent state"""
        self.interrupted = False
        
        # Capture stdout and stderr
        stdout_buffer = io.StringIO()
        stderr_buffer = io.StringIO()
        
        try:
            with redirect_stdout(stdout_buffer), redirect_stderr(stderr_buffer):
                # Compile the code first to check for syntax errors
                compiled_code = compile(code, "<string>", "exec")
                
                # Execute in the persistent global context
                exec(compiled_code, self.globals)
            
            # Send captured output
            stdout_text = stdout_buffer.getvalue()
            stderr_text = stderr_buffer.getvalue()
            
            if stdout_text:
                self.send_stream(msg_id, "stdout", stdout_text)
            if stderr_text:
                self.send_stream(msg_id, "stderr", stderr_text)
            # Collect and emit any matplotlib figures as artifact chunks
            self._emit_matplotlib_artifacts()
            
            # Send completion signal for successful execution
            self._send_chunk({
                "type": "control",
                "text": "completed"
            })
            
        except KeyboardInterrupt:
            # Handle interruption - send any partial output first
            stdout_text = stdout_buffer.getvalue()
            stderr_text = stderr_buffer.getvalue()
            
            if stdout_text:
                self.send_stream(msg_id, "stdout", stdout_text)
            if stderr_text:
                self.send_stream(msg_id, "stderr", stderr_text)
            # Collect and emit any matplotlib figures
            self._emit_matplotlib_artifacts()
            
            # Send completion signal for interrupted execution
            self._send_chunk({
                "type": "control", 
                "text": "interrupted"
            })
            
        except SystemExit as e:
            # User called exit() - send output and exit status
            stdout_text = stdout_buffer.getvalue()
            stderr_text = stderr_buffer.getvalue()
            
            if stdout_text:
                self.send_stream(msg_id, "stdout", stdout_text)
            if stderr_text:
                self.send_stream(msg_id, "stderr", stderr_text)
            # Collect and emit any matplotlib figures
            self._emit_matplotlib_artifacts()
            
            # Send completion signal for exit
            self._send_chunk({
                "type": "control",
                "text": "exit"
            })
            return False  # Signal to stop the worker
            
        except Exception as e:
            # Handle execution errors
            stdout_text = stdout_buffer.getvalue()
            stderr_text = stderr_buffer.getvalue()
            
            if stdout_text:
                self.send_stream(msg_id, "stdout", stdout_text)
            if stderr_text:
                self.send_stream(msg_id, "stderr", stderr_text)
            
            # Get traceback
            tb_lines = traceback.format_exception(type(e), e, e.__traceback__)
            tb_text = "".join(tb_lines)
            # Emit error chunk
            self._send_chunk({
                "type": "error",
                "name": type(e).__name__,
                "value": str(e),
                "traceback": tb_text,
            })
            # Collect and emit any matplotlib figures
            self._emit_matplotlib_artifacts()
            
            # Send completion signal for error execution
            self._send_chunk({
                "type": "control",
                "text": "error_completed"
            })
        
        return True  # Continue running

    def _emit_matplotlib_artifacts(self):
        """Collect and emit open matplotlib figures as artifact chunks (base64 PNG)."""
        if not self.matplotlib_enabled or self.plt is None:
            return
        try:
            fignums = list(self.plt.get_fignums())
            for num in fignums:
                try:
                    fig = self.plt.figure(num)
                    buf = io.BytesIO()
                    fig.savefig(buf, format="png")
                    buf.seek(0)
                    data = base64.b64encode(buf.read()).decode("ascii")
                    self._send_chunk({
                        "type": "artifact",
                        "artifact": {"chart": data},
                    })
                finally:
                    try:
                        self.plt.close(fig)
                    except Exception:
                        pass
        except Exception:
            return
    
    def handle_command(self, command):
        """Handle a single command from stdin"""
        try:
            msg = json.loads(command)
            msg_id = msg.get("id", "unknown")
            cmd = msg.get("cmd")
            
            if cmd == "exec":
                code = msg.get("code", "")
                return self.execute_code(msg_id, code)
            elif cmd == "shutdown":
                return False  # Exit gracefully
            else:
                self._send_chunk({
                    "type": "error",
                    "name": "InvalidCommand",
                    "value": f"Unknown command: {cmd}",
                    "traceback": ""
                })
                return True
                
        except json.JSONDecodeError as e:
            # Can't parse JSON - send error
            self._send_chunk({
                "type": "error",
                "name": "JSONDecodeError",
                "value": str(e),
                "traceback": ""
            })
            return True
        except Exception as e:
            # Unexpected error in command handling
            self._send_chunk({
                "type": "error",
                "name": type(e).__name__,
                "value": str(e),
                "traceback": ""
            })
            return True
    
    def run(self):
        """Main loop - read commands from stdin and execute them"""
        while True:
            try:
                line = sys.stdin.readline()
                if not line:
                    # EOF - exit gracefully
                    break
                
                line = line.strip()
                if not line:
                    continue
                
                should_continue = self.handle_command(line)
                if not should_continue:
                    break
                    
            except KeyboardInterrupt:
                # Top-level interrupt - just continue
                continue
            except Exception as e:
                # Unexpected error in main loop
                sys.__stderr__.write(f"Fatal error in main loop: {e}\n")
                break


if __name__ == "__main__":
    worker = REPLWorker()
    worker.run()

