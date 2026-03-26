# Contributing to Daytona Documentation

Thank you for your interest in contributing to Daytona Documentation! Whether you're fixing a typo, improving existing docs, or adding new content, your help is greatly appreciated.

We are happy to provide guidance on PRs, technical writing, and turning any feature idea into a reality.

This document provides a detailed guide for contributors, especially writers, to ensure that contributions are consistent and of high quality. If you need further assistance, don't hesitate to reach out in the [Daytona Slack Community][slack].

## Writing Overview

### Main Guidelines

- **Clarity and Conciseness**: Write clearly and concisely. Avoid complex jargon and aim for simplicity.
- **Second Person Narrative**: Address the reader as "you" to create an engaging and direct narrative.
- **Active Voice**: Use active voice whenever possible to make your writing more dynamic and clear.
- **Screenshots and Examples**: Include annotated screenshots and examples to illustrate complex points.
- **Formatting**: Use Markdown for formatting. Refer to the [Markdown Guide](https://www.markdownguide.org) if you're unfamiliar with it.
- **Code Snippets**: When including code, ensure it's properly formatted and tested.
- **Links**: Add hyperlinks to relevant sections within the docs or to external resources for additional information.

### Writing Process

1. **Familiarize Yourself**: Begin by understanding Daytona and its features. Explore the existing documentation to get a sense of the writing style and structure.
2. **Find a Topic**: Look for topics that need improvement, missing documentation, or new features that haven't been documented yet. You can check existing [issues][issues] for documentation requests or open a new issue if you identify a gap in the content.
3. **Discuss Your Ideas**: If you're addressing an unreported problem or proposing new content, open an issue to discuss your ideas. Provide a clear and concise description of what you want to add or change.
4. **Write**: Make your changes or add new content. Follow the existing documentation format and style guide. Save your files in the correct directories as per the project structure.
5. **Share**: Share your drafts with the community for feedback. Incorporate any suggestions that improve the quality and clarity of the documentation.
6. **Commit and Open a Pull Request**: Commit your changes with clear messages, push them to your fork, and submit a pull request to the main repository for review.
7. **Review**: Wait for a review and merge.

Remember to stay responsive to feedback during the review process and make any necessary revisions based on suggestions from maintainers or other contributors.

## Contributing to Docs 101

### Contributing using Daytona

To contribute using Daytona, follow these steps:

1. **Fork the Docs repo**: Fork the [Daytona Docs repository][sl] to your GitHub account to create an isolated copy where you can work without affecting the original project.
2. **Access Daytona**: Start a new workspace using the GitHub repositor link and Daytona URL of your Daytona instance, for example '<https://YOUR-DAYTONA.INSTANCE/#https://github.com/YOUR-USERNAME/docs>'. If you don't have access to Daytona you can easily [install](https://github.com/daytonaio/installer) it on your own server. Now, start a workspace through a Daytona dashboard. Optionally, you can install a preferred [Daytona extension](https://download.daytona.io/) in your IDE.
3. **Sync Your Fork**: Before making changes, sync your fork with the main repository to ensure you're working with the latest version.
4. **Branch Out**: Create a new branch for your work. Use a descriptive name that reflects the changes you're planning to make.
5. **Make Changes and Test**: Use the preferred IDE to edit, preview, and validate your changes to the documentation.
6. **Commit and Push**: Commit your changes with clear messages and push them to your fork.
7. **Open a Pull Request**: From your fork, submit a pull request to the main repository for review.

### Contributing using Codespaces

To contribute using GitHub Codespaces, follow these steps:

1. **Fork the Docs repo**: Fork the [Daytona Docs repository][sl] to your GitHub account to create an isolated copy where you can work without affecting the original project.
2. **Create a Codespace**: Go to your fork in GitHub to create a new Codespace.
3. **Sync Your Fork**: Before making changes, sync your fork with the main repository to ensure you're working with the latest version.
4. **Branch Out**: Create a new branch for your work. Use a descriptive name that reflects the changes you're planning to make.
5. **Make Changes and Test**: Use the Codespaces to edit, preview, and validate your changes to the documentation.
6. **Commit and Push**: After making changes, commit to your fork and push the updates.
7. **Open a Pull Request**: Create a pull request from your fork to the main Daytona Docs repository for review.

### Contributing using the Local Environment

To set up and contribute using your local environment, follow these steps:

1. **Fork and Clone**: Fork the [Daytona Docs repository][sl] to your GitHub account and clone it to your local machine.
2. **Sync Your Fork**: Before making changes, sync your fork with the main repository to ensure you're working with the latest version.
3. **Branch Out**: Create a new branch for your work. Use a descriptive name that reflects the changes you're planning to make.
4. **Set Up Your Environment**: Ensure you have Node.js (v16 or higher) and pnpm (v8.2 or higher) installed, then install dependencies with `pnpm i`.
5. **Make Changes Locally**: Edit the documentation as needed, following the writing guidelines and style.
6. **Test Your Changes**: Run a local development server to preview your changes.
7. **Commit and Push**: Commit your changes with descriptive messages and push them to your fork.
8. **Create a Pull Request**: Submit a pull request to the main repository for your changes to be reviewed and merged.

## Testing

### Testing visual changes while you work

Run the Astro dev server on the docs site to see how changes you make impact a project using Starlight.

To do this, move into the `docs/` directory and then run `pnpm dev run`:

```sh
cd docs
pnpm dev run
```

You should then be able to open a preview <http://localhost:4321> and see your changes.

> **Note**
> Changes to the Starlight integration will require you to quit and restart the dev server to take effect.

## Git Workflow and Commands Cheat Sheet

This cheat sheet provides the essential Git commands necessary for contributing to the Daytona Documentation as specified in the previous sections.

### Fork and Clone Repository

```sh
# Fork the repository on GitHub to your account using GitHub website

# Clone your forked repository to your local machine, when using Daytona this is done automatically when creating workspace
git clone https://github.com/YOUR-USERNAME/docs.git
cd docs
```

### Sync Your Fork with the Original Repository

```sh
# Add the original repository as an upstream remote
git remote add upstream https://github.com/daytonaio/docs.git

# Fetch the latest changes from upstream
git fetch upstream

# Check out your fork's local default branch (usually 'main')
git checkout main

# Merge changes from upstream/default branch into your local default branch
git merge upstream/main
```

### Create a New Branch for Your Changes

```sh
# Create a new branch and switch to it, e.g. we are updating Gettings Started page
git checkout -b update-getting-started
```

### Make Changes and Commit Them

```sh
# Add all new and modified files to the staging area
git add .

# Commit your changes with a descriptive message
git commit -m "Add a guide on integrating Daytona with VS Code"
```

### Push Changes to Your Fork on GitHub

```sh
# Push your branch to your GitHub fork
git push -u origin update-getting-started
```

### Create a Pull Request

```sh
# Go to the original repository on GitHub
# Click on 'New Pull Request' and select your branch
# Fill out the PR form and submit
```

### Update Your Branch with the Latest Changes from the Main Repository (if needed)

```sh
# Fetch the latest changes from the original repository
git fetch upstream

# Rebase your branch on top of the latest changes from the default branch
git rebase upstream/main

# Force push to update your GitHub fork (use with caution)
git push -f origin update-getting-started
```

### Merge Changes from Main into Your Branch (if there are conflicts after a rebase)

```sh
# Merge changes from the main branch into your feature branch
git merge main

# Resolve any conflicts, then continue with the rebase
git rebase --continue

# Push the changes after resolving conflicts
git push origin update-getting-started
```

Remember to replace `YOUR-USERNAME` with your actual GitHub username and `update-getting-started` with the name of the branch you are working on. Use these commands as a guide to maintain a clean and up-to-date Git workflow.

## Other

### Adding a new language to Starlight’s docs

To add a language, you will need its BCP-47 tag and a label. See [“Adding a new language”](https://github.com/withastro/docs/blob/main/contributor-guides/translating-astro-docs.md#adding-a-new-language) in the Astro docs repo for some helpful tips around choosing these.

- Add your language to the `locales` config in `docs/astro.config.mjs`
- Add your language’s subtag to the i18n label config in `.github/labeler.yml`
- Add your language to the `pa11y` script’s `--sitemap-exclude` flag in `package.json`
- Create the first translated page for your language.  
   This must be the Daytona Docs landing page: `docs/src/content/docs/{language}/index.mdx`.
- Open a pull request on GitHub to add your changes to Daytona Docs!

### Understanding Starlight

- Starlight is built as an Astro integration.
  Read the [Astro Integration API docs][api-docs] to learn more about how integrations work.

  The Starlight integration is exported from [`packages/starlight/index.ts`](./packages/starlight/index.ts).
  It sets up Starlight’s routing logic, parses user config, and adds configuration to a Starlight user’s Astro project.

- For tips and abilities on authoring content in Starlight follow the guide: [https://starlight.astro.build/guides/authoring-content/](https://starlight.astro.build/guides/authoring-content/)

- Most pages in a Starlight project are built using a single [`packages/starlight/index.astro`](./packages/starlight/index.astro) route.
  If you’ve worked on an Astro site before, much of this should look familiar: it’s an Astro component and uses a number of other components to build a page based on user content.

- Starlight consumes a user’s content from the `'docs'` [content collection](https://docs.astro.build/en/guides/content-collections/).
  This allows us to specify the permissible frontmatter via [a Starlight-specific schema](./packages/starlight/schema.ts) and get predictable data while providing clear error messages if a user sets invalid frontmatter in a page.

- Components that require JavaScript for their functionality are all written without a UI framework, most often as custom elements.
  This helps keep Starlight lightweight and makes it easier for a user to choose to add components from a framework of their choice to their project.

- Components that require client-side JavaScript or CSS should use JavaScript/CSS features that are well-supported by browsers.

  You can find a list of supported browsers and their versions using this [browserslist query](https://browsersl.ist/#q=%3E+0.5%25%2C+not+dead%2C+Chrome+%3E%3D+88%2C+Edge+%3E%3D+88%2C+Firefox+%3E%3D+98%2C+Safari+%3E%3D+15.4%2C+iOS+%3E%3D+15.4%2C+not+op_mini+all). To check whether or not a feature is supported, you can visit the [Can I use](https://caniuse.com) website and search for the feature.

[slack]: https://go.daytona.io/slack
[issues]: https://github.com/daytonaio/docs/issues
[sl]: https://github.com/daytonaio/docs/pulls
[api-docs]: https://docs.astro.build/en/reference/integrations-reference/
