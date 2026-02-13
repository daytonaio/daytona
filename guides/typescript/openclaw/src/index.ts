/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import "dotenv/config";
import { Daytona } from "@daytonaio/sdk";
import type { Sandbox } from "@daytonaio/sdk";
import { randomBytes } from "crypto";
import { readFileSync } from "fs";
import { join } from "path";
import { deepMerge, readEnvFile } from "./utils.js";

// Constants
const OPENCLAW_PORT = 18789; // OpenClaw Gateway and Control UI port
const SHOW_LOGS = true; // Stream OpenClaw stdout/stderr to the terminal
const MAKE_PUBLIC = true; // Expose the sandbox for public internet access
const PERSIST_SANDBOX = true; // Keep the sandbox running after the script exits
const DAYTONA_SNAPSHOT = "daytona-medium"; // This snapshot has openclaw installed

// Paths
const USER_CONFIG_PATH = join(process.cwd(), "openclaw.json");
const ENV_SANDBOX_PATH = join(process.cwd(), ".env.sandbox");

// Global variables
let currentSandbox: Sandbox | null = null;
let sandboxDeleted = false;

// Shutdown the sandbox
async function shutdown() {
  if (sandboxDeleted) return;
  sandboxDeleted = true;
  if (!PERSIST_SANDBOX) {
    console.log("\nShutting down sandbox...");
    try {
      await currentSandbox?.delete(30);
    } catch (e) {
      console.error(e);
    }
  } else {
    console.log("\nSandbox left running.");
  }
  process.exit(0);
}

// OpenClaw config to run in a Daytona sandbox
const OPENCLAW_CONFIG = {
  gateway: {
    mode: "local" as const,
    port: OPENCLAW_PORT,
    bind: "lan" as const,
    auth: { mode: "token" as const, token: "" },
    controlUi: { allowInsecureAuth: true }, // This bypasses the pairing step in the Control UI
  },
  agents: {
    defaults: {
      workspace: "~/.openclaw/workspace",
    },
  },
};

// Main function
async function main() {
  // Create a new Daytona instance
  const daytona = new Daytona();

  // Create a new sandbox
  console.log("Creating Daytona sandbox...");
  const sandbox = await daytona.create({
    snapshot: DAYTONA_SNAPSHOT,
    autoStopInterval: 0,
    envVars: readEnvFile(ENV_SANDBOX_PATH),
    public: MAKE_PUBLIC,
  });
  currentSandbox = sandbox;

  // Handle SIGINT
  process.on("SIGINT", () => shutdown());

  // Get the user home directory
  const home = await sandbox.getUserHomeDir();
  const openclawDir = `${home}/.openclaw`;

  // Read the user config and merge it with the base config
  const userConfig = JSON.parse(readFileSync(USER_CONFIG_PATH, "utf8"));
  const baseConfig = deepMerge(OPENCLAW_CONFIG, userConfig);

  // Generate a random gateway token and add it to the config
  const gatewayToken = randomBytes(24).toString("hex");
  const config = deepMerge(baseConfig, {
    gateway: {
      auth: { mode: "token" as const, token: gatewayToken },
    },
  });

  // Write the config to the sandbox
  console.log("Configuring OpenClaw...");
  await sandbox.fs.createFolder(openclawDir, "755");
  await sandbox.fs.uploadFile(
    Buffer.from(JSON.stringify(config, null, 2), "utf8"),
    `${openclawDir}/openclaw.json`,
  );

  // Start the gateway
  const sessionId = "openclaw-gateway";
  console.log("Starting OpenClaw...");
  await sandbox.process.createSession(sessionId);
  const { cmdId } = await sandbox.process.executeSessionCommand(sessionId, {
    command: "openclaw gateway run",
    runAsync: true,
  });
  console.log("(Ctrl+C to shut down and delete the sandbox)");

  // Stream OpenClaw output to the terminal and delete the sandbox when the process ends
  sandbox.process
    .getSessionCommandLogs(
      sessionId,
      cmdId,
      SHOW_LOGS ? (chunk) => process.stdout.write(chunk) : () => {},
      SHOW_LOGS ? (chunk) => process.stderr.write(chunk) : () => {},
    )
    .then(shutdown)
    .catch(shutdown);

  const signed = await sandbox.getPreviewLink(OPENCLAW_PORT);
  const dashboardUrl = `${signed.url}?token=${gatewayToken}`;
  
  console.log(`\n\x1b[1m🔗 Secret link to Control UI: ${dashboardUrl}\x1b[0m`);
  console.log(`\nOpenClaw is starting...`);
  console.log("--------------------------------");
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});