# nox

A terminal-native AI assistant. `nox` translates natural language into shell
commands, writes commit messages from your staged diff, and stays out of your
way otherwise — no TUI, no alternate screen buffer, just a command you run
from your normal shell.

See [AGENTS.md](AGENTS.md) for the full design/roadmap.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/hbasria/nox/main/install.sh | sh
```

This detects your OS/architecture and downloads the matching binary from
the [latest GitHub release](https://github.com/hbasria/nox/releases/latest)
into `/usr/local/bin` (override with `NOX_INSTALL_DIR`).

## Build from source

```sh
just build     # builds ./nox for your current OS/architecture
just release   # cross-compiles darwin/linux × amd64/arm64 into dist/
```

## First run

The first time you run `nox`, it creates `~/.nox/config.toml` with an
OpenAI-shaped default:

```toml
[default]
provider = "openai"
model = "gpt-4o-mini"
temperature = 0.2
max_tokens = 400

[providers.openai]
base_url = "https://api.openai.com/v1"
api_key_env = "OPENAI_API_KEY"
```

Set your key as an environment variable (recommended):

```sh
export OPENAI_API_KEY="sk-..."
```

...or store it directly in the config file for convenience:

```toml
[providers.openai]
base_url = "https://api.openai.com/v1"
api_key = "sk-..."
```

Any OpenAI-compatible endpoint works — Groq, Ollama, or a custom gateway.
Just point `base_url`/`model` at it and switch `provider` in `[default]`.

`nox` also keeps a small memory file at `~/.nox/memory.md` — see
[Memory](#memory) below.

## Memory

`~/.nox/memory.md` is a plain markdown file that `nox` reads on every
natural language request and feeds to the model as context. On first run
it's seeded with facts detected about your machine:

```markdown
# nox memory

Facts nox knows about this machine. Edit freely; nox reads this file as-is.

- OS: macOS (darwin)
- Architecture: arm64
- Shell: zsh
- Package manager: brew
```

This is why `nox "list open ports"` suggests `lsof` on macOS instead of the
Linux-only `ss` — the model knows what OS it's talking to before it answers.

The model can also add to memory on its own. If it learns something durable
worth keeping — a tool that's now installed, a preference you stated, or
something you explicitly asked it to remember — it appends a note and tells
you right away:

```
$ ./nox "remember that I prefer conventional commit messages"
nox: remembered: User prefers conventional commit messages
$ echo 'Understood.'
[Enter] run  [Ctrl+C] cancel
```

That note now lives in `~/.nox/memory.md` and will be included as context in
every future request. It's just a text file, so you can open it and
edit/prune it by hand any time — nothing parses its structure beyond
handing it whole to the model.

## Example scenarios

### Natural language → shell command

```
$ ./nox "list open ports"
$ lsof -i -P -n | grep LISTEN
[Enter] run  [Ctrl+C] cancel
```

Press Enter to run it, or Ctrl+C to back out. Add `--auto` to skip the
confirmation (dangerous commands like `rm -rf` or `git push --force` still
require typing "yes", even with `--auto`):

```
$ ./nox "find files larger than 100MB in this repo" --auto
```

### Commit messages from your staged diff

```
$ git add -A
$ ./nox commit
Proposed commit message:
---
Add pipe support so context flows between chained nox calls
---
[Enter] commit  [Ctrl+C] cancel
```

`--auto` commits without asking.

### Missing tools get an install suggestion first

If the generated command needs a binary that isn't installed, `nox` offers
to install it instead of just failing:

```
$ ./nox "list open ports"
"ss" was not found on this system. Suggested install:
$ brew install iproute2mac
[Enter] run  [Ctrl+C] cancel
```

### Plain Q&A with `nox ask`

Never generates or runs a command — just answers, optionally using piped
input as context:

```
$ ./nox ask "what does exit code 137 usually mean"
$ echo "top RAM user: Zen Browser, 4GB" | ./nox ask "why might this be using so much RAM"
```

### Explaining a command with `nox explain`

Breaks a command down in plain language without running it:

```
$ ./nox explain "find . -name '*.log' -mtime +30 -delete"
```

### Chaining nox calls with pipes

Piped input isn't run — it's handed to the model as context, and also
forwarded to the generated command's own stdin:

```
$ ./nox "which app is using the most RAM" --auto | ./nox "what is this app"
```

### Debugging with --verbose

Prints the raw request sent to the model and the raw response received,
useful when a provider returns something unexpected (e.g. empty content,
truncated output):

```
$ ./nox commit --verbose
```

## Flags

| Flag | Effect |
|---|---|
| `--auto` | Skip the run/commit confirmation (dangerous commands still require typing "yes" on an interactive terminal) |
| `--verbose` | Print the raw LLM request/response to stderr |
| `--format` | Shape command output into readable columns (set `default.format = true` in config to make this the default) |
