let dotenvConfigResult: { parsed?: Record<string, string>; error?: Error } | null = null;
try {
  // optional dependency: if dotenv is missing, continue with process.env as-is
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  const dotenv = require("dotenv");
  dotenvConfigResult = dotenv.config();
  if (dotenvConfigResult.error) {
    console.warn("dotenv config failed:", dotenvConfigResult.error.message);
  }
} catch (err) {
  console.warn("dotenv not installed; skipping load. Set env vars externally.", err);
}

// Ensure Node globals are typed when using plain TypeScript setup
interface ProcessLike {
  env: Record<string, string | undefined>;
  stdout: { write(chunk: string): void };
  exit(code?: number): never;
}

declare var process: ProcessLike;

// ─── Mock API Layer ───────────────────────────────────────────────────────────
// In production, replace these with: import { Daytona } from '@daytonaio/sdk'

const DAYTONA_API_KEY = process.env.DAYTONA_API_KEY || "mock-daytona-api-key";
if (!process.env.DAYTONA_API_KEY) {
  console.warn("DAYTONA_API_KEY not found in environment; using mock key for local testing.");
}

// Mock types
interface Sandbox {
  id: string;
  state: "running" | "stopped" | "failed" | "timed_out" | "completed";
  logs_url: string;
  results?: {
    ai_agent_tests?: { passed: boolean; total: number; failed: number };
  };
}

interface Organization {
  id: string;
  name: string;
}

interface User {
  id: string;
  email: string;
}

// Mock Daytona SDK class (mirrors real @daytonaio/sdk interface)
class MockDaytona {
  private apiKey: string;

  constructor(config: { apiKey: string }) {
    this.apiKey = config.apiKey;
    console.log("[MockDaytona] Initialized with API key:", this.apiKey.slice(0, 6) + "...");
  }

  organizations = {
    listOrganizations: async (): Promise<Organization[]> => {
      await sleep(300);
      return [
        { id: "org-001", name: "Acme Corp" },
        { id: "org-002", name: "Dev Team" },
      ];
    },
  };

  users = {
    listUsersInOrganization: async (orgId: string): Promise<User[]> => {
      await sleep(200);
      return [
        { id: "user-001", email: "alice@acme.com" },
        { id: "user-002", email: "bob@acme.com" },
      ];
    },
  };

  sandbox = {
    startSandbox: async (opts: {
      organizationId: string;
      template?: { image: string };
      timeout_minutes?: number;
      resources?: object;
      commands?: string[];
    }): Promise<Sandbox> => {
      await sleep(500);
      console.log("[MockDaytona] Sandbox provisioning for org:", opts.organizationId);
      return {
        id: `sandbox-${Math.random().toString(36).slice(2, 9)}`,
        state: "running",
        logs_url: "https://mock.daytona.ai/logs/sandbox-abc123",
      };
    },

    getSandboxStatus: async (sandboxId: string): Promise<Sandbox> => {
      await sleep(200);
      // Simulate progression to completed after first call
      return {
        id: sandboxId,
        state: "completed",
        logs_url: "https://mock.daytona.ai/logs/" + sandboxId,
        results: {
          ai_agent_tests: { passed: true, total: 5, failed: 0 },
        },
      };
    },

    streamExec: async (
      sandboxId: string,
      opts: { cmd: string; args: string[] }
    ): Promise<{ stream: AsyncIterable<string> }> => {
      async function* mockStream() {
        const lines = [
          `> Running: ${opts.cmd} ${opts.args.join(" ")}\n`,
          "Hello from Daytona sandbox!\n",
          "All checks passed.\n",
        ];
        for (const line of lines) {
          await sleep(150);
          yield line;
        }
      }
      return { stream: mockStream() };
    },

    installPackages: async (
      sandboxId: string,
      packages: { nodejs?: string[] }
    ): Promise<void> => {
      await sleep(400);
      console.log(`[MockDaytona] Installed packages in ${sandboxId}:`, packages.nodejs);
    },

    createFile: async (
      sandboxId: string,
      path: string,
      content: string
    ): Promise<void> => {
      await sleep(200);
      console.log(`[MockDaytona] Created file at ${path} in sandbox ${sandboxId}`);
    },

    deleteSandbox: async (sandboxId: string): Promise<boolean> => {
      await sleep(300);
      console.log(`[MockDaytona] Deleted sandbox: ${sandboxId}`);
      return true;
    },
  };
}

// ─── Utilities ────────────────────────────────────────────────────────────────

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

const TERMINAL_STATES = ["completed", "failed", "timed_out"];

async function pollUntilDone(
  daytona: MockDaytona,
  sandboxId: string,
  intervalMs = 15_000
): Promise<Sandbox> {
  while (true) {
    let status: Sandbox;
    try {
      status = await daytona.sandbox.getSandboxStatus(sandboxId);
    } catch (err) {
      console.warn("Polling error, retrying in 15s...", err);
      await sleep(intervalMs);
      continue;
    }

    console.log(`[Poll] Sandbox state: ${status.state}`);

    if (TERMINAL_STATES.includes(status.state)) {
      return status;
    }

    await sleep(intervalMs);
  }
}

// ─── Main Flow ────────────────────────────────────────────────────────────────

async function main() {
  const daytona = new MockDaytona({ apiKey: DAYTONA_API_KEY! });

  // Step 1: List organizations and pick first
  console.log("\n── Step 1: Fetch Organizations ──");
  const orgs = await daytona.organizations.listOrganizations();
  console.log("Organizations:", orgs.map((o) => o.name));
  const org = orgs[0];

  // Step 2: List users in org
  console.log("\n── Step 2: Fetch Users ──");
  const users = await daytona.users.listUsersInOrganization(org.id);
  console.log("Users:", users.map((u) => u.email));

  // Step 3: Start sandbox
  console.log("\n── Step 3: Start Sandbox ──");
  const sandbox = await daytona.sandbox.startSandbox({
    organizationId: org.id,
    template: { image: "mcr.microsoft.com/devcontainers/typescript-node" },
    timeout_minutes: 45,
    resources: { cpu: "2vCPU", memory: "4Gi" },
    commands: ["npm ci", "npm run test:ai-agent"],
  });
  console.log("Sandbox started:", sandbox.id);

  // Step 4: Stream exec output
  console.log("\n── Step 4: Stream Exec ──");
  const { stream } = await daytona.sandbox.streamExec(sandbox.id, {
    cmd: "node",
    args: ["-e", 'console.log("Hello from Daytona sandbox!")'],
  });
  for await (const chunk of stream) {
    process.stdout.write(chunk);
  }

  // Step 5: Install packages and create a file
  console.log("\n── Step 5: Install Packages & Create File ──");
  await daytona.sandbox.installPackages(sandbox.id, { nodejs: ["express"] });
  await daytona.sandbox.createFile(
    sandbox.id,
    "/app/server.js",
    `const express = require('express');
const app = express();
app.get('/', (req, res) => res.send('API ready!'));
app.listen(3000);`
  );

  // Step 6: Poll until done
  console.log("\n── Step 6: Poll Until Complete ──");
  const finalStatus = await pollUntilDone(daytona, sandbox.id, 1_000); // 1s for mock
  console.log("Logs URL:", finalStatus.logs_url);

  // Step 7: Evaluate results
  console.log("\n── Step 7: Evaluate Results ──");
  const testResults = finalStatus.results?.ai_agent_tests;
  if (testResults?.passed) {
    console.log(`✓ AI agent tests PASSED (${testResults.total} total, ${testResults.failed} failed)`);
  } else {
    console.error("✗ AI agent tests FAILED", testResults);
    await daytona.sandbox.deleteSandbox(sandbox.id);
    process.exit(1);
  }

  // Step 8: Cleanup
  console.log("\n── Step 8: Cleanup ──");
  await daytona.sandbox.deleteSandbox(sandbox.id);
  console.log("Done.");
}

main().catch((err) => {
  console.error("Unhandled error:", err);
  process.exit(1);
});

export {};