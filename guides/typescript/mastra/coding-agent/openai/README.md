# Code Execution Agent

An advanced Mastra template that provides a coding agent capable of planning, writing, executing, and iterating on code in secure, isolated sandbox environments with comprehensive file management and development workflow capabilities.

## Overview

This template demonstrates how to build an AI coding assistant that can work with real development environments. The agent can create sandboxes, manage files and directories, execute code in multiple languages, and monitor development workflows - all within secure, isolated sandbox environments.

## Features

- **Secure Code Execution**: Run Python, JavaScript, and TypeScript code in isolated sandboxes
- **Complete File Management**: Create, read, write, delete files and directories with batch operations
- **Multi-Language Support**: Execute code in Python, JavaScript, and TypeScript environments
- **Live Development Monitoring**: Watch directory changes and monitor development workflows
- **Command Execution**: Run shell commands, install packages, and manage dependencies
- **Memory System**: Persistent conversation memory with semantic recall and working memory
- **Development Workflows**: Professional development patterns with build automation

## Prerequisites

- Node.js 20 or higher
- API key for Daytona ([Daytona](https://www.daytona.io/))
- API key for your chosen model provider

## Setup

1. **Install dependencies:**

   Make sure you are in this project directory before running the install command:

   ```bash
   pnpm install
   ```

2. **Set up environment variables:**

   ```bash
   cp .env.example .env
   # Edit .env and add your API keys
   ```

   **Configure Daytona and your model provider** by setting the API keys:

   ```env
   DAYTONA_API_KEY="your-daytona-api-key-here"
   OPENAI_API_KEY="your-openai-api-key-here"
   ```

3. **Start the development server:**

   ```bash
   pnpm run dev
   ```

## Model Configuration

This template supports any AI model provider through Mastra's model router. You can use models from:

- **OpenAI**: `openai/gpt-4o-mini`, `openai/gpt-4o`
- **Anthropic**: `anthropic/claude-sonnet-4-5-20250929`, `anthropic/claude-haiku-4-5-20250929`
- **Google**: `google/gemini-2.5-pro`, `google/gemini-2.0-flash-exp`
- **Groq**: `groq/llama-3.3-70b-versatile`, `groq/llama-3.1-8b-instant`
- **Cerebras**: `cerebras/llama-3.3-70b`
- **Mistral**: `mistral/mistral-medium-2508`

Set the `MODEL` environment variable in your `.env` file to your preferred model.

## Architecture

### Core Components

#### **Coding Agent** (`src/mastra/agents/coding-agent.ts`)

The main agent with comprehensive development capabilities:

- **Sandbox Management**: Creates and manages isolated execution environments
- **Code Execution**: Runs code with real-time output capture
- **File Operations**: Complete CRUD operations for files and directories
- **Development Monitoring**: Watches for changes and monitors workflows
- **Memory Integration**: Maintains conversation context and project history

#### **Sandbox Tools** (`src/mastra/tools/`)

Complete toolkit for sandbox interaction with Daytona:

**Sandbox Management:**

- `createSandbox` - Initialize new isolated environments
- Connection management with timeout handling

**Code Execution:**

- `runCode` - Execute Python, JavaScript, TypeScript code
- Real-time output capture and error handling
- Environment variable and timeout configuration

**File Operations:**

- `writeFile` - Create individual files
- `writeFiles` - Batch create multiple files for project setup
- `readFile` - Read file contents for analysis and validation
- `listFiles` - Explore directory structures
- `deleteFile` - Clean up files and directories
- `createDirectory` - Set up project structures

**File Information & Monitoring:**

- `getFileInfo` - Get detailed file metadata
- `checkFileExists` - Validate file existence for conditional logic
- `getFileSize` - Monitor file sizes and track changes
- `watchDirectory` - Live monitoring of file system changes

**Development Workflow:**

- `runCommand` - Execute shell commands, build scripts, package management

### Memory System

The agent includes a configured memory system:

- **Thread Management**: Automatic conversation title generation
- **Semantic Recall**: Search through previous interactions
- **Working Memory**: Maintains context across interactions
- **Vector Storage**: Semantic search capabilities with `LibSQLVector`

## Configuration

### Environment Variables

**Sandbox Provider:**

```bash
DAYTONA_API_KEY=your_daytona_api_key_here
```

**Model Provider:**

```bash
OPENAI_API_KEY=your_openai_api_key_here
```

### Customization

You can customize the agent behavior by modifying the instructions in `src/mastra/agents/coding-agent.ts`:

```typescript
export const codingAgent = new Agent({
  id: 'coding-agent',
  name: 'Coding Agent',
  instructions: `
    // Customize agent instructions here
    // Focus on specific languages, frameworks, or development patterns
  `,
  model: openai('gpt-4.1'),
  // ... other configuration
});
```

## Common Issues

### "Sandbox creation failed"

- Check Daytona API key and account status
- Ensure you haven't exceeded sandbox limits
- Verify network connectivity to Daytona services

### "Code execution timeout"

- Increase timeout values for long-running operations
- Break down complex operations into smaller steps
- Monitor resource usage and optimize code

### "File operation errors"

- Validate file paths and permissions
- Check sandbox file system limits
- Ensure directories exist before file operations

### "Agent stopping with tool-call reason"

- Increase `maxSteps` in the agent configuration

## Development

### Project Structure

```text
src/mastra/
      agents/
        coding-agent.ts              # Main coding agent with development capabilities
      tools/
        index.ts                     # Daytona tool exports
        daytona/
          tools.ts                   # Daytona sandbox implementation
          utils.ts                   # Daytona helper functions
      index.ts                       # Mastra configuration with storage and logging
```

## License

See the main project LICENSE file for details.

## References

- [Mastra](https://mastra.ai/)
- [Daytona](https://daytona.io)
