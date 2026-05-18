import express from "express";

const app = express();
app.use(express.json());

/**
 * 1. Simulated AI Code Generator
 */
function generateCode(prompt: string): string {
  if (prompt.includes("factorial")) {
    return `
    function factorial(n) {
      if (n === 0) return 1;
      return n * factorial(n - 1);
    }
    return factorial(5);
    `;
  }

  if (prompt.includes("fibonacci")) {
    return `
    function fib(n) {
      if (n <= 1) return n;
      return fib(n - 1) + fib(n - 2);
    }
    return fib(6);
    `;
  }

  if (prompt.includes("error")) {
    return `return x + 1;`; // intentional error
  }

  return `return "Unsupported prompt";`;
}

/**
 * 2. Sandbox Execution
 * ⚠️ Demo only — not production-safe
 */
function runSandbox(code: string) {
  try {
    const fn = new Function(code);
    const result = fn();

    return {
      output: String(result),
      error: null,
    };
  } catch (err: any) {
    return {
      output: null,
      error: err.message,
    };
  }
}

/**
 * 3. Output Validation
 */
function validate(prompt: string, output: string | null): boolean {
  const expected: Record<string, string> = {
    factorial: "120",
    fibonacci: "8",
  };

  if (!output) return false;

  for (const key in expected) {
    if (prompt.includes(key)) {
      return output === expected[key];
    }
  }

  return false;
}

/**
 * 4. API Endpoint (Full Pipeline)
 */
app.post("/execute", (req, res) => {
  const { prompt } = req.body;

  if (!prompt) {
    return res.status(400).json({
      success: false,
      error: "Prompt is required",
    });
  }

  // Step 1: Generate code
  const code = generateCode(prompt);

  // Step 2: Execute
  const result = runSandbox(code);

  // Step 3: Validate
  const isValid = validate(prompt, result.output);

  // Step 4: Response
  res.json({
    prompt,
    generatedCode: code,
    output: result.output,
    error: result.error,
    valid: isValid,
  });
});

/**
 * 5. Start Server
 */
const PORT = 3000;
app.listen(PORT, () => {
  console.log(`🚀 Server running on http://localhost:${PORT}`);
});