# Contributing

The team at Daytona welcomes contributions from the community. There are many ways to get involved!

## Provide Feedback

You might find things that can be improved while you are using Daytona. You can help by [submitting an issue](https://github.com/daytonaio/daytona-core-wip/issues/new) when:

* Daytona crashes, or you encounter a bug that can only be resolved by restarting the Daytona server.
* An error occurs that is unrecoverable, causes workspace integrity problems or loss, or generally prevents you from using a workspace.
* A new feature or an enhancement to an existing feature will improve the utility or usability of Daytona.

Before creating a new issue, please confirm that an existing issue doesn't already exist.

## Participate in the Community
You can engage with our community by:

* Helping other users on [Slack](https://join.slack.com/t/daytonacommunity/shared_invite/zt-273yohksh-Q5YSB5V7tnQzX2RoTARr7Q).
* Improving documentation
* Participating in general discussions about development and DevOps
* Authoring new dev containers and sharing examples

## Contributing Code
You can contribute to Daytona by:

* Adding a plugin 
* Enhancing current functionality
* Fixing bugs
* Adding new features and capabilities

Before starting your contribution, especially for core features, we encourage you to reach out to us on Slack. This allows us to ensure that your proposed feature aligns with the project's roadmap and goals. Developers are the key to making Daytona Core the best tool it can be, and we value input from the community.

We look forward to working with you to improve Daytona Core and make remote development environments even more accessible and useful for developers everywhere. 

## Step 1: Make a fork

Fork the repository to your GitHub organization. This means that you'll have a copy of the repository under _your-GitHub-username/repository-name_.

## Step 2: Clone the repository to your local machine

```sh
git clone -b next https://github.com/{your-GitHub-username}/daytona-core-wip.git

```

## Step 3: Prepare the development environment

Set up and run the development environment on your local machine following the [README](./README.md#Building)

## Step 4: Create a branch
Create a new branch for your changes.
In order to keep branch names uniform and easy-to-understand, please use the following conventions for branch naming.
Generally speaking, it is a good idea to add a group/type prefix to a branch.
Here is a list of good examples:
- for docs change : `docs/{ISSUE_NUMBER}-{CUSTOM_NAME}` for e.g. docs/2233-update-contributing-docs
- for new features : `feat/{ISSUE_NUMBER}-{CUSTOM_NAME}` for e.g. feat/1144-add-plugins
- for bug fixes : `fix/{ISSUE_NUMBER}-{CUSTOM_NAME}` for e.g. fix/9878-fix-invite-wrong-url
- for anything else: `chore/{ISSUE_NUMBER}-{CUSTOM_NAME}` for e.g. chore/111-update-ci-url

```sh
git checkout -b branch-name-here
```

## Step 5: Make your changes

Update the code with your bug fix or new feature.

## Step 6: Add the changes that are ready to be committed

Stage the changes that are ready to be committed:

```sh
git add .
```

## Step 7: Commit the changes (Git)

Commit the changes with a short message. (See below for more details on how we structure our commit messages)

```sh
git commit -m "<type>(<package>): <subject>"
```

## Step 8: Push the changes to the remote repository

Push the changes to the remote repository using:

```sh
git push origin branch-name-here
```

## Step 9: Create Pull Request

In GitHub, do the following to submit a pull request to the upstream repository:

1. Add yourself to the [Contributors](./CONTRIBUTORS.md) list. You can add this as a separate commit but its needed before your issue can be reviewed.

1. Give the pull request a title and a short description of the changes made following the template. Include also the issue or bug number associated with your change. Explain the changes that you made, any issues you think exist with the pull request you made, and any questions you have for the maintainer.  <br/> ⚠️ **Make sure your pull request target the `next` branch.**
 
  > Pull request title should be in the form of `<type>(<package>): <subject>` as per commit messages.
Remember, it's okay if your pull request is not perfect (no pull request ever is). The reviewer will be able to help you fix any problems and improve it!

1. Wait for the pull request to be reviewed by a maintainer.

1. Make changes to the pull request if the reviewing maintainer recommends them.

Celebrate your success after your pull request is merged.

## Git Commit Messages

We structure our commit messages like this:

```
<type>(<package>): <subject>
```

Example

```
fix(server): missing entity on init
```

### Types:

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Changes to the documentation
- **style**: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc.)
- **refactor**: A code change that neither fixes a bug nor adds a feature
- **perf**: A code change that improves performance
- **test**: Adding missing or correcting existing tests
- **chore**: Changes to the build process or auxiliary tools and libraries such as documentation generation

### Packages:

TODO
- **server**
- **client**
- **extensions**
- **cmd**
