# @daytonaio/sdk is now @daytona/sdk

> **This package has been renamed.** Please use [`@daytona/sdk`](https://www.npmjs.com/package/@daytona/sdk) instead.

## Migration

Update your dependency:

```bash
npm uninstall @daytonaio/sdk
npm install @daytona/sdk
```

or with yarn:

```bash
yarn remove @daytonaio/sdk
yarn add @daytona/sdk
```

Then update your imports:

```diff
- import { Daytona } from '@daytonaio/sdk'
+ import { Daytona } from '@daytona/sdk'
```

The API is identical — only the package name has changed.

## About @daytona/sdk

The official TypeScript SDK for [Daytona](https://daytona.io), secure and elastic infrastructure for running AI-generated code.

For documentation, examples, and guides, visit [daytona.io/docs](https://www.daytona.io/docs/en/typescript-sdk/).
