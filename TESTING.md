<a id="top"></a>

# Testing conventions

Scribe-Server uses Go's standard `testing` package, with `testify` for assertions. These conventions keep the test suite consistent as coverage is added incrementally.

## Contents

- [Organizing tests](#organizing-tests)
- [Assertions and helpers](#assertions-and-helpers)
- [Isolating global state](#isolating-global-state)
- [HTTP handlers](#http-handlers)
- [Databases and fixtures](#databases-and-fixtures)
- [Running tests and coverage](#running-tests-and-coverage)
- [Local verification before committing](#local-verification-before-committing)
- [Pull request checklist](#pull-request-checklist)

## Organizing tests

- Keep tests beside the code they cover in package-local `*_test.go` files. Do not create a separate top-level test directory.
- Use the same package name as the code when a test needs package-private access. Use an external `<package>_test` package when testing only the public API provides useful separation.
- Name test files after the source or behavior they cover, such as `language_validator_test.go`.
- Name top-level tests `Test<Subject>` or `Test<Subject>_<Scenario>`. Use descriptive subtest names for individual cases.
- Prefer table-driven tests with `t.Run` when several cases share the same setup and assertion structure.

For example, tests for the language validator stay in the same directory as the production code:

```text
api/validators/
├── language_validator.go
└── language_validator_test.go
```

`language_validator_test.go` can use the same `validators` package and group related inputs in a table:

```go
package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidTranslationLangCode(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{name: "two-letter code", code: "de", want: true},
		{name: "three-letter code", code: "pnb", want: true},
		{name: "uppercase code", code: "DE", want: false},
		{name: "empty code", code: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidTranslationLangCode(tt.code))
		})
	}
}
```

<sub><a href="#top">Back to top.</a></sub>

## Assertions and helpers

- Use `require` when a failed check must stop the current test or subtest, such as failed setup, an unexpected error, or a required response body.
- Use `assert` when later checks are still meaningful after a failure.
- Mark reusable test helpers with `t.Helper()` so failures point to the calling test.
- Register teardown with `t.Cleanup` as soon as a test acquires a resource.
- Use `t.Setenv` for test-specific environment variables and `t.TempDir` for temporary files. Do not depend on a contributor's machine, home directory, or persistent files.

In this example, a missing or invalid fixture stops the test with `require`; once setup succeeds, `assert` checks independent properties and reports all mismatches:

```go
func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return data
}

func TestErrorResponseFixture(t *testing.T) {
	data := readFixture(t, "error_response.json")

	var response models.ErrorResponse
	require.NoError(t, json.Unmarshal(data, &response))

	assert.NotEmpty(t, response.Error)
	assert.Equal(t, "invalid language code", response.Error)
}
```

<sub><a href="#top">Back to top.</a></sub>

## Isolating global state

Tests must be independent and produce the same result in any order.

- Put Gin in test mode when testing routes or handlers. Restore its prior mode with `t.Cleanup` when a test changes it.
- Prefer a new Viper instance created with `viper.New()` over package-level Viper state. If production code requires the global instance, call `viper.Reset()` during setup and cleanup.
- Do not call `t.Parallel()` in tests that mutate Gin mode, global Viper state, package globals, process-wide environment, or shared storage.
- Avoid real clocks, random results, external networks, and order-dependent assertions. Inject or fix those inputs when they affect behavior.

Use a helper to make the lifecycle obvious. A test that calls this helper must not call `t.Parallel()`:

```go
func isolateGlobalState(t *testing.T) *viper.Viper {
	t.Helper()

	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() {
		gin.SetMode(previousMode)
	})

	// Prefer this isolated instance when production code accepts one.
	config := viper.New()
	config.Set("GIN_MODE", "test")

	// If the code under test uses package-level Viper state, reset it both
	// before and after the test so another test cannot influence this one.
	viper.Reset()
	t.Cleanup(viper.Reset)

	return config
}

func TestConfigUsesTestEnvironment(t *testing.T) {
	config := isolateGlobalState(t)
	t.Setenv("ENV", "test")

	assert.Equal(t, "test", os.Getenv("ENV"))
	assert.Equal(t, "test", config.GetString("GIN_MODE"))
}
```

<sub><a href="#top">Back to top.</a></sub>

## HTTP handlers

- Test Gin handlers with `net/http/httptest`; a running server is not required.
- Build the smallest router needed for the behavior under test.
- Assert the HTTP status, relevant headers, and decoded response body.
- Replace databases, files, and external services with isolated dependencies where the production design permits it.

For example, `HandleError` can be exercised through a minimal Gin router and an in-memory response recorder:

```go
func TestHandleError(t *testing.T) {
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		HandleError(c, http.StatusBadRequest, "invalid language code")
	})

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))

	var response models.ErrorResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	assert.Equal(t, "invalid language code", response.Error)
}
```

<sub><a href="#top">Back to top.</a></sub>

## Databases and fixtures

- Keep unit tests independent of MariaDB and other external services.
- Treat a test that requires a database as an integration test. Give it isolated setup and teardown, and document how to run it.
- Do not make the default unit-test suite depend on a developer's local database or existing data.
- Store static fixtures in a `testdata/` directory next to the package that uses them. Keep small expected values inline when that makes the test easier to read.
- Never modify a checked-in fixture during a test. Copy it to `t.TempDir()` first when mutation is required.

For example, contract fixtures used by handler tests belong to the handler package:

```text
api/handlers/
├── language.go
├── language_test.go
└── testdata/
    └── de.yaml
```

Use a helper like this when a test needs to pass a writable fixture to production code:

```go
func writableFixture(t *testing.T, name string) string {
	t.Helper()

	source := filepath.Join("testdata", name)
	data, err := os.ReadFile(source)
	require.NoError(t, err)

	destination := filepath.Join(t.TempDir(), filepath.Base(name))
	require.NoError(t, os.WriteFile(destination, data, 0o600))
	return destination
}

func TestLoadSingleContract(t *testing.T) {
	contractPath := writableFixture(t, "de.yaml")

	contract, err := loadSingleContract(filepath.Dir(contractPath), "de")
	require.NoError(t, err)
	assert.Contains(t, contract, "de")

	// The checked-in testdata/de.yaml file remains unchanged.
}
```

Database integration tests should be clearly separated from the default unit suite. For example, put this build constraint at the top of `connection_integration_test.go`:

```go
//go:build integration

package database

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestDatabaseConnection_Integration(t *testing.T) {
	dsn := os.Getenv("SCRIBE_TEST_DATABASE_DSN")
	require.NotEmpty(t, dsn, "set SCRIBE_TEST_DATABASE_DSN to an isolated test database")

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	require.NoError(t, db.Ping())
}
```

Run explicitly tagged integration tests only after preparing an isolated database:

```bash
go test -tags=integration ./database/...
```

<sub><a href="#top">Back to top.</a></sub>

## Running tests and coverage

Run all tests before opening a pull request:

```bash
make test
```

Generate `coverage.out` and print coverage by function:

```bash
make test-cover
```

For an HTML coverage report after running `make test-cover`:

```bash
go tool cover -html=coverage.out
```

<sub><a href="#top">Back to top.</a></sub>

## Local verification before committing

Changes to Go code or the testing infrastructure should not introduce unexpected dependency changes. Run:

```bash
go mod tidy
git diff --exit-code go.mod go.sum
```

If `go mod tidy` changes `go.mod` or `go.sum` unexpectedly, inspect why before committing. Keep intentional dependency changes and explain them in the pull request.

Run formatting and lint checks as well:

```bash
make fmt
make lint
```

The lint target requires [`revive`](https://github.com/mgechev/revive). If it is not installed, the Makefile prints the installation command.

A complete local verification sequence is:

```bash
git status

make fmt
make test
make test-cover
make lint

go tool cover -func=coverage.out | tail -1

git status
git diff
```

Pull requests that add tests should report the total coverage before and after the change, for example `Coverage: 3.4% -> 5.1%`. Coverage should increase incrementally. If it does not, explain why in the pull request.

CI runs the same coverage target. A test must not rely on local configuration, network access, execution order, or files outside the repository.

<sub><a href="#top">Back to top.</a></sub>

## Pull request checklist

- Tests follow the package, naming, and table-driven conventions above.
- Success, failure, and relevant edge cases are covered.
- Test state and resources are isolated and cleaned up.
- `make test` passes locally.
- `make test-cover` passes locally.
- The pull request includes the before and after total coverage.

<sub><a href="#top">Back to top.</a></sub>
