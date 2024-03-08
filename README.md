<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/daytonaio/daytona/raw/main/assets/images/Daytona-logotype-white.png">
    <img alt="Daytona logo" src="https://github.com/daytonaio/daytona/raw/main/assets/images/Daytona-logotype-black.png" width="40%">
  </picture>
</div>

<br><br>

<div align="center">

[![License](https://img.shields.io/badge/License-Apache--2.0-blue)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/daytonaio/daytona)](https://goreportcard.com/report/github.com/daytonaio/daytona)
[![Issues - daytona](https://img.shields.io/github/issues/daytonaio/daytona)](https://github.com/daytonaio/daytona/issues)
![GitHub Release](https://img.shields.io/github/v/release/daytonaio/daytona)
[![Open Bounties](https://img.shields.io/endpoint?url=https%3A%2F%2Fconsole.algora.io%2Fapi%2Fshields%2Fdaytonaio%2Fbounties%3Fstatus%3Dopen)](https://console.algora.io/org/daytonaio/bounties?status=open)
[![Rewarded Bounties](https://img.shields.io/endpoint?url=https%3A%2F%2Fconsole.algora.io%2Fapi%2Fshields%2Fdaytonaio%2Fbounties%3Fstatus%3Dcompleted)](https://console.algora.io/org/daytonaio/bounties?status=completed)

</div>


<h1 align="center">The Open Source Development Environment Manager</h1>
<div align="center">
Set up a development environment on any infrastructure, with a single command.
</div>
</br>


<p align="center">
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=bug&projects=&template=bug_report.md&title=%F0%9F%90%9B+Bug+Report%3A+">Report Bug</a>
    ·
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=enhancement&projects=&template=feature_request.md&title=%F0%9F%9A%80+Feature%3A+">Request Feature</a>
    ·
  <a href="https://join.slack.com/t/daytonacommunity/shared_invite/zt-273yohksh-Q5YSB5V7tnQzX2RoTARr7Q">Join Our Slack</a>
    ·
    <a href="https://twitter.com/Daytonaio">Twitter</a>
  </p>

<div align="center"><img src="https://github.com/daytonaio/daytona/raw/main/assets/images/daytona_demo.gif" width="40%" ></div>

## Features
* __Single Command__: Activate a fully configured development environment with a single command.
* __Runs everywhere__: spin up your development environment on any machine — whether it's local, remote, cloud-based, physical server, or a VM & any architecture x86 or ARM.
* __Configuration File Support__: Initially support for [dev container](https://containers.dev/), ability to expand to DevFile, Nix & Flox (Contributions welcome here!).
* __Prebuilds System__: Drastically improve environment setup times (Contributions welcome here!).
* __IDE Support__ : Seamlessly supports [VS Code](https://github.com/microsoft/vscode) & [JetBrains](https://www.jetbrains.com/remote-development/gateway/) locally, ready to use without configuration. Includes a built-in Web IDE for added convenience.
* __Git Provider Integration__: GitHub, GitLab & Bitbucket can be connected, allowing easy repo branch or PR pull and commit back from the workspaces.
* __Multiple Project Workspace__: Support for multiple project repositories in the same workspace, making it easy to develop using a micro-service architecture.
* __Reverse Proxy Integration__: Enable collaboration and streamline feedback loops by leveraging reverse proxy functionality. Access preview ports and the Web IDE seamlessly, even behind firewalls.
* __Extensibility__: Enable extensibility with plugin or provider development. Moreover, in any dynamic language, not just Go(Contributions welcome here!).
* __Security__: Automatically creates a VPN connection between the client machine and the development environment, ensuring a fully secure connection.
* __All Ports__: The VPN connection enables access to all ports on the development environments, removing the need to setup port forwards over SSH connection.
* __Works on my Machine__: Never experience it again.
## Quick Start
To install Daytona all you need to do is execute this script:


```bash
# Install Daytona into ~/bin or ~/.local/bin
curl -sf -L https://download.daytona.io/daytona/get-server.sh | bash

# OR if you want to install Daytona to /usr/loca/bin or /opt/bin
# curl -sf -L https://download.daytona.io/daytona/get-server.sh | sudo bash

# Start Daytona server
daytona server
```
Create your first dev environment by opening a new terminal, and executing just this command:

```bash

daytona create
```

Start coding.


----

</br>

## Why Daytona?
Daytona is a radically simple open source development environment manager.

Setting up development environments has become increasingly challenging over time, especially when aiming to set up remotely, where the complexity increases by a order of magnitude. The process is so complex that we've compiled a [comprehensive guide](https://www.daytona.io/dotfiles/diy-guide-to-transform-any-machine-into-a-codespace) detailing all the necessary steps to set one up—spanning __5,000 words__, __7 steps__, and requiring anywhere from 15 to __45 minutes__.

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
curl https://download.daytona.io/daytona/get-server.sh | bash
```


### Initializing Daytona
To initialize Daytona, follow these steps:

__1. Start the Daytona Service:__
This initiates the Daytona service, which must always be running for Daytona to function. Use the command:
```bash
daytona server
```
__2. Add Your Git Provider of Choice:__
Daytona supports GitHub, GitLab, and Bitbucket. To add them to your profile, use the command:
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
The default setting for Daytona is VS Code locally. If you prefer, you can switch to VS Code - Browser or any IDE from the JetBrains portfolio (Contributions welcome here!) using the command:
```bash
daytona ide
```
Now that you have installed and initialized Daytona, you can proceed to setting up your development environments and start coding instantly.





### Creating Dev Environments
Creating development environments with Daytona is a straightforward process, accomplished with just one command:
```bash
daytona create
```
Upon executing this command, you will be prompted with two questions:
1. Choose the provider to decide where to create a dev environment.
2. Select or type the Git repository you wish to use to create a dev environment.

After making your selections, press enter, and Daytona will handle the rest. All that remains for you to do is to execute the following command to open your default IDE:
```bash
daytona code
```

This command opens your development environment in your preferred IDE, allowing you to start coding instantly.

## How to Extend Daytona

Daytona offers flexibility for extension through the creation of plugins and providers.


### Providers
Daytona is designed to be infrastructure-agnostic, capable of creating and managing development environments across various platforms. Providers are the components that encapsulate the logic for provisioning compute resources on a specific target platform. They allow for the configuration of different targets within a single provider, enabling, for instance, multiple AWS profiles within an AWS provider.

How does it work? When executing the `daytona create` command, Daytona communicates the environment details to the selected provider, which then provisions the necessary compute resources. Once provisioned, Daytona sets up the environment on these resources, allowing the user to interact with the environment seamlessly.

Providers are independent projects that adhere to the Daytona Provider interface. They can be developed in nearly any major programming language. More details coming soon.


### Plugins
Plugins enhance Daytona's core functionalities by adding new CLI commands, API methods, or services within the development environments. They offer configurable settings to tailor the plugin's behavior to the user's needs.

Similar to providers, plugins are independent projects that conform to the Daytona Plugin interface and can be developed in a wide range of programming languages. More details coming soon.




## Contributing


Daytona is Open Source under the [Apache License 2.0](LICENSE), and is the [copyright of its contributors](NOTICE). If you would like to contribute to the software, you must:

1. Read the [Contributors](CONTRIBUTORS.md) file.
2. Agree to the terms by having a commit in your pull request "signing" the file by adding your name and GitHub handle on a new line at the bottom of the file.
3. Make sure your commits Author metadata matches the name and handle you added to the file.

This ensures that users, distributors, and other contributors can rely on all the software related to Daytona being contributed under the terms of the [License](LICENSE). No contributions will be accepted without following this process.

Afterwards, navigate to the [contributing guide](CONTRIBUTING.md) to get started.

## License



This repository contains Daytona, covered under the [Apache License 2.0](LICENSE), except where noted (any Daytona logos or trademarks are not covered under the Apache License, and should be explicitly noted by a LICENSE file.)

Daytona is a product produced from this open source software, exclusively by Daytona Platforms, Inc. It is distributed under our commercial terms.

Others are allowed to make their own distribution of the software, but they cannot use any of the Daytona trademarks, cloud services, etc.

We explicitly grant permission for you to make a build that includes our trademarks while developing Daytona itself. You may not publish or share the build, and you may not use that build to run Daytona for any other purpose.

## Code of Conduct


This project has adapted the Code of Conduct from the [Contributor Covenant](https://www.contributor-covenant.org/). For more information see the [Code of Conduct](CODE_OF_CONDUCT.md) or contact [codeofconduct@daytona.io.](mailto:codeofconduct@daytona.io) with any additional questions or comments.

## Questions


For more information on how to use and develop Daytona, talk to us on
[Slack](https://join.slack.com/t/daytonacommunity/shared_invite/zt-273yohksh-Q5YSB5V7tnQzX2RoTARr7Q).

