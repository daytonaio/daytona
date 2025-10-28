from daytona import CreateSandboxFromImageParams, Daytona, Resources, CodeInterpreter
import time

def run_code(code_interpreter: CodeInterpreter, code: str, timeout: int = None):
  code_interpreter.execute(code, timeout=timeout, on_stdout=lambda x: print(f"stdout: {x}"), on_stderr=lambda x: print(f"stderr: {x}"), on_error=lambda x, y, z: print(f"error:\n\tname:\n{x}\n\tvalue:\n{y}\n\ttraceback:\n{z}"), on_artifact=lambda x: print(f"artifact: {x}"), on_control=lambda x: print(f"control: {x}"))


def main():
  daytona = Daytona()

  params = CreateSandboxFromImageParams(image="python:3.9.23-slim")
  sandbox = daytona.create(params, timeout=150, on_snapshot_create_logs=print)

  try:
    preview_link = sandbox.get_preview_link(2280)
    code_interpreter = CodeInterpreter(
      preview_link.url, 
      headers={
        **sandbox._toolbox_api.api_client.default_headers,
        "X-Daytona-Preview-Token": preview_link.token,
      },
      )
    
    print("+++++++++++++++++++++++++++++")
    start_time = time.time()
    code = "import time; print('Hello before sleep!'); time.sleep(5); x = 12; print('Hello after sleep!');"
    print(f"Running code: {code}")
    run_code(code_interpreter, code, timeout=2)
    end_time = time.time()
    print(f"Time taken: {end_time - start_time} seconds")
    print("+++++++++++++++++++++++++++++")
  finally:
    daytona.delete(sandbox)

if __name__ == "__main__":
    main()
