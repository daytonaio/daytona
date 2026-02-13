/**
 * Redirect map for pages that changed URLs during the content architecture refactor.
 *
 * The single source of truth used by the Express production server (server/index.mjs)
 * and the Astro middleware (src/middleware.ts).
 */
export const redirects = {
  'inngest-agentkit-coding-agent': 'guides/agentkit/inngest-agentkit-coding-agent',
  'claude-agent-sdk-connect-service-sandbox': 'guides/claude/claude-agent-sdk-connect-service-sandbox',
  'claude-agent-sdk-interactive-terminal-sandbox': 'guides/claude/claude-agent-sdk-interactive-terminal-sandbox',
  'claude-code-run-tasks-stream-logs-sandbox': 'guides/claude/claude-code-run-tasks-stream-logs-sandbox',
  'codex-sdk-interactive-terminal-sandbox': 'guides/codex/codex-sdk-interactive-terminal-sandbox',
  'data-analysis-with-ai': 'guides/data-analysis-with-ai',
  'google-adk-code-generator': 'guides/google-adk-code-generator',
  'langchain-data-analysis': 'guides/langchain/langchain-data-analysis',
  'letta-code-agent': 'guides/letta-code/letta-code-agent',
  'mastra-coding-agent': 'guides/mastra/mastra-coding-agent',
  'opencode-web-agent': 'guides/opencode/opencode-web-agent',
  'recursive-language-models': 'guides/rlm/recursive-language-models',
  'guides/recursive-language-models': 'guides/rlm/recursive-language-models',
  'trl-grpo-training': 'guides/reinforcement-learning/trl-grpo-training',
  'preview-and-authentication': 'preview',
  'regions-and-runners': 'regions',
  'claude': 'guides/claude',
  'computer-use-macos': 'computer-use',
  'computer-use-windows': 'computer-use',
  'computer-use-linux': 'computer-use',
}
