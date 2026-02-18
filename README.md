# iLaunch

Production-grade CLI/TUI utility for bootstrapping Node.js projects.

## Features

- Environment checks (`node`, `npm`/`pnpm`, Node.js >= 18).
- Event-driven TUI built with Bubble Tea + Lip Gloss.
- Non-blocking subprocess runner with streamed logs.
- `.env` creation from `.env.example` with validation.
- Dependency installation (`pnpm` preferred over `npm`).
- Git initialization workflow.
- `--non-interactive` mode for CI.

## Project structure

```text
cmd/
  root.go
internal/
  app/
    model.go
    run.go
    update.go
    view.go
  env/
    parser.go
    writer.go
  runner/
    process.go
  system/
    checks.go
  ui/
    components/
main.go
```

## Build & run

```bash
make build
./bin/ilaunch
```

or:

```bash
go run .
```

Non-interactive mode:

```bash
go run . --non-interactive
```

## Controls (TUI)

- `↑` / `↓`: navigate
- `Enter`: select
- `Esc`: go back / exit
- `Ctrl+C`: graceful shutdown

## Notes

- Ensure `.env.example` exists before using "Create .env file".
- In non-interactive mode, `.env` is generated from default values in `.env.example`.
