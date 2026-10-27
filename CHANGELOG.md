# Changelog

See the [releases for Scribe-Server](https://github.com/scribe-org/Scribe-Server/releases) for an up to date list of versions and their release dates.

Scribe tries to follow [semantic versioning](https://semver.org/), a MAJOR.MINOR.PATCH version where increments are made of the:

- MAJOR version when we make incompatible API changes
- MINOR version when we add functionality in a backwards compatible manner
- PATCH version when we make backwards compatible bug fixes

Emojis for the following are chosen based on [gitmoji](https://gitmoji.dev/).

## Scribe-Server 1.0.0

### 🚀 Deployment

- Scribe-Server is now deployed to Wikimedia Toolforge with an automated data update and deployment workflow ([#36](https://github.com/scribe-org/Scribe-Server/issues/36), [#37](https://github.com/scribe-org/Scribe-Server/issues/37), [#38](https://github.com/scribe-org/Scribe-Server/issues/38)).
  - The service can be found at [scribe-server.toolforge.org](https://scribe-server.toolforge.org/)
  - A repository check was added to the update workflow to warn if it's run from an incorrect repo ([#55](https://github.com/scribe-org/Scribe-Server/pull/55)).
  - Matrix notifications were added so the team is alerted on data update workflow runs ([#41](https://github.com/scribe-org/Scribe-Server/issues/41)).
- The Toolforge build was fixed to compile PyICU with the correct Toolforge ICU paths ([#43](https://github.com/scribe-org/Scribe-Server/issues/43)).
- Deployments only keep the two most recent backups of the SQLite data on Toolforge.
- The Toolforge webservice was switched from the deprecated `jdk17` image to the `trixie` image, and the server is built with `CGO_ENABLED=0` so that it doesn't depend on the image's system libraries ([#107](https://github.com/scribe-org/Scribe-Server/issues/107)).

### ✨ Features

- Scribe-Server's REST API was rebuilt on the [Gin](https://gin-gonic.com/) framework with CORS support, replacing the original `net/http` implementation ([#5](https://github.com/scribe-org/Scribe-Server/issues/5)).
  - Versioned language data API endpoints were added as part of this migration.
  - The Scribe-Server API is able to be easily versioned for future changes to response structures ([#24](https://github.com/scribe-org/Scribe-Server/issues/24)).
- An `update_data.sh` script was added to run Scribe-Data within Scribe-Server, along with language validation enhancements ([#28](https://github.com/scribe-org/Scribe-Server/issues/28)).
- Scribe-Data data contracts are served via Scribe-Server to tell the client applications how to process the data ([#18](https://github.com/scribe-org/Scribe-Server/issues/18), [#57](https://github.com/scribe-org/Scribe-Server/issues/57)).
  - Data contracts were switched from JSON to YAML for easier maintenance ([#56](https://github.com/scribe-org/Scribe-Server/pull/56)).
- Statistics for available languages are now shown via the API ([#44](https://github.com/scribe-org/Scribe-Server/issues/44)).
  - Language statistics include counts of prepositions and profanity, which are null for languages without them.
- An entry/landing page was set up for Scribe-Server ([#48](https://github.com/scribe-org/Scribe-Server/issues/48)), followed by a dedicated deployment and download page ([#52](https://github.com/scribe-org/Scribe-Server/pull/52)).
- A translation data retrieval endpoint with validation logic was implemented ([#58](https://github.com/scribe-org/Scribe-Server/issues/58), [#59](https://github.com/scribe-org/Scribe-Server/issues/59)).
- SQLite databases are available for download directly from the Scribe-Server UI ([#6](https://github.com/scribe-org/Scribe-Server/issues/6)).
- A profanity table is also served alongside the other data from Wikidata to allow client applications to filter out these words ([#75](https://github.com/scribe-org/Scribe-Server/issues/75)).
- Prepositions are served for German and Russian so client applications can show the grammatical case they require ([#98](https://github.com/scribe-org/Scribe-Server/issues/98)).

### 🐞 Bug Fixes

- Large Go files were split up and marked for maintainability ([#46](https://github.com/scribe-org/Scribe-Server/issues/46)).
- Data extraction and update script issues were fixed following the contract-based filtering rollout, including validation logic and deprecated test failures.
- The `updated_at` date for language data is now the most recent last modified date of the data, matching `/data-version`, rather than the date of the request.
- Data contracts exported from Scribe-Data are now copied to the server and deployed to Toolforge with the data so that `/contracts` returns them, and a clear error is returned if they're not available ([#113](https://github.com/scribe-org/Scribe-Server/issues/113)).
  - `/contracts?lang={lang}` returns only that language's contract exactly as exported by Scribe-Data as JSON (e.g. `en.json`), keeping the order of the contract's keys.
  - `/contracts` returns a `scribe_data_contracts.zip` archive with a `scribe_data_contracts` directory of JSON contracts for all languages.
- The `update_data.sh` download command was fixed to specify the correct dump snapshot location ([#65](https://github.com/scribe-org/Scribe-Server/pull/65)).
- Trusted proxies are now explicitly configured for security.
- `/packs` requests with `..` in the path are rejected so they can't show whether directories outside of the packs directory exist, `/packs/sqlite/list` returns `[]` instead of `null` when there are no SQLite files, and SQLite download file names are quoted ([#91](https://github.com/scribe-org/Scribe-Server/issues/91)).
- Database queries for languages, data types, table schemas and table data return an error instead of partial results if reading rows fails partway through ([#92](https://github.com/scribe-org/Scribe-Server/issues/92)).
- The SQLite to MariaDB migration tool exits with an error if any table fails to migrate rather than exiting successfully so that the data update workflow fails, and reading SQLite tables and schemas returns an error instead of partial results if reading rows fails partway through ([#119](https://github.com/scribe-org/Scribe-Server/issues/119)).
- German data contract fixes for missing `displayValue` fields on declensions, and fully indexed conjugations/declensions.

### 📝 Documentation

- Documentation was expanded and reorganized throughout the [README.md](README.md) and [CONTRIBUTING guide](CONTRIBUTING.md), including environment setup, deployment testing, and mentorship/growth sections.
- Available data and download instructions were linked from the [README.md](README.md).
- Documentation on how to install and deploy the tools needed for Scribe-Server on Wikimedia's Toolforge was added in [TOOLFORGE.md](TOOLFORGE.md) ([#62](https://github.com/scribe-org/Scribe-Server/issues/62)).
- OpenAPI/Swagger documentation generation was added for the API, viewable at [scribe-server.toolforge.org/swagger/index.html](https://scribe-server.toolforge.org/swagger/index.html) and [scribe-server.toolforge.org/docs/index.html](https://scribe-server.toolforge.org/docs/index.html) ([#9](https://github.com/scribe-org/Scribe-Server/issues/9), [#19](https://github.com/scribe-org/Scribe-Server/issues/19)).

### 🎨 Design

- The website for Scribe-Server has been developed with a modern design with easy to follow documentation ([#50](https://github.com/scribe-org/Scribe-Server/issues/50), [#71](https://github.com/scribe-org/Scribe-Server/issues/71)).
- Icons and favicons were added to make the website presentable ([#33](https://github.com/scribe-org/Scribe-Server/issues/33)).

### ✅ Tests

- Basic health and CORS test scaffolding was added for the API ([#29](https://github.com/scribe-org/Scribe-Server/pull/29)).
- The CI workflow was updated to select the Go version via `go-version-file` ([#8](https://github.com/scribe-org/Scribe-Server/issues/8)) and later improved further ([#30](https://github.com/scribe-org/Scribe-Server/issues/30)).
- Unit tests were added for the language validator functions ([#78](https://github.com/scribe-org/Scribe-Server/issues/78)).
- Unit tests were added for the YAML normalizer function ([#81](https://github.com/scribe-org/Scribe-Server/issues/81)).
- Unit tests were added for the available languages handler ([#84](https://github.com/scribe-org/Scribe-Server/issues/84)).
- Unit tests were added for the translation data retrieval endpoint handler ([#85](https://github.com/scribe-org/Scribe-Server/issues/85)).
- Unit tests were added for the table name validators ([#80](https://github.com/scribe-org/Scribe-Server/issues/80)).
- A CI check was added to verify that generated Swagger docs are up to date ([#82](https://github.com/scribe-org/Scribe-Server/issues/82)).
- A CI check was added that fails if the total test coverage is below a minimum, which is raised as tests are added ([#117](https://github.com/scribe-org/Scribe-Server/issues/117)).
- Unit tests were added for the language statistics retrieval endpoint handler ([#86](https://github.com/scribe-org/Scribe-Server/issues/86)).
- Unit tests were added for the `/packs` static file and SQLite download handler, including directory traversal attempts, and for the Swagger documentation routes ([#91](https://github.com/scribe-org/Scribe-Server/issues/91)).
- Unit tests were added for API route registration, the home page, server startup and configuration loading ([#90](https://github.com/scribe-org/Scribe-Server/issues/90)).
- Unit tests were added for the database package using [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) ([#92](https://github.com/scribe-org/Scribe-Server/issues/92)).
- Unit tests were added for the SQLite to MariaDB migration tool, reading test SQLite databases and checking the MariaDB queries with go-sqlmock ([#119](https://github.com/scribe-org/Scribe-Server/issues/119)).
- Unit tests were added for the language data retrieval endpoint handler ([#87](https://github.com/scribe-org/Scribe-Server/issues/87)).
- Unit tests were added for the language version retrieval endpoint handler ([#89](https://github.com/scribe-org/Scribe-Server/issues/89)).
- Unit tests were added for the dbqueries package using [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) ([#130](https://github.com/scribe-org/Scribe-Server/issues/130)).

### ♻️ Code Refactoring

- The migration/database package structure was refactored to improve the database migration process.
- `handlers.go` was restructured into separate concerns, and `database.go` was split into smaller files.
- The SQLite driver was switched from `mattn/go-sqlite3` to `glebarez/sqlite`.
- CI linting was consolidated into the existing workflow ([#14](https://github.com/scribe-org/Scribe-Server/pull/14)), [revive](https://github.com/mgechev/revive) linting was integrated for local development ([#32](https://github.com/scribe-org/Scribe-Server/issues/32)), and a pre-commit `gofmt` hook was added ([#25](https://github.com/scribe-org/Scribe-Server/issues/25)).
- Unneeded `sqlc` and OpenAPI generation libraries were removed in favor of a simplified stack.
- The table name validators use precompiled patterns, and checks that the patterns already enforce were removed.

### ⚖️ Legal

- SPDX license identifiers were added across the project, with the license header check switched over to `spdx-checker` in CI.

### ⬆️ Dependencies

- Local development environment support was added via [`air`](https://github.com/air-verse/air) for hot reload ([#20](https://github.com/scribe-org/Scribe-Server/issues/20)).
- Docker Compose based local tooling for MariaDB was added ([#7](https://github.com/scribe-org/Scribe-Server/issues/7)).
