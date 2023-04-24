# libSQL Shell

## Table of Contents

- [libSQL Shell](#libsql-shell)
  - [Table of Contents](#table-of-contents)
  - [Setup](#setup)
    - [Install git hooks](#install-git-hooks)

## Running the project

We can run the project using `go run` or creating an executable binary using `go build`. The only mandatory argument that you need to pass is `DB`, that can be a file path to a libsql/SQLite database or the URL to a Turso database.

### libSQL/SQLite
To start a shell for **libSQL/SQLite** databases just point to your database location
```sh
go run ./cmd/libsql-shell/main.go my_libsql.db
```

### Turso
To start a shell for **Turso** databases, you need your database URL and a database token:

1. Login on Turso: `turso auth login`
2. Create a database: `turso db create`
3. Get your database URL: `turso db show <db_name>`. The URL will be on the format `libsql://<db_name>-<username>.turso.io`
4. Create a database token: `turso db tokens create <db_name>`

Then, to connect to Turso database you need to pass the database token as query parameter of URL, like the following command:
```sh
go run ./cmd/libsql-shell/main.go libsql://<db_name>-<username>.turso.io/?jwt=<db_token>`
```

## Development

### Install git hooks

Git Hooks are shell scripts that run automatically before or after Git executes an important command, such as *Commit* or *Push*. To install it, run the command:

```bash
./scripts/git/install-git-hooks.sh
```

### Configure tests

The tests are configured through `.test_config.yaml` file. Copy `.test_config.yaml.example` to create your `.test_config.yaml` and see below the description for each field:
  - `turso_db_path`: URL for turso database used during the tests. It should be a dedicated database once it's cleared after each test. Follow steps from [Running the Project/Turso](#Turso) to generate a Turso db path. For this database, it's useful to generate a non-expiration token with `turso db tokens create --expiration none <turso_test_db_name>`
  - `skip_turso_tests`: Used to skip Turso related tests
