<div align="center">

[![Documentation](https://img.shields.io/github/v/release/daytonaio/docs?label=Docs&color=23cc71)](https://www.daytona.io/docs)
![License](https://img.shields.io/badge/License-AGPL--3-blue)
[![Go Report Card](https://goreportcard.com/badge/github.com/daytonaio/daytona)](https://goreportcard.com/report/github.com/daytonaio/daytona)
[![Issues - daytona](https://img.shields.io/github/issues/daytonaio/daytona)](https://github.com/daytonaio/daytona/issues)
![GitHub Release](https://img.shields.io/github/v/release/daytonaio/daytona)

</div>

&nbsp;

<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/daytonaio/daytona/raw/main/assets/images/Daytona-logotype-white.png">
    <source media="(prefers-color-scheme: light)" srcset="https://github.com/daytonaio/daytona/raw/main/assets/images/Daytona-logotype-black.png">
    <img alt="Daytona logo" src="https://github.com/daytonaio/daytona/raw/main/assets/images/Daytona-logotype-black.png" width="50%">
  </picture>
</div>

<h3 align="center">
  Run AI Code.
  <br/>
  Secure and Elastic Infrastructure for
  Running Your AI-Generated Code.
</h3>

<p align="center">
    <a href="https://www.daytona.io/docs"> Documentation </a>·
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=bug&projects=&template=bug_report.md&title=%F0%9F%90%9B+Bug+Report%3A+"> Report Bug </a>·
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=enhancement&projects=&template=feature_request.md&title=%F0%9F%9A%80+Feature%3A+"> Request Feature </a>·
    <a href="https://go.daytona.io/slack"> Join our Slack </a>·
    <a href="https://x.com/daytonaio"> Connect on X </a>
</p>

<p align="center">
    <a href="https://www.producthunt.com/posts/daytona-2?embed=true&utm_source=badge-top-post-badge&utm_medium=badge&utm_souce=badge-daytona&#0045;2" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/top-post-badge.svg?post_id=957617&theme=neutral&period=daily&t=1746176740150" alt="Daytona&#0032; - Secure&#0032;and&#0032;elastic&#0032;infra&#0032;for&#0032;running&#0032;your&#0032;AI&#0045;generated&#0032;code&#0046; | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>
    <a href="https://www.producthunt.com/posts/daytona-2?embed=true&utm_source=badge-top-post-topic-badge&utm_medium=badge&utm_souce=badge-daytona&#0045;2" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/top-post-topic-badge.svg?post_id=957617&theme=neutral&period=monthly&topic_id=237&t=1746176740150" alt="Daytona&#0032; - Secure&#0032;and&#0032;elastic&#0032;infra&#0032;for&#0032;running&#0032;your&#0032;AI&#0045;generated&#0032;code&#0046; | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>
</p>

---

## Documentation

Daytona provides a website with comprehensive documentation for [SDK](https://www.daytona.io/docs/en/getting-started#sdk), [API](https://www.daytona.io/docs/en/tools/api), and [CLI](https://www.daytona.io/docs/en/tools/cli) references, [guides](https://www.daytona.io/docs/en/guides), and [examples](https://www.daytona.io/docs/en/getting-started#examples).

### Structure

Documentation resides in the `apps/docs` directory, built with Astro and Starlight. The documentation website is deployed in server-rendered (SSR) mode using an Express adapter, with custom middleware handling redirects and routing.

Documentation content is authored in Markdown (`.md`) and Markdown Extended (`.mdx`) files, organized into the following directory structure:

```
apps/docs/
├── astro.config.mjs                 # Astro + Starlight configuration
├── tools/                           # Automation tools
│   ├── update-api-reference.js      # Generate API docs from OpenAPI spec
│   ├── update-cli-reference.js      # Generate CLI docs from CLI source
│   ├── update-llms.js               # Generate llms.txt and llms-full.txt
│   └── update-search.js             # Generate search indexes
├── public/                          # Static assets
├── server/                          # Express middleware and utilities
└── src/
    ├── components/                  # Custom Astro/React components
    ├── content/
    │   ├── config.ts                # Content and sidebar configuration
    │   ├── docs/
    │   │   ├── en/                  # Documentation content
    │   │   │   ├── <filename>.mdx   # Documentation pages
    │   │   │   ├── guides/          # Integration guides
    │   │   │   ├── tools/           # CLI and API reference pages
    │   │   │   ├── typescript-sdk/  # TypeScript SDK reference
    │   │   │   ├── python-sdk/      # Python SDK reference
    │   │   │   ├── ruby-sdk/        # Ruby SDK reference
    │   │   │   └── go-sdk/          # Go SDK reference
    ├── assets/                      # Images, icons, and themes
    ├── styles/                      # Global SCSS stylesheets
    ├── pages/                       # Astro routing pages
    │   ├── index.astro              # Homepage
    │   └── [...slug].astro          # Dynamic documentation pages
    └── utils/                       # Shared utilities
```

### Development

Run the following commands to start the development server and preview changes locally:

```bash
# Install dependencies
yarn
# Start development server
yarn nx serve docs
```

The development server is available on <http://localhost:4321/docs>.

### Build

Run the following commands to build the documentation for local or production deployment:

```bash
npm install @daytona/sdk
```

The processes that occur during the build:

- pages and components compilation, content processing, route and middleware generation
- [search index updates](#update-search) with the latest documentation content
- [LLM-optimized documentation](#update-llms) generation (llms.txt and llms-full.txt)

The generated build output is available in `dist/apps/docs/`.

### Tools

The documentation app includes automation scripts in the `tools/` directory for maintaining auto-generated documentation.

#### Update CLI Reference

Update CLI reference documentation from the Daytona CLI source code by reading YAML documentation files and converting them to MDX documentation files.

Navigate to the `apps/docs` directory and run the following command:

```bash
node tools/update-cli-reference.js
```

- **Input**: `apps/cli/hack/docs/*.yaml` (auto-generated by CLI)
- **Output**: `apps/docs/src/content/docs/en/tools/cli.mdx`

```jsx
import { Daytona } from '@daytona/sdk'

Update API reference documentation from the OpenAPI specification by converting the Swagger/OpenAPI JSON to structured MDX documentation files.

Navigate to the `apps/docs` directory and run the following command:

```bash
node tools/update-api-reference.js
```

- **Input**: `dist/apps/api/openapi.3.1.0.json`
- **Output**: `apps/docs/src/content/docs/en/tools/api.mdx`

#### Update LLMs

Update LLM-optimized documentation (llms.txt and llms-full.txt) from all documentation content by converting MDX/MD files to plain markdown and aggregating them into single files.

Navigate to the `apps/docs` directory and run the following command:

```bash
node tools/update-llms.js
```

- **Input**: `.mdx` and `.md` files in `src/content/docs/en/`
- **Output**: `dist/apps/docs/client/llms.txt` and `dist/apps/docs/client/llms-full.txt`

#### Update search

Update Algolia search indexes with the latest documentation content.

Navigate to the `apps/docs` directory and run the following command:

```bash
node tools/update-search.js
```

## Contributing

Daytona is Open Source under the [GNU AFFERO GENERAL PUBLIC LICENSE](LICENSE), and is the [copyright of its contributors](../../NOTICE). If you would like to contribute to the software, read the Developer Certificate of Origin Version 1.1 (https://developercertificate.org/). Afterwards, navigate to the [contributing guide](CONTRIBUTING.md) to get started.
