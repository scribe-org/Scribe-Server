# Changelog

See the [releases for Scribe-Server](https://github.com/scribe-org/Scribe-Server/releases) for an up to date list of versions and their release dates.

Scribe tries to follow [semantic versioning](https://semver.org/), a MAJOR.MINOR.PATCH version where increments are made of the:

- MAJOR version when we make incompatible API changes
- MINOR version when we add functionality in a backwards compatible manner
- PATCH version when we make backwards compatible bug fixes

Emojis for the following are chosen based on [gitmoji](https://gitmoji.dev/).

## Scribe-Server 1.0.0

### 🚀 Deployment

- Scribe-Server is now deployed to Wikimedia Toolforge with an automated data update and deployment workflow ([#40](https://github.com/scribe-org/Scribe-Server/pull/40)).
  - The service can be found at [scribe-server.toolforge.org](https://scribe-server.toolforge.org/)
  - A repository check was added to the update workflow to warn if it's run from an incorrect repo ([#55](https://github.com/scribe-org/Scribe-Server/pull/55)).
  - Matrix notifications were added so the team is alerted on data update workflow runs ([#54](https://github.com/scribe-org/Scribe-Server/pull/54)).
- A Toolforge deployment guide was added for maintainers ([#63](https://github.com/scribe-org/Scribe-Server/pull/63)).
- The Toolforge build was fixed to compile PyICU with the correct Toolforge ICU paths ([#64](https://github.com/scribe-org/Scribe-Server/pull/64), [#60](https://github.com/scribe-org/Scribe-Server/pull/60)).

### ✨ Features

- Scribe-Server's REST API was rebuilt on the [Gin](https://gin-gonic.com/) framework with CORS support, replacing the original `net/http` implementation ([#29](https://github.com/scribe-org/Scribe-Server/pull/29)).
  - Versioned language data API endpoints were added as part of this migration.
- An `update_data.sh` script was added to run Scribe-Data within Scribe-Server, along with language validation enhancements ([#35](https://github.com/scribe-org/Scribe-Server/pull/35)).
- Resulting data can now be filtered based on data contracts ([#42](https://github.com/scribe-org/Scribe-Server/pull/42)).
  - Data contracts were switched from JSON to YAML for easier maintenance ([#56](https://github.com/scribe-org/Scribe-Server/pull/56)).
- Statistics for available languages are now shown via the API ([#45](https://github.com/scribe-org/Scribe-Server/pull/45)).
- An entry/landing page was set up for Scribe-Server ([#49](https://github.com/scribe-org/Scribe-Server/pull/49)), followed by a dedicated deployment and download page ([#52](https://github.com/scribe-org/Scribe-Server/pull/52)).
- A translation data retrieval endpoint with validation logic was implemented ([#61](https://github.com/scribe-org/Scribe-Server/pull/61)).
- OpenAPI/Swagger documentation generation was added for the API, viewable at [scribe-server.toolforge.org/swagger/index.html](https://scribe-server.toolforge.org/swagger/index.html) and [scribe-server.toolforge.org/docs/index.html](https://scribe-server.toolforge.org/docs/index.html) ([#39](https://github.com/scribe-org/Scribe-Server/pull/39)).

### 🐞 Bug Fixes

- Large Go files were split up and marked for maintainability ([#47](https://github.com/scribe-org/Scribe-Server/pull/47)).
- Data extraction and update script issues were fixed following the contract-based filtering rollout, including validation logic and deprecated test failures.
- The `update_data.sh` download command was fixed to specify the correct dump snapshot location ([#65](https://github.com/scribe-org/Scribe-Server/pull/65)).
- Trusted proxies are now explicitly configured for security.
- German data contract fixes for missing `displayValue` fields on declensions, and fully indexed conjugations/declensions.

### 📝 Documentation

- Documentation was expanded and reorganized throughout the [README](README.md) and [CONTRIBUTING guide](CONTRIBUTING.md), including environment setup, deployment testing, and mentorship/growth sections.
- Available data and download instructions were linked from the README.

### ♻️ Code Refactoring

- The migration/database package structure was refactored to improve the database migration process.
- `handlers.go` was restructured into separate concerns, and `database.go` was split into smaller files.
- The SQLite driver was switched from `mattn/go-sqlite3` to `glebarez/sqlite`.
- CI linting was consolidated into the existing workflow ([#14](https://github.com/scribe-org/Scribe-Server/pull/14)), [revive](https://github.com/mgechev/revive) linting was integrated for local development ([#34](https://github.com/scribe-org/Scribe-Server/pull/34)), and a pre-commit `gofmt` hook was added ([#26](https://github.com/scribe-org/Scribe-Server/pull/26)).
- Unneeded `sqlc` and OpenAPI generation libraries were removed in favor of a simplified stack.

### ✅ Tests

- Basic health and CORS test scaffolding was added for the API ([#29](https://github.com/scribe-org/Scribe-Server/pull/29)).
- The CI workflow was updated to select the Go version via `go-version-file` ([#15](https://github.com/scribe-org/Scribe-Server/pull/15)) and later improved further ([#31](https://github.com/scribe-org/Scribe-Server/pull/31)).

### ⚖️ Legal

- SPDX license identifiers were added across the project, with the license header check switched over to `spdx-checker` in CI.

### ⬆️ Dependencies

- Local development environment support was added via [`air`](https://github.com/air-verse/air) for hot reload ([#21](https://github.com/scribe-org/Scribe-Server/pull/21)).
- Docker Compose based local tooling for MariaDB was added ([#13](https://github.com/scribe-org/Scribe-Server/pull/13)).
