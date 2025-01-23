<div align="center">

[![Documentation](https://img.shields.io/github/v/release/daytonaio/docs?label=Docs&color=23cc71)](https://www.daytona.io/docs)
[![License](https://img.shields.io/badge/License-Apache--2.0-blue)](#license)
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
  Set up a development environment on any infrastructure using a single command:
</h3>

<div style="width: 80%; margin: 0 auto;">

![Daytona Demo](/assets/images/daytona_demo.gif)

</div>

<p align="center">
    <a href="https://www.daytona.io/docs"> Documentation </a>·
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=bug&projects=&template=bug_report.md&title=%F0%9F%90%9B+Bug+Report%3A+"> Report Bug </a>·
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=enhancement&projects=&template=feature_request.md&title=%F0%9F%9A%80+Feature%3A+"> Request Feature </a>·
    <a href="https://go.daytona.io/slack"> Join Our Slack </a>·
    <a href="https://x.com/daytonaio"> Connect On X </a>
</p>

<div align="center">

# Open Source Development Environment Manager
</div>

For detailed/manual setup steps click [here](https://www.daytona.io/docs/installation/installation/#installation)

### Mac / Linux

```bash
curl -sfL get.daytona.io | sudo bash && daytona server -y && daytona
```

### Windows

```pwsh
powershell -Command "irm https://get.daytona.io/windows | iex; daytona serve"
```

### Create your first dev environment by opening a new terminal, and running:

```bash
daytona create
```

**Start coding.**

---

## Features

- **Quick Setup**: Activate a fully configured development environment with a single command - `daytona create`.
- **Runs everywhere**: Spin up your development environment on any machine; local, remote, cloud-based, physical server or a VM & on any architecture; x86 or ARM.
- **Various Providers Support**: Choose popular providers like AWS, GCP, Azure, DigitalOcean & [more](https://github.com/orgs/daytonaio/repositories?q=daytona-provider) or use Docker on bare metal.
- **IDE Support** : Seamlessly supports [VS Code](https://github.com/microsoft/vscode), [JetBrains](https://www.jetbrains.com/remote-development/gateway/) products and more, ready to use without configuration. Also includes a built-in Web IDE for convenience.
- **Git Provider Integration**: GitHub, GitLab, Bitbucket and [other](https://www.daytona.io/docs/configuration/git-providers/#add-a-git-provider) Git providers can be connected allowing you to start working on a specific branch or PR and to push changes immediately. 
- **Configuration File Support**: Support for [dev container](https://containers.dev/) and an upcoming expansion to DevFile, Nix & Flox.
- **Prebuilds System**: Drastically improve environment build times by prebuilding them based on Git Providers' hook events.
- **Reverse Proxy Integration**: Enable collaboration and streamline feedback loops by leveraging our reverse proxy. Access preview ports and the Web IDE seamlessly, even behind firewalls. 
- **Security**: Automatically creates a VPN connection between the client machine and the development environment, ensuring a fully secure connection.
- **Works on my Machine**: Never experience it again.

*For a complete feature set including Authentication, Authorization, Observability, Resource Management and IDP, check out our [enterprise offering](https://daytona.zapier.app/).

---

## Getting Started

### Requirements

Before starting the installation script, if developing locally, ensure [Docker](https://www.docker.com/products/docker-desktop/) is installed and running.

### Initializing Daytona

To initialize Daytona, follow these steps:

**1. Start the Daytona Server:**
Use this command to initiate the Daytona Server in daemon mode or use `daytona serve` to run it in the foreground:

```bash
daytona server
```

**2. Register Your Git Provider of Choice:**
Daytona supports GitHub, GitLab, Bitbucket and [more](https://www.daytona.io/docs/configuration/git-providers/#add-a-git-provider) Git Providers. Use this command to set them up:

```bash
daytona git-provider create
```

**3. Create Your First Target:** (optional)
By default, Daytona uses the Docker provider to spin up environments on your local machine. For remote development environment setups, use the following command:

```bash"
daytona target create
```

**4. Choose Your IDE:**
The default IDE for Daytona is the local VS Code installation. To switch to the Web IDE or any other IDE, use:

```bash
daytona ide
```

Now that you have installed and initialized Daytona, you may proceed to setting up your development environments and starting to code instantly.

**4. Create Your First Daytona Development Environment:**

Creating development environments with Daytona is a straightforward process accomplished with just one command which prompts you for two things:

1. Choose the target/provider to decide where to create the dev environment.
2. Select or type in the Git repository you wish to start off with.

After making your selections, press enter, and Daytona will handle the rest.

```bash
daytona create
```

*You can add the `--no-ide` flag to skip opening the IDE and then use `daytona code` once you're ready to start coding. More info [here](https://www.daytona.io/docs/about/getting-started/).

**5. Manage the Daytona Server daemon:**

```bash
daytona server [start|stop|restart]
```

---

## Extend Daytona Through Providers

Daytona is designed to be infrastructure-agnostic, capable of creating and managing development environments across various platforms. Providers are the components that encapsulate the logic for provisioning compute resources on a specific platform. They allow for the configuration of different target configurations thus enabling, for instance, multiple AWS profiles within an AWS provider.

How does it work? When executing the `daytona create` command, Daytona communicates the environment details to the selected provider, which then provisions the necessary compute resources. Once provisioned, Daytona sets up the environment on these resources, allowing the user to interact with the environment seamlessly.

Providers are independent projects that adhere to the Daytona Provider interface. View all currently supported providers [here](https://github.com/orgs/daytonaio/repositories?q=daytona-provider).

## Contributing

Daytona is Open Source under the [Apache License 2.0](LICENSE), and is the [copyright of its contributors](NOTICE). If you would like to contribute to the software, read the Developer Certificate of Origin Version 1.1 (https://developercertificate.org/). Afterwards, navigate to the [contributing guide](CONTRIBUTING.md) to get started.