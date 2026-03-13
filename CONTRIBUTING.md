<a id="top"></a>

# Contributing to Scribe-Server

Thank you for your interest in contributing!

Please take a moment to review this document in order to make the contribution process easy and effective for everyone involved.

Following these guidelines helps to communicate that you respect the time of the developers managing and developing this open-source project. In return, and in accordance with this project's [code of conduct](https://github.com/scribe-org/Scribe-Server/blob/main/.github/CODE_OF_CONDUCT.md), other contributors will reciprocate that respect in addressing your issue or assessing changes and features.

If you have questions or would like to communicate with the team, please [join us in our public Matrix chat rooms](https://matrix.to/#/#scribe_community:matrix.org). We'd be happy to hear from you!

## Contents

- [First steps as a contributor](#first-steps-as-a-contributor)
- [Mentorship and Growth](#mentorship-and-growth)
- [Learning the tech stack](#learning-the-tech)
- [Development environment](#development-environment)
- [Issues and projects](#issues-and-projects)
- [Bug reports](#bug-reports)
- [Feature requests](#feature-requests)
- [Pull requests](#pull-requests)
- [Data edits](#data-edits)
- [Documentation](#documentation)
- [Deployment testing](#deployment-testing)

## First steps as a contributor

Thank you for your interest in contributing to Scribe-Server! We look forward to welcoming you to the community and working with you to build an tools for language learners to communicate effectively :) The following are some suggested steps for people interested in joining our community:

- Please join the [public Matrix chat](https://matrix.to/#/#scribe_community:matrix.org) to connect with the community
  - [Matrix](https://matrix.org/) is a network for secure, decentralized communication
  - We'd suggest that you use the [Element](https://element.io/) client and [Element X](https://element.io/app) for a mobile app
  - The [General](https://matrix.to/#/!yQJjLmluvlkWttNhKo:matrix.org?via=matrix.org) and [Data](https://matrix.to/#/#ScribeData:matrix.org) channels would be great places to start!
  - Feel free to introduce yourself and tell us what your interests are if you're comfortable :)
- Read through this contributing guide for all the information you need to contribute
- Look into issues marked [`good first issue`](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22) and the [Projects board](https://github.com/orgs/scribe-org/projects/1) to get a better understanding of what you can work on
- Check out our [public designs on Figma](https://www.figma.com/file/c8945w2iyoPYVhsqW7vRn6/scribe_public_designs?type=design&node-id=405-464&mode=design&t=E3ccS9Z8MDVSizQ4-0) to understand Scribes's goals and direction
- Consider joining our [bi-weekly developer sync](https://etherpad.wikimedia.org/p/scribe-dev-sync)!

> [!NOTE]
> Those new to Go or wanting to work on their Go skills are more than welcome to contribute! The team would be happy to help you on your development journey :)

<sub><a href="#top">Back to top.</a></sub>

## Mentorship and Growth

Onboarding and mentoring new members is vital to a healthy open-source community.

We need contributors who are onboarded to gain new skills and take on greater roles by triaging issues, reviewing contributions, and maintaining the project. We also need them to help new contributors to grow as well. Please let us know if you have goals to develop as an open-source contributor and we'll work with you to achieve them.

We also have expectations about the behavior of those who want to grow with us. Mentorship is earned, not given.

To be blunt, those who are mainly sending AI generated contributions are not demonstrating an interest in growing their skills and are not helping to develop the project. This is not to say that all uses of AI for contributions are bad, but **AI should be a tool, not the contributor itself**.

Continued constructive contributions, new open issues, and clear communication helps the project. We would be happy to help community members who can make these contributions to expand their skills and take on further responsibilities.

If you like the sound of this, then we look forward to working with you!

<sub><a href="#top">Back to top.</a></sub>

## Learning the tech stack

Scribe is very open to contributions from people in the early stages of their coding journey! The following is a select list of documentation pages to help you understand the technologies we use.

<details><summary>Docs for those new to programming</summary>
<p>

- [Mozilla Developer Network Learning Area](https://developer.mozilla.org/en-US/docs/Learn)
  - Doing MDN sections for HTML, CSS and JavaScript is the best ways to get into web development!
- [Open Source Guides](https://opensource.guide/)
  - Guides from GitHub about open-source software including how to start and much more!

</p>
</details>

<details><summary>Go learning docs</summary>
<p>

- [Go getting started guide](https://go.dev/learn/)
- [Go documentation](https://go.dev/doc/)

</p>
</details>

<sub><a href="#top">Back to top.</a></sub>

## Development environment

Scribe-Server is developed using the [Go](https://go.dev/) programming language. Those new to Go or wanting to develop their skills are more than welcome to contribute! The first step on your Go journey would be to read through the [Go documentation](https://go.dev/doc), with the [Effective Go](https://go.dev/doc/effective_go) page in particular having great insights into the language's good practices and standards. The general steps to setting up a development environment are:

1. Download and install [Go](https://go.dev/doc/install)

2. [Fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo) the [Scribe-Server repo](https://github.com/scribe-org/Scribe-Server), clone your fork, and configure the remotes:

> [!NOTE]
>
> <details><summary>Consider using SSH</summary>
>
> <p>
>
> Alternatively to using HTTPS as in the instructions below, consider SSH to interact with GitHub from the terminal. SSH allows you to connect without a user-pass authentication flow.
>
> To run git commands with SSH, remember then to substitute the HTTPS URL, `https://github.com/...`, with the SSH one, `git@github.com:...`.
>
> - e.g. Cloning now becomes `git clone git@github.com:<your-username>/Scribe-Server.git`
>
> GitHub also has their documentation on how to [Generate a new SSH key](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent) 🔑
>
> </p>
> </details>

```bash
# Clone your fork of the repo into the current directory.
git clone https://github.com/<your-username>/Scribe-Server.git
# Navigate to the newly cloned directory.
cd Scribe-Server
# Assign the original repo to a remote called "upstream".
git remote add upstream https://github.com/scribe-org/Scribe-Server.git
```

- Now, if you run `git remote -v` you should see two remote repositories named:
  - `origin` (forked repository)
  - `upstream` (Scribe-Server repository)

3. Navigate to the root directory of the project

4. Create a `config.yaml` file with the configuration needed for the project
   - Reference the [`config-example.yaml`](./config-example.yaml) to get started

## Important Note on PATH Configuration for Go Tools

> After installing `Go`, it's highly recommended to add your `Go` binary directory (`$(go env GOPATH)/bin`) to your system's `PATH`. This ensures that tools installed via `go install` (like `swag`, `oapi-codegen`, etc.) are directly accessible from your terminal.

You can typically do this by adding the following line to your shell configuration file (e.g., `~/.bashrc`, `~/.zshrc`, or `~/.profile`):

```bash
export PATH=$(go env GOPATH)/bin:$PATH
```

After adding this line, remember to apply the changes by sourcing the file (e.g., `source ~/.bashrc`) or by opening a new terminal session.

5. Install [MariaDB](https://mariadb.com/) locally via its [installation guide](https://mariadb.com/docs/server/server-management/install-and-upgrade-mariadb/installing-mariadb/binary-packages).
   - Create a database using the `database.name` value from your `config.yaml` with the following commands (using Homebrew, for example):

   ```bash
   brew services start mariadb
   mariadb -u root  # you may need to sudo this command

   # To stop the server:
   brew services stop mariadb
   ```

   - You can now run the commands found in [CREATE_SCRIBE_SERVER_DB.md](./CREATE_SCRIBE_SERVER_DB.md) to make the needed MariaDB database.

6. Start a local Scribe-Server:

   ```bash
   # Run the following target from the 'Makefile'.
   # Migrate SQLite files from Scribe-Data to MariaDB for use in Scribe-Server:
   make build
   make migrate
   # Start Scribe-Server on your local host:
   make run
   ```

   - NOTE: This `make` target simply runs `go run .` on the project
   - Scribe-Server should now be running locally!

7. To generate the documentation for Scribe-Server, please run the following:

   ```bash
   make docs
   ```

   Once the server is running (via `make run` or `make dev`), you can access the API documentation at:
   - Swagger UI: http://localhost:8080/swagger/index.html</br>
   - Alternative docs: http://localhost:8080/docs/index.html

> [!NOTE]
> Feel free to contact the team in the [Data room on Matrix](https://matrix.to/#/#ScribeData:matrix.org) if you're having problems getting your environment setup!

<sub><a href="#top">Back to top.</a></sub>

## Issues and projects

The [issue tracker for Scribe-Server](https://github.com/scribe-org/Scribe-Server/issues) is the preferred channel for [bug reports](#bug-reports), [features requests](#feature-requests) and [submitting pull requests](#pull-requests). Scribe also organizes related issues into [projects](https://github.com/scribe-org/Scribe-Server/projects).

> [!NOTE]\
> Just because an issue is assigned on GitHub doesn't mean the team isn't open to your contribution! Feel free to write [in the issues](https://github.com/scribe-org/Scribe-Server/issues) and we can potentially reassign it to you.

Be sure to check the [`-next release-`](https://github.com/scribe-org/Scribe-Server/labels/-next%20release-) and [`-priority-`](https://github.com/scribe-org/Scribe-Server/labels/-priority-) labels in the [issues](https://github.com/scribe-org/Scribe-Server/issues) for those that are most important, as well as those marked [`good first issue`](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) that are tailored for first-time contributors.

<sub><a href="#top">Back to top.</a></sub>

## Bug reports

A bug is a _demonstrable problem_ that is caused by the code in the repository. Good bug reports are extremely helpful - thank you!

Guidelines for bug reports:

1. **Use the GitHub issue search** to check if the issue has already been reported.

2. **Check if the issue has been fixed** by trying to reproduce it using the latest `main` or development branch in the repository.

3. **Isolate the problem** to make sure that the code in the repository is _definitely_ responsible for the issue.

**Great Bug Reports** tend to have:

- A quick summary
- Steps to reproduce
- What you expected would happen
- What actually happens
- Notes (why this might be happening, things tried that didn't work, etc)

To make the above steps easier, the Scribe team asks that contributors report bugs using the [bug report](https://github.com/scribe-org/Scribe-Server/issues/new?assignees=&labels=feature&template=bug_report.yml) template, with these issues further being marked with the [`Bug`](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aissue%20state%3Aopen%20type%3ABug) type.

Again, thank you for your time in reporting issues!

<sub><a href="#top">Back to top.</a></sub>

## Feature requests

Feature requests are more than welcome! Please take a moment to find out whether your idea fits with the scope and aims of the project. When making a suggestion, provide as much detail and context as possible, and further make clear the degree to which you would like to contribute in its development. Feature requests are marked with the [`Feature`](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aissue%20state%3Aopen%20type%3AFeature) type, and can be made using the [feature request](https://github.com/scribe-org/Scribe-Server/issues/new?assignees=&labels=feature&template=feature_request.yml) template.

<sub><a href="#top">Back to top.</a></sub>

## Pull requests

Good pull requests - patches, improvements and new features - are the foundation of our community making Scribe-Server. They should remain focused in scope and avoid containing unrelated commits. Note that all contributions to this project will be made under [the specified license](https://github.com/scribe-org/Scribe-Server/blob/main/LICENSE.txt) and should follow the code style standards ([contact us](https://matrix.to/#/#scribe_community:matrix.org) if unsure).

**Please ask first** before embarking on any significant pull request (implementing features, refactoring code, etc), otherwise you risk spending a lot of time working on something that the developers might not want to merge into the project. With that being said, major additions are very appreciated!

When making a contribution, adhering to the [GitHub flow](https://guides.github.com/introduction/flow/index.html) process is the best way to get your work merged:

1. If you cloned a while ago, get the latest changes from upstream:

   ```bash
   git checkout <dev-branch>
   git pull upstream <dev-branch>
   ```

2. Create a new topic branch (off the main project development branch) to contain your feature, change, or fix:

   ```bash
   git checkout -b <topic-branch-name>
   ```

3. Commit your changes in logical chunks, and please try to adhere to [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

> [!NOTE]
> The following are tools and methods to help you write good commit messages ✨
>
> - [commitlint](https://commitlint.io/) helps write [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
> - Git's [interactive rebase](https://docs.github.com/en/github/getting-started-with-github/about-git-rebase) cleans up commits

4. Locally merge (or rebase) the upstream development branch into your topic branch:

   ```bash
   git pull --rebase upstream <dev-branch>
   ```

5. Push your topic branch up to your fork:

   ```bash
   git push origin <topic-branch-name>
   ```

6. [Open a Pull Request](https://help.github.com/articles/using-pull-requests/) with a clear title and description.

Thank you in advance for your contributions!

<sub><a href="#top">Back to top.</a></sub>

## Data edits

> [!NOTE]\
> Please see the [Wikidata and Scribe Guide](https://github.com/scribe-org/Organization/blob/main/WIKIDATAGUIDE.md) for an overview of [Wikidata](https://www.wikidata.org/) and how Scribe uses it.

Scribe does not accept direct edits to the grammar JSON files as they are sourced from [Wikidata](https://www.wikidata.org/). Edits can be discussed and the [Scribe-Data](https://github.com/scribe-org/Scribe-Data) queries will be changed and ran before an update. If there is a problem with one of the files, then the fix should be made on [Wikidata](https://www.wikidata.org/) and not on Scribe. Feel free to let us know that edits have been made by [opening a data issue](https://github.com/scribe-org/Scribe-Server/issues/new?assignees=&labels=data&template=data_wikidata.yml) or contacting us in the [issues for Scribe-Data](https://github.com/scribe-org/Scribe-Data/issues) and we'll be happy to integrate them!

<sub><a href="#top">Back to top.</a></sub>

## Documentation

Documentation is an invaluable way to contribute to coding projects as it allows others to more easily understand the project structure and contribute. Issues related to documentation are marked with the [`documentation`](https://github.com/scribe-org/Scribe-Server/labels/documentation) label.

<sub><a href="#top">Back to top.</a></sub>

## Deployment testing

This guide explains how to test the GitHub Actions workflow that updates data and deploys to Toolforge via GitHub Actions.

### 1. Generate a New SSH Key (One-Time)

Create a new key pair (without passphrase) specifically for GitHub Actions and Toolforge use:

```bash
ssh-keygen -t ed25519 -C "github-actions-scribe-data" -f ~/.ssh/scribe_toolforge_deploy
```

- Press **Enter** when asked for a passphrase (leave it empty).

### 2. Add Your Public Key to Toolforge

Copy the public key content:

```bash
cat ~/.ssh/scribe_toolforge_deploy.pub
```

Then log into Toolforge and add it:

```bash
ssh your-username@login.toolforge.org
become &lt;your-tool-name&gt;
echo "paste-your-public-key-here" &gt;&gt; ~/.ssh/authorized_keys
```

### 3. Add GitHub Secrets

Copy your private key content:

```bash
cat ~/.ssh/scribe_toolforge_deploy
```

In your GitHub repository → **Settings → Secrets and variables → Actions**, add the following:

| Secret Name       |                                               Value |
| :---------------- | --------------------------------------------------: |
| TOOLFORGE_SSH_KEY |            Paste the full output of the private key |
| TOOLFORGE_USER    | Your Toolforge username (e.g. <code>johndoe</code>) |

#### REQUIRED: Also copy public key to Toolforge (for manual SSH if needed), If not the login process will not work!

**🔑 Visit your Toolforge account and add your public SSH key at:** [https://toolsadmin.wikimedia.org/profile/settings/ssh-keys/](https://toolsadmin.wikimedia.org/profile/settings/ssh-keys/)

### 4. Run the Workflow

- Go to **GitHub → Actions → "Update Scribe Data and Deploy to Toolforge"**
- Click **"Run workflow"** → Choose branch if needed → **Run workflow**
- Check logs for status and output

<sub><a href="#top">Back to top.</a></sub>
