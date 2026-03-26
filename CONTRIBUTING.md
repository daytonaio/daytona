# Contributing to Daytona

The team at Daytona welcomes contributions from the community. There are many ways to get involved!

Thanks for taking the time to contribute! ❤️

> And if you like the project but don't have time to contribute, that's perfectly okay. There are other simple ways to support the project and show your appreciation, which we would greatly appreciate:
>
> - Star the project
> - Tweet about it
> - Contribute to our [Docs](https://github.com/daytonaio/docs/)
> - Refer this project in your project's readme
> - Mention the project at local meetups and tell your friends/colleagues

## Code of Conduct

This project and everyone participating in it is governed by the
[Daytona Code of Conduct](https://github.com/daytonaio/daytona?tab=coc-ov-file#readme).
By participating, you are expected to uphold this code. Please report unacceptable behavior
to [info@daytona.io](mailto:info@daytona.io).

## Provide Feedback

You might find things that can be improved while you are using Daytona. You can help by [submitting an issue](https://github.com/daytonaio/daytona/issues/new) when:

- A new feature or an enhancement to an existing feature will improve the utility or usability of Daytona.
- Daytona crashes, or you encounter a bug that can only be resolved by restarting Daytona.
- An error occurs that is unrecoverable, causes Sandbox integrity problems or loss, or generally prevents you from using a Sandbox.

Before creating a new issue, please confirm that an existing issue doesn't already exist.

We will then take care of the issue as soon as possible.

## Participate in the Community

You can engage with our community by:

- Helping other users on [Daytona Community Slack](https://go.daytona.io/slack).
- Improving [documentation](https://github.com/daytonaio/docs/)
- Participating in general discussions about development and DevOps
- Authoring new Daytona Plugins and sharing those Plugins
- Authoring new dev containers and sharing examples

## Contributing Code

You can contribute to Daytona by:

- Enhancing current functionality
- Fixing bugs
- Adding new features and capabilities

Before starting your contribution, especially for core features, we encourage you to reach out to us on [Slack](https://go.daytona.io/slack). This allows us to ensure that your proposed feature aligns with the project's roadmap and goals. Developers are the key to making Daytona the best tool it can be, and we value input from the community.

We look forward to working with you to improve Daytona and make development environments as easy as possible for developers everywhere.

### Steps to Contribute Code

Follow the following steps to ensure your contribution goes smoothly.

1. Read and follow the steps outlined in the [Daytona Contributing Policy](README.md#contributing).
1. Configure your development environment by either following the guide below.
1. [Fork](https://help.github.com/articles/working-with-forks/) the GitHub Repository allowing you to make the changes in your own copy of the repository.
1. Create a [GitHub issue](https://github.com/daytonaio/daytona/issues) if one doesn't exist already.
1. [Prepare your changes](/PREPARING_YOUR_CHANGES.md) and ensure your commits are descriptive. The document contains an optional commit template, if desired.
1. Ensure that you sign off on all your commits to comply with the DCO v1.1. We have more details in [Prepare your changes](/PREPARING_YOUR_CHANGES.md).
1. Ensure to generate new docs after making command related changes, by running `./hack/generate-cli-docs.sh` in the daytona root directory.
1. Ensure to generate a new API client after making changes related to the API spec, by running `./hack/swagger.sh` in the daytona root directory.
1. Ensure that you are using `yarn` as the package manager for any Node.js dependencies.
1. Ensure that you have no lint errors. We use `golangci-lint` as our linter which you can install by following instructions found [here](https://golangci-lint.run/welcome/install/#local-installation) (or simply open Daytona in a Dev Container). You can check for linting errors by running `golangci-lint run` in the root of the project.
1. Create a pull request on GitHub. If you're new to GitHub, read about [pull requests](https://help.github.com/articles/about-pull-requests/). You are welcome to submit your pull request for commentary or review before it is complete by creating a [draft pull request](https://help.github.com/en/articles/about-pull-requests#draft-pull-requests). Please include specific questions or items you'd like feedback on.
1. A member of the Daytona team will review your PR within three business days (excluding any holidays) and either merge, comment, and/or assign someone for review.
1. Work with the reviewer to complete a code review. For each change, create a new commit and push it to make changes to your pull request. When necessary, the reviewer can trigger CI to run tests prior to merging.
1. Once you believe your pull request is ready to be reviewed, ensure the pull request is no longer a draft by [marking it ready for review](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/changing-the-stage-of-a-pull-request).
1. The reviewer will look over your contribution and either approve it or provide comments letting you know if there is anything left to do. We try to give you the opportunity to make the required changes yourself, but in some cases, we may perform the changes ourselves if it makes sense to (minor changes or for urgent issues). We do our best to review PRs promptly, but complex changes could require more time.
1. After completing your review, a Daytona team member will trigger merge to run all tests. Upon passing, your change will be merged into `main`, and your pull requests will be closed. All merges to `main` create a new release, and all final changes are attributed to you.

Note: In some cases, we might decide that a PR should be closed without merging. We'll make sure to provide clear reasoning when this happens.

### Coding Style and Conventions

To make the code base consistent, we follow a few guidelines and conventions listed below.

It is possible that the code base does not currently comply with all these guidelines.
While working on a PR, if you see something that can be refactored to comply, go ahead, but keep in mind that we are not looking for massive PRs that only address that.

API and service method conventions:

1. Avoid using model names in service methods
   - e.g. `Create` instead of `CreateSandbox`, `Find` instead of `FindSandbox`
1. Use appropriate verbs in the UI
   - e.g. `Create API Key` instead of `Generate API Key` since the method is called `Create`
1. Refer to the table below for a connection between API and service methods

| HTTP Method | Controller / Service / Store |
| ----------- | ---------------------------- |
| POST        | Create or Update             |
| DELETE      | Delete                       |
| PUT         | Save                         |
| GET         | Find or List                 |

#### What Does Contributing Mean for You?

Here is what being a contributor means for you:

- License all our contributions to the project under the AGPL 3.0 License or the Apache 2.0 License
- Have the legal rights to license our contributions ourselves, or get permission to license them from our employers, clients, or others who may have them

For more information, see the [README](README.md) and feel free to reach out to us on [Slack](https://go.daytona.io/slack).
