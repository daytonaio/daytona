/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

// Builds the Daytona-specific instructions dropped into the sandbox's
// `.opencode/instructions/daytona.md`, which opencode loads via opencode.json.
export function buildSandboxInstructions({
  repoPath,
  previewUrlTemplate,
}: {
  repoPath: string
  previewUrlTemplate: string
}): string {
  // Build an example URL by replacing <port> with 8000
  const exampleUrl = previewUrlTemplate.replace('<port>', '8000')

  return `## Daytona Sandbox Integration
This session is integrated with a Daytona sandbox.
The main project repository is located at: ${repoPath}

### Running Servers
When starting long-running processes like servers, use \`nohup\` to prevent them from being killed when the bash command times out:
\`\`\`bash
nohup <command> > /tmp/server.log 2>&1 &
\`\`\`
For example:
\`\`\`bash
nohup python3 -m http.server 8000 > /tmp/http-server.log 2>&1 &
\`\`\`

### Preview URLs
Before showing a preview URL, ensure the server is running in the sandbox on that port.
To access a running server from a browser, use the Daytona proxy URL format:
\`\`\`
${previewUrlTemplate}
\`\`\`
For example, if a server is running on port 8000:
\`\`\`
${exampleUrl}
\`\`\`
`
}
