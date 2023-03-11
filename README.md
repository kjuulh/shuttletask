# Shuttletask

A task orchestrator for [shuttle](https://github.com/lunarway/shuttle) built in
golang.

The goal of this project is to combine the utility of
[mage](https://github.com/magefile/mage) with shuttle. This is especially useful
when paired with a project like [dagger](https://dagger.io/)

## Usage

Add the dependency using:

```bash
go get github.com/kjuulh/shuttletask
```

or as standalone:

```bash
go install github.com/kjuulh/shuttletask@main
```

```yaml
plan: false
vars:
  name: example
scripts:
  build:
    description: |
      Build the project for fun and profit
    args:
      name: tag
      default: false
    actions:
      - shell: shuttletask build
```

```bash
mkdir -p shuttletask
cd shuttletask
go mod init shuttletask
go get github.com/kjuulh/shuttletask
echo "
//go:build shutletask

package main

func Build(ctx context.Context) error {
  ...
}
" > build.go
```

### Disclaimer

Shuttletask can quite easily be added natively to shuttle, if this project
reaches stability it may be done so, for now it is just a tool which understand
shuttle project files and plans, and can be executed directly.

## How

Shuttletask looks in a shuttletask folder for gofiles. As opposed to each mage,
each file corresponds to a command. Shuttletasks will look in both the root of
the repository, as well as in the downloaded shuttle plan. It expects the
shuttle plan to already be downloaded and won't check if it is missing.

Shuttletask will calculate a hash for each directory, if there is a mismatch it
will copy files together in a tmp directory, a child file takes presedence. This
follows the same principle as shuttle shell actions.

A shuttletask should not call another action directly, but instead invoke it
through:

```golang
shuttletask.Execute("format")
```

Calling the function directly may introduce inconsistency and not work as
intended.
