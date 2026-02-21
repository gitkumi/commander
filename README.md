# commander

Simple command runner TUI.

Discovers commands from `commander.yaml`, `package.json` scripts, and `Makefile` targets.

## Installation

```
go install github.com/gitkumi/commander@latest
```

## Usage

Run `commander` in a directory with any supported config file.

### commander.yaml

```yaml
commands:
  - title: Ping
    command: ping -c 4 google.com
  - title: Greet
    command: echo "Hello, {{.Name}}"
    inputs:
      - key: Name
        defaultValue: World
  - title: Select
    command: echo "{{.Choice}}"
    inputs:
      - key: Choice
        choices:
          - Option A
          - Option B

environment:
  DB_URL: ./test.db
```
