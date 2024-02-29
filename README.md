<div align="center"><img src="https://github.com/ivan-burazin/daytona/blob/main/public/images/Daytona-logotype-black.svg" >
<br><br>

[![License](https://img.shields.io/badge/License-Apache--2.0-blue)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/daytonaio/daytona)](https://goreportcard.com/report/github.com/daytonaio/daytona)
[![issues - daytona](https://img.shields.io/github/issues/daytonaio/daytona)](https://github.com/daytonaio/daytona/issues)
![GitHub Release](https://img.shields.io/github/v/release/daytonaio/daytona)
[![See latest](https://img.shields.io/static/v1?label=Docs&message=see%20latest&color=blue)](https://daytona.io/docs)



</div>


<h1 align="center">The Open Source Development Environment Manager</h1>
<div align="center">
Empowering you to effortlessly create a functional development environment on any infrastructure with a single command.
</div>
</br>


<p align="center">
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=type%3A+bug&template=bug_report.yml&title=%F0%9F%90%9B+Bug+Report%3A+">Report Bug</a>
    ·
    <a href="https://github.com/daytonaio/daytona/issues/new?assignees=&labels=feature&template=feature_request.yml&title=%F0%9F%9A%80+Feature%3A+">Request Feature</a>
    ·
  <a href="https://join.slack.com/t/daytonacommunity/shared_invite/zt-273yohksh-Q5YSB5V7tnQzX2RoTARr7Q">Join Our Slack</a>
    ·
    <a href="https://twitter.com/Daytona">Twitter</a>
  </p>

## Backstory
We spent most of our careers building cloud development environments. In 2009, we launched what was likely the first commercial Cloud IDE project. At that time, technology was lacking, forcing us to develop everything from scratch—the IDE, the environment orchestrator, and almost everything else. A lot of people were interested, and over 2.5 million developers signed up! But we were too early, and we asked too much from our users to change how they worked.

Now, 15 years since its inception, we have noticed quite a few things. First, the technology we wished for back then exists now. Second, approximately 50% of developers work in remote dev environments, and third, and most importantly, setting up development environments has become more complex than ever, both locally and to a greater magnitude for remote.

So, we took everything we learned and decided to solve these issues once and for all as a fully open-source project. Our goal was to create a single binary that allows you to set up a development environment anywhere you wish, completely free, and finally fulfill the promise that many have attempted to make.


## Why Daytona?


Daytona is a radically simple open source development environment standard. It automates the entire process; provisioning the instance, reading the configuration and executing it, setting up prebuilds, attaching your local or a Web IDE, and adding a fully qualified domain name to the dev environment to share and collaborate.

Daytona builds upon a decade and a half of experience of the team responsible for pioneering the Cloud IDE movement. This expertise encompasses the development of IDEs and orchestrators from scratch, facilitating the provision of instant cloud development environments. By integrating modern technologies and advancements with this historical knowledge,  Daytona is committed to eradicating the "works on my machine" syndrome once and for all.

More information on the [Daytona](https://daytona.io/) website.

----

## Features
* __Single Command__: Activate a fully configured development environment with a single command.
* __Runs everywhere__: Spin up your development environment on any machine—whether it's local, remote, a physical server, VM, or cloud-based, & any architecture x86 or Arm.
* __Confiuration File Support__: Initially support for [dev container](https://containers.dev/), ability to expand to DevFile, Nix & Backstage (Contributions welcome here!).
* __Prebuilds System__: Has a prebuilds system, built in, to drasticly improve environment setup times(Contributions welcome here!).
* __IDE Support__ : Seamlessly supports [VS Code](https://github.com/microsoft/vscode) & [JetBrains](https://www.jetbrains.com/remote-development/gateway/) locally, ready to use without configuration. Includes a built-in Web IDE for added convenience.
* __Git Provider Integration__: Start with GitHub support, allowing easy repo or PR selection. Future plans to expand to GitLab & Bitbucket (Contributions welcome here!).
* __Multiple Project Workspace__: Support for multiple project repositories in a dev environment, making it easy to develop using micro-service architecture.
* __Reverse Proxy Integration__: Enable collaboration and streamline feedback loops by leveraging reverse proxy functionality. Access preview environments and the Web IDE seamlessly, even behind firewalls.
* __Extensibility__: Enable extensibility with plugin or provider development. Moreover, in any dynamic language, not just Go(Contributions welcome here!).
* __Security__: Compatibility with all existing technologies development environment technologies
* __Works on my Machine__: Never experince it again.

## Getting Started


### Installing Daytona
Daytona allows you to manage your Development Environments using the Daytona CLI. To install it, please execute the following command:

```bash
curl https://download.daytona.io/daytona/get-server.sh | bash
```

You can also download and compile Daytona from this repository. To do so please check out our docs.

Note the packaged Daytona includes a set of built-in extensions located in the extensions folder, which are added there for your convince.

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
Follow the steps provided. Here's a link to the [documentation](https://daytona.io/docs) for more details.

__3. Add Your Provider:__  
This step is for choosing where to deploy Development Environments. By default, Daytona includes a Docker provider to spin up environments on your local machine. For remote development environments, use the command:
```bash
daytona providers add
```
This command allows adding connections to an SSH machine or one of the cloud providers (Contributions welcome here!).

__4. Choose Your Default IDE:__  
The default setting for Daytona is VS Code locally. If you prefer, you can switch to VS Code - Browser or any IDE from the JetBrains portfolio (Contributions welcome here!) using the command:
```bash
daytona ide
```
Now that you have installed and initialized Daytona, you can proceed to setting up your development environments and start coding instantly.  


## Using Daytona


### Creating Dev Environments 
Creating development environments with Daytona is a straightforward process, accomplished with just one command:
```bash
daytona create
```
Upon executing this command, you will be prompted with two questions:
1. Where would you like to create this environment? Here, you choose by selecting the provider.
2. Which repository will you use to create this environment? Here , you can selcet your Git repo or type it in.

After making your selections, press enter, and Daytona will handle the rest. All that remains for you to do is to execute the following command to open your default IDE:
```bash
daytona code
```

This command opens your development environment in your preferred IDE, allowing you to start coding instantly.

### Manipulating Dev Environments  
To manage your development environments, Daytona provides several commands that facilitate various operations:  

__- Listing Dev Environments:__ To view a list of all your dev environments, use:
```bash
daytona list
```

__- Deleting a Dev Environment:__ To remove a specific development environment, execute:
```bash
daytona delete
```
This command deletes the specified development environment.

__- Displaying Dev Environment Information:__ or details about a specific workspace, including its status and configuration, use:
```bash
daytona info
```

__- Starting and Stopping Dev Environments:__   
- To start a workspace, making it active and accessible, use:
```bash
daytona start
```
- To stop a workspace, thereby deactivating it, use:
```bash
daytona start
```

__- Managing Port Forwarding:__ If you need to manage the ports forwarded to your project, facilitating access to services running in your development environment, use:
```bash
daytona ports
```

__- SSH Access:__ For direct SSH access to a development environment using the terminal, execute:
```bash
daytona ssh
```
This allows for a secure command-line interface with your development environment.

### Other Commands
In addition to the creation and management of development environments, Daytona provides several commands for customization and accessing information:

__- Managing Profiles:__ Daytona allows for the management of multiple profiles, enabling users to switch between personal use and connecting to a company's installation of the Daytona platform. To manage profiles, use the following command:
```bash
daytona profile
```

__- View Version:__ To find out the version of Daytona you are using, the following command can be used to print the version number:
```bash
daytona version
```

For more detailed information about each command, please refer to Daytona's documentation.







  

## Architecture 
TODO: desc

## How to Extend Daytona

Daytona offers flexibility for extension through the creation of plugins and providers.


### Providers 
Daytona is designed to be infrastructure-agnostic, capable of creating and managing development environments across various platforms. Providers are the components that encapsulate the logic for provisioning compute resources on a specific target platform. They allow for the configuration of different profiles within a single provider, enabling, for instance, multiple AWS profiles within an AWS provider.

How does it work? When executing the `daytona create` command, Daytona communicates the environment details to the selected provider, which then provisions the necessary compute resources. Once provisioned, Daytona sets up the environment on these resources, allowing the user to interact with the environment seamlessly.

PProviders are independent projects that adhere to the Daytona Provider interface. They can be developed in nearly any major programming language. For more details, see [Providers](providers/readme.md)


### Plugins
Plugins enhance Daytona's core functionalities by adding new CLI commands, API methods, or services within the development environments. They offer configurable settings to tailor the plugin's behavior to the user's needs.

Similar to providers, plugins are independent projects that conform to the Daytona Plugin interface and can be developed in a wide range of programming languages. For more information, visit  [Plugins](plugins/readme.md)




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
[Slack](https://join.slack.com/t/daytonacommunity/shared_invite/zt-273yohksh-Q5YSB5V7tnQzX2RoTARr7Q) and check out our [documentation](https://www.daytona.io/docs/installation/server/).


