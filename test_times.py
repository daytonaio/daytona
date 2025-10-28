from daytona import CreateSandboxFromImageParams, Daytona, Resources, CodeInterpreter
import time

def run_code(code_interpreter: CodeInterpreter, code: str, timeout: int = None):
  code_interpreter.execute(code, timeout=timeout, on_stdout=lambda x: print(f"stdout: {x}"), on_stderr=lambda x: print(f"stderr: {x}"), on_error=lambda x, y, z: print(f"error:\n\tname:\n{x}\n\tvalue:\n{y}\n\ttraceback:\n{z}"), on_artifact=lambda x: print(f"artifact: {x}"), on_control=lambda x: print(f"control: {x}"))


def main():
  daytona = Daytona()

  params = CreateSandboxFromImageParams(image="python:3.9.23-slim")
  first_run_times = []
  second_run_times = []
  for i in range(10):
    sandbox = daytona.create(params, timeout=150, on_snapshot_create_logs=print)

    time.sleep(1)

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
      code = "print('Hello, World 1!')"
      print(f"Running code: {code}")
      start_time = time.time()
      run_code(code_interpreter, code)
      end_time = time.time()
      first_run_times.append((end_time - start_time) * 1000)
      print(f"First run time: {end_time - start_time} seconds")
      print("+++++++++++++++++++++++++++++")
      code = "print('Hello, World 2!')"
      print(f"Running code: {code}")
      start_time = time.time()
      run_code(code_interpreter, code)
      end_time = time.time()
      second_run_times.append((end_time - start_time) * 1000)
      print(f"Second run time: {end_time - start_time} seconds")
      print("+++++++++++++++++++++++++++++")
    except Exception as e:
      print(f"Execution Error: {e}")
    finally:
      daytona.delete(sandbox)

  print(f"sample size: {len(first_run_times)}")
  print(f"Average first run times: {sum(first_run_times) / len(first_run_times)} ms")
  print(f"Average second run times: {sum(second_run_times) / len(second_run_times)} ms")

if __name__ == "__main__":
    main()
