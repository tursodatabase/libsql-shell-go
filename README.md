# libSQL Shell

## Table of Contents

- [libSQL Shell](#libsql-shell)
  - [Table of Contents](#table-of-contents)
  - [Setup](#setup)
    - [Install git hooks](#install-git-hooks)

## Setup

### Install git hooks

Git Hooks are shell scripts that run automatically before or after Git executes an important command, such as *Commit* or *Push*. To install it, run the command:

```bash
./scripts/git/install-git-hooks.sh
```

### Configure tests

The tests are configured through `.test_config.yaml` file. Copy `.test_config.yaml.example` to create your `.test_config.yaml` and see below the description for each field:
  - `turso_db_path`: URL for turso database used during the tests. It should be a dedicated database once it's cleared after each test
  - `skip_turso_tests`: Used to skip Turso related tests
