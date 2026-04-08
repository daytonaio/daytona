# @daytonaio/opencode is now @daytona/opencode

> **This package has been renamed.** Please use [`@daytona/opencode`](https://www.npmjs.com/package/@daytona/opencode) instead.

## Migration

Update your OpenCode configuration:

```diff
{
  "$schema": "https://opencode.ai/config.json",
- "plugin": ["@daytonaio/opencode"]
+ "plugin": ["@daytona/opencode"]
}
```

The plugin is identical — only the package name has changed.

## About @daytona/opencode

An OpenCode plugin that automatically runs all sessions in Daytona sandboxes for isolated, reproducible development environments.

For documentation and setup instructions, see the [@daytona/opencode README](https://www.npmjs.com/package/@daytona/opencode).
