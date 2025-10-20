<div align="center">
  <a href="https://github.com/scribe-org/Scribe-Server"><img src="https://raw.githubusercontent.com/scribe-org/Scribe-Server/main/.github/resources/images/ScribeServerGitHubOrgBanner.png" width=1024 alt="Scribe-Server Logo"></a>
</div>

[![platforms](https://img.shields.io/static/v1?message=Toolforge&logo=wikimedia-foundation&color=990000&logoColor=white&label=%20)](https://github.com/scribe-org/Scribe-Server)
[![issues](https://img.shields.io/github/issues/scribe-org/Scribe-Server?label=%20&logo=github)](https://github.com/scribe-org/Scribe-Server/issues)
[![language](https://img.shields.io/badge/Go%201.20-00ADD8.svg?logo=go&logoColor=ffffff)](https://github.com/scribe-org/Scribe-Server/blob/main/CONTRIBUTING.md)
[![license](https://img.shields.io/github/license/scribe-org/Scribe-Server.svg?label=%20)](https://github.com/scribe-org/Scribe-Server/blob/main/LICENSE.txt)
[![coc](https://img.shields.io/badge/Contributor%20Covenant-ff69b4.svg)](https://github.com/scribe-org/Scribe-Server/blob/main/.github/CODE_OF_CONDUCT.md)
[![mastodon](https://img.shields.io/badge/Mastodon-6364FF.svg?logo=mastodon&logoColor=ffffff)](https://wikis.world/@scribe)
[![matrix](https://img.shields.io/badge/Matrix-000000.svg?logo=matrix&logoColor=ffffff)](https://matrix.to/#/#scribe_community:matrix.org)

### Backend service for Scribe data downloads

**Scribe-Server** is a backend service that provides the API by which data is available for download within Scribe apps. The goal is to create a [Scribe-Data](https://github.com/scribe-org/Scribe-Data) based regularly updating dataset that can signal new data availability as well as allow for language pack downloads. Scribe-Server can be accessed via [scribe-server.toolforge.org](https://scribe-server.toolforge.org/), with the SQLite data packs being available for download via [scribe-server.toolforge.org/packs/sqlite](https://scribe-server.toolforge.org/packs/sqlite/).

> [!NOTE]\
> The [contributing](#contributing) section has information for those interested, with the articles and presentations in [featured by](#featured-by) also being good resources for learning more about Scribe.

Scribe apps are available on [iOS](https://github.com/scribe-org/Scribe-iOS), [Android](https://github.com/scribe-org/Scribe-Android) (planned) and [Desktop](https://github.com/scribe-org/Scribe-Desktop) (planned). For the data formatting processes see [Scribe-Data](https://github.com/scribe-org/Scribe-Data).

Check out Scribe's [architecture diagrams](https://github.com/scribe-org/Organization/blob/main/ARCHITECTURE.md) for an overview of the organization including our applications, services and processes. It depicts the projects that [Scribe](https://github.com/scribe-org) is developing as well as the relationships between them and the external systems with which they interact.

<a id="contents"></a>

# **Contents**

- [Contributing](#contributing)
- [Environment Setup](#environment-setup)
- [Supported Languages](#supported-languages)
- [Featured By](#featured-by)

<a id="contributing"></a>

# Contributing [`⇧`](#contents)

<a href="https://matrix.to/#/#scribe_community:matrix.org">
  <img src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/MatrixLogoGrey.png" width="175" alt="Public Matrix Chat" align="right">
</a>

Scribe uses [Matrix](https://matrix.org/) for communications. You're more than welcome to [join us in our public chat rooms](https://matrix.to/#/#scribe_community:matrix.org) to share ideas, ask questions or just say hi to the team :) We'd suggest that you use the [Element](https://element.io/) client and [Element X](https://element.io/app) for a mobile app.

Please see the [contribution guidelines](https://github.com/scribe-org/Scribe-Server/blob/main/CONTRIBUTING.md) if you are interested in contributing to Scribe-Server. Work that is in progress or could be implemented is tracked in the [issues](https://github.com/scribe-org/Scribe-Server/issues) and [projects](https://github.com/scribe-org/Scribe-Server/projects).

> [!NOTE]\
> Just because an issue is assigned on GitHub doesn't mean the team isn't open to your contribution! Feel free to write [in the issues](https://github.com/scribe-org/Scribe-Server/issues) and we can potentially reassign it to you.

Those interested can further check the [`-next release-`](https://github.com/scribe-org/Scribe-Server/labels/-next%20release-) and [`-priority-`](https://github.com/scribe-org/Scribe-Server/labels/-priority-) labels in the [issues](https://github.com/scribe-org/Scribe-Server/issues) for those that are most important, as well as those marked [`good first issue`](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) that are tailored for first-time contributors. For those new to coding or our tech stack, we've collected [links to helpful documentation pages](https://github.com/scribe-org/Scribe-Server/blob/main/CONTRIBUTING.md#learning-the-tech) in the [contribution guidelines](https://github.com/scribe-org/Scribe-Server/blob/main/CONTRIBUTING.md).

After your first few pull requests organization members would be happy to discuss granting you further rights as a contributor, with a maintainer role then being possible after continued interest in the project. Scribe seeks to be an inclusive and supportive organization. We'd love to have you on the team!

### Ways to Help [`⇧`](#contents)

- [Reporting bugs](https://github.com/scribe-org/Scribe-Server/issues/new?assignees=&labels=bug&template=bug_report.yml) as they're found 🐞
- Working on [new features](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aissue+is%3Aopen+label%3Afeature) ✨
- [Localization](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aissue+is%3Aopen+label%3Alocalization) for the app and Google Play 🌐
- [Documentation](https://github.com/scribe-org/Scribe-Server/issues?q=is%3Aissue+is%3Aopen+label%3Adocumentation) for onboarding and project cohesion 📝
- Adding language data to [Scribe-Data](https://github.com/scribe-org/Scribe-Data/issues) via [Wikidata](https://www.wikidata.org/)! 🗃️

### Road Map [`⇧`](#contents)

The Scribe road map can be followed in the organization's [project board](https://github.com/orgs/scribe-org/projects/1) where we list the most important issues along with their priority, status and an indication of which sub projects they're included in (if applicable).

> [!NOTE]\
> Consider joining our [bi-weekly developer syncs](https://etherpad.wikimedia.org/p/scribe-dev-sync)!

### Data Edits [`⇧`](#contents)

> [!NOTE]\
> Please see the [Wikidata and Scribe Guide](https://github.com/scribe-org/Organization/blob/main/WIKIDATAGUIDE.md) for an overview of [Wikidata](https://www.wikidata.org/) and how Scribe uses it.

Scribe does not accept direct edits to the grammar JSON files as they are sourced from [Wikidata](https://www.wikidata.org/). Edits can be discussed and the [Scribe-Data](https://github.com/scribe-org/Scribe-Data) queries will be changed and ran before an update. If there is a problem with one of the files, then the fix should be made on [Wikidata](https://www.wikidata.org/) and not on Scribe. Feel free to let us know that edits have been made by [opening a data issue](https://github.com/scribe-org/Scribe-Server/issues/new?assignees=&labels=data&template=data_wikidata.yml) or contacting us in the [issues for Scribe-Data](https://github.com/scribe-org/Scribe-Data/issues) and we'll be happy to integrate them!

<a id="environment-setup"></a>

# Environment Setup [`⇧`](#contents)

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

<a id="supported-languages"></a>

# Supported Languages [`⇧`](#contents)

Scribe's goal is functional, feature-rich keyboards for all languages. You can check the currently available languages and data for Scribe applications on our website at [scri.be/docs/server/available-data](https://scri.be/docs/server/available-data).

See [scribe_data/wikidata/language_data_extraction](https://github.com/scribe-org/Scribe-Data/tree/main/src/scribe_data/wikidata/language_data_extraction) for queries in the [Scribe-Data](https://github.com/scribe-org/Scribe-Data) project for currently supported languages and those that have substantial data on [Wikidata](https://www.wikidata.org/). Also see the [`new keyboard`](https://github.com/scribe-org/Scribe-iOS/issues?q=is%3Aissue+is%3Aopen+label%3A%22new+keyboard%22) label in the [Issues](https://github.com/scribe-org/Scribe-iOS/issues) for keyboards that are currently in progress or being discussed, and [suggest a new keyboard](https://github.com/scribe-org/Scribe-iOS/issues/new?assignees=&labels=new+keyboard&template=new_keyboard.yml&title=Add+%3Clanguage%3E+keyboard) if you don't see it being worked on already!

<a id="featured-by"></a>

# Featured By [`⇧`](#contents)

Please see the [blog posts page on our website](https://scri.be/docs/about/blog-posts) for a list of articles on Scribe, and feel free to open a pull request to add one that you've written at [scribe-org/scri.be](github.com/scribe-org/scri.be)!

### Organizations

The following organizations have supported the development of Scribe projects through various programs. Thank you all! 💙

<div align="center">
  <br>
    <a href="https://tech-news.wikimedia.de/en/2022/03/18/lexicographical-data-for-language-learners-the-wikidata-based-app-scribe/"><img width="180" src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/WikimediaDeutschlandLogo.png" alt="Wikimedia Deutschland logo linking to an article on Scribe in the tech news blog."></a>
    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
    <a href="https://www.mediawiki.org/wiki/New_Developers#Scribe"><img width="180" src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/WikimediaFoundationLogo.png" alt="Wikimedia Foundation logo linking to the MediaWiki new developers page."></a>
    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
  <br>
</div>

<div align="center">
  <br>
    <a href="https://summerofcode.withgoogle.com/"><img width="140" src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/GSoCLogo.png" alt="Google Summer of Code logo linking to its website."></a>
    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
    <a href="https://www.outreachy.org/"><img width="350" src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/OutreachyLogo.png" alt="Outreachy logo linking to its website."></a>
    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
  <br>
</div>

# Powered By [`⇧`](#contents)

### Contributors

Many thanks to all the [Scribe-Server contributors](https://github.com/scribe-org/Scribe-Server/graphs/contributors)! 🚀

<a href="https://github.com/scribe-org/Scribe-Server/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=scribe-org/Scribe-Server" />
</a>

### Code and Dependencies

The Scribe community would like to thank all the great software that made Scribe-Server's development possible.

### Wikimedia Communities

<div align="center">
  <br>
    <a href="https://www.wikidata.org/">
      <img width="240" src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/WikidataLogo.png" alt="Wikidata logo">
    </a>
    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
    <a href="https://www.wikipedia.org/">
      <img width="160" src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/WikipediaLogo.png" alt="Wikipedia logo">
    </a>
    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
    <a href="https://admin.toolforge.org/">
      <img width="175" src="https://raw.githubusercontent.com/scribe-org/Organization/main/resources/images/logos/WikimediaToolforgeLogo.png" alt="Wikimedia Toolforge logo">
    </a>
    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
  <br>
</div>
