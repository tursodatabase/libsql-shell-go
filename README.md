# libSQL Shell

libSQL shell is a standalone program written in Go for querying SQLite database files and libSQL databases running [sqld](https://github.com/libsql/sqld).

## Table of Contents

- [libSQL Shell](#libsql-shell)
  - [Table of Contents](#table-of-contents)
  - [Running the shell](#running-the-shell)
    - [Query SQLite database files](#query-sqlite-database-files)
    - [Query sqld](#query-sqld)
    - [Query a Turso database](#query-a-turso-database)
    - [Built-in help](#built-in-help)
  - [Development](#development)
    - [Install git hooks](#install-git-hooks)
    - [Install golangci-lint](#install-golangci-lint)
    - [Configure tests](#configure-tests)

## Running the shell

After cloning this repo, you can start the shell using `go run` or create an executable using `go build`.

There is only one argument: the database to query. When the shell is connected to the database, it allows interactive SQL queries.

### Query SQLite database files

To start a shell that queries a [SQLite database file](https://www.sqlite.org/fileformat.html), provide a path to the file:

```sh
go run ./cmd/libsql-shell/main.go my_libsql.db
```

### Query sqld

To start a shell that queries a libSQL database running sqld, provide the database connection URL. For example, for sqld running locally:

```sh
go run ./cmd/libsql-shell/main.go ws://127.0.0.1:8080
```

### Query a Turso database

To query sql managed by [Turso](https://turso.tech), you need a database URL and token. Assuming you have already logged in to the CLI and created a database:

1. Get your database URL: `turso db show <db_name>`. The URL will be on the format `libsql://<db_name>-<username>.turso.io`
1. Create a database token: `turso db tokens create <db_name>`

Add the token to the URL in the `authToken` query string parameter and provide that to the shell:

```sh
go run ./cmd/libsql-shell/main.go libsql://<db_name>-<username>.turso.io/?authToken=<db_token>`
```

### Built-in help

The shell has built-in commands similar to the [SQLite CLI](https://www.sqlite.org/cli.html). Get a list of commands with `.help`.

## Development

### Install git hooks

Git Hooks are shell scripts that run automatically before or after Git executes an important command, such as *Commit* or *Push*. To install it, run the command:

```bash
./scripts/git/install-git-hooks.sh
```

### Install golangci-lint

The above git hooks require `golangci-lint` in your PATH. [Install golangci-lint](https://golangci-lint.run/usage/install/).

### Configure tests

The tests are configured through `.test_config.yaml` file. Copy `.test_config.yaml.example` to create your `.test_config.yaml` and see below the description for each field:

  - `sqld_db_uri`: URL for sqld database used during the tests. It should be a dedicated database once it's cleared after each test. In particular, for turso databases, follow the steps from [Running the Project/Turso](#Turso) to generate a Turso db path. For this database, it's useful to generate a non-expiration token with `turso db tokens create --expiration none <turso_test_db_name>`
    If you're working with a local server pass the local address inside this env variable. In case it's a local file, provide the file path.
  - `skip_sqld_tests`: Used to skip SQLD related tests

### Running tests against a locally-running sqld instance

An easy way to get an SQLD server locally is by running the remote docker image:

``` shell
docker run -p 8080:8080 -d ghcr.io/libsql/sqld:latest
```

This command runs the latest version available, exposing the server on (http://localhost:8080). You can checkout other versions
[here](https://github.com/libsql/sqld/pkgs/container/sqld).
Now, in `libsql-shell-go`, edit your `test_config.yaml` file by adding the URL of the locally running sqld server under `sqld_db_uri`. Remember to set the variable `skip_sqld_tests` to `false`, otherwise, sqld-related tests will be skipped.  Once you have this configuration and the sqld instance running simply run the tests normally:

``` shell
go test -v ./...
```

You can confirm that the sqld instance was hit by checking the server logs.
