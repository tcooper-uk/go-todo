# todo

A personal command-line todo list, stored locally in SQLite (or a JSON file, or Firestore).

## Installation

Requires Go 1.20+ and CGo (for the SQLite driver).

```sh
go install github.com/tcooper-uk/go-todo/cmd@latest
```

Or build from source:

```sh
git clone https://github.com/tcooper-uk/go-todo
cd go-todo
go build -o todo ./cmd
```

Data is stored in `~/.todo/` — the directory is created automatically on first run.

## Usage

```
todo [--backend=sqlite|file|cloud] [COMMAND] [FLAGS] [ARGS]
```

Running `todo` with no arguments lists all open items.

### Commands

#### list

```sh
todo list
todo list --all           # include completed items
todo list --done          # only completed items
todo list --priority high # filter by priority: low|medium|high
todo list --tag work      # filter by tag
todo list --overdue       # items past their due date
```

Aliases: `l`, `ls`, `ps`

#### add

```sh
todo add Buy oat milk
todo add --priority high --due 2026-07-01 --tag work Fix the CI pipeline
todo add --tag home --tag errands                    # opens $EDITOR if no text given
```

Flags: `--priority low|medium|high`, `--due YYYY-MM-DD`, `--tag <tag>` (repeatable)

If no item text is provided, your `$EDITOR` opens for you to type the name.

#### edit

```sh
todo edit 3                        # opens $EDITOR pre-filled with current name
todo edit 3 New name here
todo edit 3 --priority medium
todo edit 3 --due 2026-08-01
todo edit 3 --due -                # clears the due date
todo edit 3 --tag work --tag urgent  # replaces all existing tags
todo edit 3 --done true
```

Aliases: `e`, `update`

Flags: `--name`, `--priority`, `--due`, `--tag` (repeatable, replaces existing), `--done true|false`

#### done / reopen

```sh
todo done 3      # mark item 3 as complete
todo reopen 3    # mark item 3 as open again
```

#### delete

```sh
todo delete 3
todo delete 3 5 7   # delete multiple items
```

Aliases: `remove`, `d`, `rm`

#### Other

```sh
todo 3          # show full detail for item 3
todo clearall   # delete everything
todo help       # show command reference
```

### List output

```
ID   St   Pri  Item              Tags          Due             Created
[1]  [ ]  [H]  Fix the CI        #work         Mon 30 Jun 26   Sun 29 Jun 26
[2]  [ ]  [ ]  Buy oat milk      #errands                      Sun 29 Jun 26
[3]  [x]  [L]  Update README                                   Sat 28 Jun 26
```

- **St** — `[ ]` open, `[x]` done
- **Pri** — `[H]` high, `[M]` medium, `[L]` low, `[ ]` none
- Overdue due dates are highlighted in red

## Editor setup

Set `$EDITOR` (or `$VISUAL`) in your shell profile:

```sh
# VS Code
export EDITOR="code --wait"

# Neovim
export EDITOR=nvim

# nano
export EDITOR=nano
```

When `add` or `edit` is called without item text, that editor opens. Save and close the file/tab to submit.

## Backend selection

The default backend is SQLite (`~/.todo/todo.db`). Override at runtime with the `--backend` flag or `$TODO_BACKEND` environment variable:

| Value | Storage |
|---|---|
| `sqlite` / `db` | `~/.todo/todo.db` (default) |
| `file` | `~/.todo/todo.json` |
| `cloud` | Google Firestore (requires `~/.todo/firestore_key.json`) |

```sh
# one-off
todo --backend=file list

# persistent for the session
export TODO_BACKEND=file
todo list
```

### Firestore setup

1. Create a GCP project and enable Firestore.
2. Generate a service account key and save it to `~/.todo/firestore_key.json`.
3. Set the project ID constant in `internal/storage/db/firestore-storage.go` to match your project.

### Migration utilities

If you have existing data to migrate between backends:

```sh
go run ./util/file-to-db       # JSON file → SQLite
go run ./util/db-to-firestore  # SQLite → Firestore
```
