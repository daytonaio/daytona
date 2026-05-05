---
title: Advanced Examples
description: Real-world Daytona SDK usage examples
---

This section demonstrates real-world usage of the Daytona SDK for executing AI-generated code securely inside isolated sandboxes. These examples go beyond basic usage and show how developers can build AI-powered systems using Daytona.

---

## 🤖 Example 1: AI-Generated Code Execution

This example shows how to take a natural language prompt, generate code using an AI model, and execute it securely in a Daytona sandbox.

### 🔹 Workflow

```
apps/docs/src/content/docs/en/examples/daytona_ai_workflow_diagram.png
```

---

### 🔹 Input Prompt

```
"Write Python code to sort a list of numbers"
```

---

### 🔹 Generated Code (from AI)

```python
numbers = [5, 2, 9, 1]
print(sorted(numbers))
```

---

### 🔹 Python Implementation

```python
from daytona import Daytona

# Step 1: User input prompt
prompt = "Write Python code to sort a list of numbers"

# Step 2: Simulated AI-generated code
generated_code = """
numbers = [5, 2, 9, 1]
print(sorted(numbers))
"""

# Step 3: Initialize Daytona
daytona = Daytona()

# Step 4: Create sandbox
sandbox = daytona.create()

# Step 5: Execute generated code
response = sandbox.execute(
    command="python3",
    code=generated_code
)

# Step 6: Print output
print("Execution Output:", response.output)
```

---

### 🔹 Output

```
[1, 2, 5, 9]
```

---

### 🔹 Explanation

* The user provides a natural language prompt
* An AI model converts the prompt into executable Python code
* The code is executed securely inside a Daytona sandbox
* The output is returned without affecting the host system

---

## ⚙️ Example 2: Using File, Git, and Execute APIs

This example demonstrates how to combine multiple Daytona APIs in a single workflow.

### 🔹 Use Case

Clone a repository, modify a file, and execute a script.

---

### 🔹 Python Example

```python
from daytona import Daytona

daytona = Daytona()
sandbox = daytona.create()

# Clone repository
sandbox.git.clone("https://github.com/example/repo.git")

# Write to a file
sandbox.fs.write_file("repo/test.py", "print('Hello from Daytona')")

# Execute the file
result = sandbox.execute(
    command="python3 repo/test.py"
)

print(result.output)
```

---

### 🔹 Explanation

* Git API is used to clone a repository
* File API is used to modify files inside the sandbox
* Execute API runs the script inside the isolated environment

---

## 🔄 Example 3: Persistent Sandbox Workflow

This example shows how to reuse the same sandbox across multiple operations.

---

### 🔹 Python Example

```python
from daytona import Daytona

daytona = Daytona()
sandbox = daytona.create()

# Step 1: Install dependency
sandbox.execute(command="pip install numpy")

# Step 2: Run computation
result = sandbox.execute(
    command="python3",
    code="import numpy as np; print(np.array([1,2,3]))"
)

print(result.output)
```

---

### 🔹 Explanation

* The sandbox persists across multiple commands
* Dependencies installed remain available
* Useful for long-running workflows and AI agents

---

## ⚡ Example 4: Parallel Sandbox Execution

This example demonstrates running multiple sandboxes concurrently.

---

### 🔹 Python Example

```python
from daytona import Daytona
import concurrent.futures

daytona = Daytona()

def run_code(code):
    sandbox = daytona.create()
    return sandbox.execute(command="python3", code=code).output

codes = [
    "print(1+1)",
    "print(2+2)",
    "print(3+3)"
]

with concurrent.futures.ThreadPoolExecutor() as executor:
    results = list(executor.map(run_code, codes))

print(results)
```

---

### 🔹 Explanation

* Multiple sandboxes run independently
* Enables parallel execution of AI-generated tasks
* Useful for scaling workloads

---

## ❌ Error Handling and Retry Logic

Robust applications must handle failures gracefully.

---

### 🔹 Python Example

```python
from daytona import Daytona

daytona = Daytona()
sandbox = daytona.create()

code = "print(unknown_variable)"  # This will fail

for attempt in range(3):
    try:
        result = sandbox.execute(command="python3", code=code)
        print(result.output)
        break
    except Exception as e:
        print(f"Attempt {attempt+1} failed:", e)
```

---

### 🔹 Explanation

* Errors during execution are caught
* Retry logic improves reliability
* Essential for production-grade AI workflows

---

## 🤖 Optional: Agentic Workflow (Advanced)

This example shows how Daytona can be used inside an AI agent loop for self-correcting code execution.

---

### 🔹 Python Example

```python
while True:
    code = llm.generate(prompt)
    
    result = sandbox.execute(command="python3", code=code)
    
    if "error" not in result.output:
        break
    
    prompt = f"Fix this error: {result.output}"
```

---

### 🔹 Explanation

* The agent generates code
* Executes it using Daytona
* Fixes errors iteratively
* Demonstrates autonomous AI workflows

---

## 📊 Architecture Overview

```
Client → AI Model → Daytona → Sandbox → Execution → Response
```

---

## 📚 Additional Resources

* Python Notebook: `examples/python/advanced_usage.ipynb`
* TypeScript Example: `examples/typescript/advanced_usage.ts`

---

## 🧠 Summary

These examples demonstrate how Daytona can be used as a secure execution layer for AI-generated code, enabling developers to build scalable, reliable, and safe AI-powered applications.
