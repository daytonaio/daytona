<br>

<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/daytonaio/daytona/raw/main/assets/images/Daytona-logotype-white.png">
    <img alt="Daytona logo" src="https://github.com/daytonaio/daytona/raw/main/assets/images/Daytona-logotype-black.png" width="40%">
  </picture>
</div>

<br>

<div align="center">

[![Documentation](https://img.shields.io/github/v/release/daytonaio/docs?label=Docs&color=23cc71)](https://www.daytona.io/docs)
[![License](https://img.shields.io/badge/License-Apache--2.0-blue)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/daytonaio/daytona)](https://goreportcard.com/report/github.com/daytonaio/daytona)
[![Issues - daytona](https://img.shields.io/github/issues/daytonaio/daytona)](https://github.com/daytonaio/daytona/issues)
![GitHub Release](https://img.shields.io/github/v/release/daytonaio/daytona)
<br>
[![Open Bounties](https://img.shields.io/endpoint?url=https%3A%2F%2Fconsole.algora.io%2Fapi%2Fshields%2Fdaytonaio%2Fbounties%3Fstatus%3Dopen)](https://console.algora.io/org/daytonaio/bounties?status=open)
[![Rewarded Bounties](https://img.shields.io/endpoint?url=https%3A%2F%2Fconsole.algora.io%2Fapi%2Fshields%2Fdaytonaio%2Fbounties%3Fstatus%3Dcompleted)](https://console.algora.io/org/daytonaio/bounties?status=completed)

<br>

<a href="https://www.producthunt.com/posts/daytona?utm_source=badge-top-post-badge&utm_medium=badge&utm_souce=badge-daytona" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/top-post-badge.svg?post_id=445392&theme=light&period=daily" alt="Daytona - Dev&#0032;environment&#0032;manager&#0032;that&#0032;makes&#0032;you&#0032;2x&#0032;more&#0032;productive | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>
<a href="https://www.producthunt.com/posts/daytona?utm_source=badge-top-post-topic-badge&utm_medium=badge&utm_souce=badge-daytona" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/top-post-topic-badge.svg?post_id=445392&theme=light&period=weekly&topic_id=267" alt="Daytona - Dev&#0032;environment&#0032;manager&#0032;that&#0032;makes&#0032;you&#0032;2x&#0032;more&#0032;productive | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>

</div>


<h1 align="center">The Open Source Development Environment Manager</h1>
<div align="center">
Set up a development environment on any infrastructure, with a single command.
</div>
</br>


<p align="center">
    <a href="https://www.daytona.io/docs">Documentation</a>
    ·
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=bug&projects=&template=bug_report.md&title=%F0%9F%90%9B+Bug+Report%3A+">Report Bug</a>
    ·
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=enhancement&projects=&template=feature_request.md&title=%F0%9F%9A%80+Feature%3A+">Request Feature</a>
    ·
  <a href="https://go.daytona.io/slack">Join Our Slack</a>
    ·
    <a href="https://twitter.com/daytonaio">Twitter</a>
  </p>

<div align="center"><img src="https://github.com/daytonaio/daytona/raw/main/assets/images/daytona_demo.gif" width="50%" ></div>

## Features
* __Single Command__: Activate a fully configured development environment with a single command.
* __Runs everywhere__: spin up your development environment on any machine — whether it's local, remote, cloud-based, physical server, or a VM & any architecture x86 or ARM.
* __Configuration File Support__: Initially support for [dev container](https://containers.dev/), ability to expand to DevFile, Nix & Flox (Contributions welcome here!).
* __Prebuilds System__: Drastically improve environment setup times (Contributions welcome here!).
* __IDE Support__ : Seamlessly supports [VS Code](https://github.com/microsoft/vscode) & [JetBrains](https://www.jetbrains.com/remote-development/gateway/) locally, ready to use without configuration. Includes a built-in Web IDE for added convenience.
* __Git Provider Integration__: GitHub, GitLab, Bitbucket, Bitbucket Server, Gitea, Gitness, Azure DevOps, AWS CodeCommit & Gogs can be connected, allowing easy repo branch or PR pull and commit back from the workspaces.
* __Multiple Project Workspace__: Support for multiple project repositories in the same workspace, making it easy to develop using a micro-service architecture.
* __Reverse Proxy Integration__: Enable collaboration and streamline feedback loops by leveraging reverse proxy functionality. Access preview ports and the Web IDE seamlessly, even behind firewalls.
* __Extensibility__: Enable extensibility with plugin or provider development. Moreover, in any dynamic language, not just Go(Contributions welcome here!).
* __Security__: Automatically creates a VPN connection between the client machine and the development environment, ensuring a fully secure connection.
* __All Ports__: The VPN connection enables access to all ports on the development environments, removing the need to setup port forwards over SSH connection.
* __Works on my Machine__: Never experience it again.
## Quick Start
### Mac / Linux
```bash
curl -sfL https://download.daytona.io/daytona/install.sh | sudo bash && daytona server -y && daytona
```
### Windows
<details>
<summary>Windows PowerShell</summary>
This command downloads and installs Daytona and runs the Daytona Server:

```pwsh
$architecture = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" } else { "arm64" }
md -Force "$Env:APPDATA\bin\daytona"; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]'Tls,Tls11,Tls12';
Invoke-WebRequest -URI "https://download.daytona.io/daytona/latest/daytona-windows-$architecture.exe" -OutFile "$Env:APPDATA\bin\daytona\daytona.exe";
$env:Path += ";" + $Env:APPDATA + "\bin\daytona"; [Environment]::SetEnvironmentVariable("Path", $env:Path, [System.EnvironmentVariableTarget]::User);
daytona serve;
```

</details>

### Create your first dev environment by opening a new terminal, and running:

```bash
daytona create
```

**Start coding.**

----

</br>

## Why Daytona?
Daytona is a radically simple open source development environment manager.

Setting up development environments has become increasingly challenging over time, especially when aiming to set up remotely, where the complexity increases by an order of magnitude. The process is so complex that we've compiled a [comprehensive guide](https://www.daytona.io/dotfiles/diy-guide-to-transform-any-machine-into-a-codespace) detailing all the necessary steps to set one up—spanning __5,000 words__, __7 steps__, and requiring anywhere from 15 to __45 minutes__.

This complexity is unnecessary.

With Daytona, you need only to execute a single command: `daytona create`.

Daytona automates the entire process; provisioning the instance, interpreting and applying the configuration, setting up prebuilds, establishing a secure VPN connection, securely connecting your local or a Web IDE, and assigning a fully qualified domain name to the development environment for easy sharing and collaboration.

As a developer, you can immediately start focusing on what matters most—your code.



## Backstory
We spent most of our careers building cloud development environments. In 2009, we launched what was likely the first commercial [Cloud IDE](https://codeanywhere.com) project. At that time, technology was lacking, forcing us to develop everything from scratch—the IDE, the environment orchestrator, and almost everything else. A lot of people were interested, and over 2.5 million developers signed up! But we were too early, and we asked too much from our users to change how they worked.

Now, 15 years since its inception, we have noticed quite a few things. First, the technology we wished for back then exists now. Second, approximately 50% of developers work in remote dev environments, and third, and most importantly, setting up development environments has become more complex than ever, both locally and to a greater magnitude for remote.

So, we took everything we learned and decided to solve these issues once and for all as a fully open-source project. Our goal was to create a single binary that allows you to set up a development environment anywhere you wish, completely free, and finally fulfill the promise that many have attempted to make.






## Getting Started
### Requirements
Before starting the installation script, please go over all the necessary requirements:
- __Hardware Resources__: Depending on the project requirements, ensure your machine has sufficient resources. Minimum hardware specification is 1cpu, 2GB of RAM and 10GB of disk space.
- __Docker__: Ensure [Docker](https://www.docker.com/products/docker-desktop/) is installed and running.



### Installing Daytona
Daytona allows you to manage your Development Environments using the Daytona CLI. To install it, please execute the following command:

```bash
# Install Daytona into /usr/local/bin
curl -sf -L https://download.daytona.io/daytona/install.sh | sudo bash

# OR if you want to install Daytona to some other path where you don`t need sudo
# curl -sf -L https://download.daytona.io/daytona/install.sh | DAYTONA_PATH=/home/user/bin bash
```
<details open>
  <summary> Manual installation </summary>
  If you don't want to use the provided script, download the binary directly from the URL for your specific OS:

  ```bash
  curl -sf -L https://download.daytona.io/daytona/latest/daytona-darwin-amd64 -o daytona
  curl -sf -L https://download.daytona.io/daytona/latest/daytona-darwin-arm64 -o daytona
  curl -sf -L https://download.daytona.io/daytona/latest/daytona-linux-amd64 -o daytona
  curl -sf -L https://download.daytona.io/daytona/latest/daytona-linux-arm64 -o daytona
  curl -sf -L https://download.daytona.io/daytona/latest/daytona-windows-amd64.exe -o daytona
  curl -sf -L https://download.daytona.io/daytona/latest/daytona-windows-arm64.exe -o daytona
  ```
  Make sure that path where `daytona` binary is downloaded is in your system PATH.
</details>



### Initializing Daytona
To initialize Daytona, follow these steps:

__1. Start the Daytona Server:__
This initiates the Daytona Server in daemon mode. Use the command:
```bash
daytona server
```
__2. Add Your Git Provider of Choice:__
Daytona supports GitHub, GitLab, Bitbucket, Bitbucket Server, Gitea, Gitness, AWS CodeCommit, Azure DevOps and Gogs. To add them to your profile, use the command:
```bash
daytona git-providers add

```
Follow the steps provided.

__3. Add Your Provider Target:__
This step is for choosing where to deploy Development Environments. By default, Daytona includes a Docker provider to spin up environments on your local machine. For remote development environments, use the command:
```bash
daytona target set
```
Following the steps this command adds SSH machines to your targets.

__4. Choose Your Default IDE:__
The default setting for Daytona is VS Code locally. If you prefer, you can switch to VS Code - Browser or any IDE from the JetBrains portfolio using the command:
```bash
daytona ide
```
Now that you have installed and initialized Daytona, you can proceed to setting up your development environments and start coding instantly.





### Creating Dev Environments
Creating development environments with Daytona is a straightforward process, accomplished with just one command:
```bash
daytona create
```

You can add the `--no-ide` flag if you don't wish to open the IDE immediately after creating the environment.

Upon executing this command, you will be prompted with two questions:
1. Choose the provider to decide where to create a dev environment.
2. Select or type the Git repository you wish to use to create a dev environment.

After making your selections, press enter, and Daytona will handle the rest. All that remains for you to do is to execute the following command to open your default IDE:
```bash
daytona code
```

This command opens your development environment in your preferred IDE, allowing you to start coding instantly.

### Stopping the Daytona Server:
```bash
daytona server stop
```

### Restarting the Daytona Server:
```bash
daytona server restart
```

## How to Extend Daytona

Daytona offers flexibility for extension through the creation of plugins and providers.


### Providers
Daytona is designed to be infrastructure-agnostic, capable of creating and managing development environments across various platforms. Providers are the components that encapsulate the logic for provisioning compute resources on a specific target platform. They allow for the configuration of different targets within a single provider, enabling, for instance, multiple AWS profiles within an AWS provider.

How does it work? When executing the `daytona create` command, Daytona communicates the environment details to the selected provider, which then provisions the necessary compute resources. Once provisioned, Daytona sets up the environment on these resources, allowing the user to interact with the environment seamlessly.

Providers are independent projects that adhere to the Daytona Provider interface. They can be developed in nearly any major programming language. More details coming soon.


### Plugins
Plugins enhance Daytona's core functionalities by adding new CLI commands, API methods, or services within the development environments. They offer configurable settings to tailor the plugin's behavior to the user's needs.

Similar to providers, plugins are independent projects that conform to the Daytona Plugin interface and can be developed in a wide range of programming languages. More details coming soon.




## Contributing To Daytona

We welcome contributions to Daytona! Whether you're fixing bugs, improving documentation, suggesting new features, or reporting issues, your help is greatly appreciated.

### Open Source Licensing

Daytona is Open Source under the [Apache License 2.0](LICENSE), and is the [copyright of its contributors](NOTICE). 

If you would like to contribute to the software, you must:

  1. **Read the Developer Certificate of Origin Version 1.1** 

      Please review the [Developer Certificate of Origin Version 1.1](https://developercertificate.org/) to understand the contribution requirements.

  2. **Signing all commits to the Daytona project** 

      Ensure that all your commits are signed to comply with the Daytona project's contribution policies.

      This ensures that users, distributors, and other contributors can rely on all the software related to Daytona being contributed under the terms of the [Apache License 2.0](LICENSE). No contributions will be accepted without following this process.

### Ways to Contribute

### 1. Reporting Issues and Suggesting Features

Creating issues is a valuable way to contribute by reporting bugs, suggesting features, or improving documentation.

Before creating a new issue, search the existing issues [here](https://github.com/daytonaio/daytona/issues) to see if your concern has already been addressed.

* If no existing issue matches your contribution, follow these steps:
   1. **Identify the Type of Issue**
      * **Bug Report:** If you encounter unexpected behavior or errors.
      * **Feature Request:** If you have an idea for a new feature or improvement.
      * **Documentation Improvement:** If you notice gaps or areas for improvement in the documentation.

   2. **Create a new issue**
      * Navigate to Issues: Go to the Issues tab [here](https://github.com/daytonaio/daytona/issues).
      * Click on "New Issue": Choose the appropriate template (Bug Report, Feature Request, etc.) if available.
      * Fill Out the Issue Template: Provide a clear and concise description of the issue, including steps to reproduce (for bugs) or detailed feature descriptions.
      * Submit the Issue: Click "Submit new issue" to create the issue.
   3. Engage with the Community
      * **Respond to Feedback:** Be prepared to provide additional information or clarification if maintainers or other contributors have questions.
      * **Collaborate on Solutions:** If you have ideas for resolving the issue, share them in the comments.

### 2. Contributing Code

If you're interested in contributing code to Daytona, follow these steps:

1. **Fork the Daytona repository**

    [Fork](https://github.com/daytonaio/daytona/fork) the GitHub repository to create your own copy of the repository.
  
2. **Create a Workspace with Daytona**

    Use the Daytona CLI to create a workspace for your forked repository. Replace YOUR-FORK-URL with the URL of your forked repository.
    ```bash
    daytona create YOUR-FORK-URL
    ```
3.  **Create a new branch**
    
    Once in the development conatiner, create a new branch for your changes:
    ```bash
    git checkout -b my-new-feature
    ```

4. **Run the Project Locally**
    ```bash
    dtn serve
    ```

5. **Make changes to the project**

    Prepare your changes and ensure your commits are descriptive. The document contains an optional commit template, if desired.

6. **Test your changes**

    Ensure to test your changes by running the project locally. You can also run the tests by running `go test ./...` in the root of the project.

7. **Generate docs**

    Ensure to generate new docs after making command related changes, by running ./hack/generate-cli-docs.sh in the daytona root directory.

    ```bash
    ./hack/generate-cli-docs.sh
    ```

8. **Generate new API client**

    Ensure to generate a new API client after making changes related to the API spec, by running ./hack/swagger.sh in the daytona root directory.

    ```bash
    ./hack/swagger.sh
    ```

9. **Check for lint errors**

    Ensure that you have no lint errors. We use golangci-lint as our linter which you can install by following instructions found [here](https://golangci-lint.run/usage/install/) (or simply open Daytona in a Dev Container). You can check for linting errors by running golangci-lint run in the root of the project.

    ```bash
    golangci-lint run
    ```

10. **Sign off on your commits**

    Ensure that you sign off on all your commits to comply with the DCO v1.1. We have more details in Prepare your changes.

11. **Push your changes and create a pull request**

    Push your changes to your forked repository and create a pull request from your branch in your forked repository to the main Daytona repository.
    If you're new to GitHub, read about [pull requests](https://help.github.com/articles/about-pull-requests/). You are welcome to submit your pull request for commentary or review before it is complete by creating a [draft pull request](https://help.github.com/en/articles/about-pull-requests#draft-pull-requests). Please include specific questions or items you'd like feedback on.

12. **Wait for review**

    A Daytona team member will review your PR within three business days (excluding any holidays) and either merge, comment, and/or assign someone for review.

## License

This repository contains Daytona, covered under the [Apache License 2.0](LICENSE), except where noted (any Daytona logos or trademarks are not covered under the Apache License, and should be explicitly noted by a LICENSE file.)

Daytona is a product produced from this open source software, exclusively by Daytona Platforms, Inc. It is distributed under our commercial terms.

Others are allowed to make their own distribution of the software, but they cannot use any of the Daytona trademarks, cloud services, etc.

We explicitly grant permission for you to make a build that includes our trademarks while developing Daytona itself. You may not publish or share the build, and you may not use that build to run Daytona for any other purpose.

You can read more in our [packinging guidelines](PACKAGING.md).

## Code of Conduct


This project has adapted the Code of Conduct from the [Contributor Covenant](https://www.contributor-covenant.org/). For more information see the [Code of Conduct](CODE_OF_CONDUCT.md) or contact [codeofconduct@daytona.io.](mailto:codeofconduct@daytona.io) with any additional questions or comments.

## Questions


For more information on how to use and develop Daytona, talk to us on
[Slack](https://go.daytona.io/slack).
