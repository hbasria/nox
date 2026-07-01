# nox — terminal-native AI assistant

> A CLI-first tool that keeps the user in their existing shell flow while using
> small/fast LLMs for commit message generation, natural-language-to-command
> execution, and error diagnosis.

## 1. Philosophy

- **Never leave the terminal.** No TUI, no alternate screen buffer, no heavy
  spinners. The user stays in their normal shell; `nox` is not a "new
  interface", it's a command they run.
- **Confirm by default, automation optional.** Every command suggestion is
  shown by default as `[Enter] run / [Ctrl+C] cancel`. `--auto` completes the
  whole flow without confirmation.
- **Small, fast models first.** Default models are cheap/fast (e.g. Groq
  `llama-3.1-8b-instant`, small local Ollama models). Provider/model can be
  overridden per command.
- **OpenAI-compatible everywhere.** OpenAI, Groq, Ollama, OVHCloud AI
  Endpoints, etc. — one `chat/completions` interface.
- **Pipe-friendly Unix philosophy.** Chainable like
  `command 2>&1 | nox ask "..."`, `nox commit --dry-run | pbcopy`.

## 2. Tech Stack

| Layer | Choice | Note |
|---|---|---|
| Language | **Go** | Single binary, ~10ms startup, easy distribution (Homebrew/curl install) |
| Config parsing | `pelletier/go-toml` | TOML, human-readable |
| HTTP client | stdlib `net/http` | OpenAI-compatible `/v1/chat/completions` |
| Shell hook | POSIX `precmd`/`preexec` (zsh), `PROMPT_COMMAND` (bash) | OSC 133 compatible, optional |
| Distribution | `curl \| sh` install script + Homebrew tap (later) | similar to the install.sh model in ovh/shai |

**Reference/inspiration project:** [`ovh/shai`](https://github.com/ovh/shai) (Rust, coding agent).
`nox` deliberately stays narrow in scope — no file writing / full agent mode,
targeting only a "terminal command assistant" layer. Ideas borrowed from
`ovh/shai`: project context file, the `on`/`off` toggle name, headless pipe
chaining, low-friction default provider.

## 3. Config Schema (v0 — single file, flat structure)

`~/.config/nox/config.toml` (global) + `.noxconfig.toml` at repo root (project override).

```toml
[default]
provider = "groq"
model = "llama-3.1-8b-instant"
temperature = 0.2
max_tokens = 400

[providers.groq]
base_url = "https://api.groq.com/openai/v1"
api_key_env = "GROQ_API_KEY"

[providers.openai]
base_url = "https://api.openai.com/v1"
api_key_env = "OPENAI_API_KEY"

[providers.ollama]
base_url = "http://localhost:11434/v1"
api_key_env = ""   # empty = no auth

[commands.commit]
provider = "groq"
model = "llama-3.1-8b-instant"
auto_confirm = false

[commands.ask]
provider = "openai"
model = "gpt-4o-mini"

[commands.run]           # natural language -> command
provider = "ollama"
model = "qwen2.5-coder:1.5b"
auto_confirm = false
danger_check = true       # extra confirmation for commands like rm -rf, dd, force-push

[hook]
enabled = false
retention = "24h"         # default 1 day, user-configurable
max_lines_per_command = 200
```

> **Security note:** API keys are never written in plain text to the config
> file — only an `api_key_env` environment variable reference is given.

**v1 plan (future):** Config B (providers separate, commands reference them —
already the model above) + Config C (profile-based: switching via
`--profile` flags like `fast`/`quality`) will be added as layers.

## 4. Commands

### Phase 0-2 scope (MVP)

| Command | Description |
|---|---|
| `nox ask "..."` | Plain Q&A, with or without pipe/hook context |
| `nox commit` | Reads `git diff --staged`, suggests a commit message, runs `git commit -m` on Enter |
| `nox commit --auto` | Commits directly without confirmation |
| `nox "natural language request"` | Generates a shell command from the LLM, shows it, runs it on Enter |
| `nox "natural language request" --auto` | Runs without confirmation |
| `nox on` / `nox off` | Toggle the shell hook on/off |

### Phase 3+ scope

| Command | Description |
|---|---|
| `nox fix` | If the last command's exit code ≠ 0, auto-diagnoses and suggests a fix |
| `nox cleanup` | Scans for unnecessary files (node_modules, __pycache__, etc.), deletes with confirmation |
| `nox explain <command>` | Explains a complex command line-by-line without running it |

### Next wave (backlog)

`pr`, `branch`, `conflict`, `changelog`, `deps`, `audit`, `test-fail`, `ci-log`,
natural language port/process queries (`"who's using port 8080"`), `kill`,
`man` summaries — see Section 6.

## 5. Shell Hook (Context Capture)

Default retention **1 day**, user-configurable via `hook.retention`
(e.g. `"6h"`, `"3d"`).

- Optional install: `nox init-shell` (or `nox on`) adds a `precmd`/`preexec`
  hook to `.zshrc`/`.bashrc`.
- After each command: the command text + short output (stdout/stderr,
  line-limited) + exit code are written to a rotating buffer
  (`~/.cache/nox/history.log`).
- If the hook isn't installed, `nox ask`/`nox fix` do the same job manually
  **via pipe**: `command 2>&1 | nox ask "what does this error mean"`.
- Dangerous command guard: commands like `rm -rf`, `dd`, `mkfs`,
  `git push --force` require extra confirmation even under `--auto`
  (Enter isn't enough, you must type "yes").

## 6. TODO — Simple to Complex

### Phase 0 — Skeleton
- [ ] Go project setup (`nox`), TOML parsing, single-file config loading
- [ ] OpenAI-compatible client (chat completions, provider agnostic)
- [ ] `nox ask "..."` — plain context-free Q&A

### Phase 1 — Core features
- [ ] `nox commit` — `git diff --staged` → generate message → commit on Enter
- [ ] `--auto` global flag
- [ ] Pipe support (auto context when stdin is present)

### Phase 2 — Natural language → command execution
- [ ] `nox "natural language command"` (general purpose find/ps/lsof-style)
- [ ] Dangerous command guard (rm -rf, dd, force-push)
- [ ] Multi-step plan confirmation (multiple commands → list → confirm one-by-one/all)

### Phase 3 — Context enrichment
- [ ] `nox init-shell` / `nox on` / `nox off` — hook install and toggle
  - [ ] Rotating buffer, default retention **1 day**, configurable
- [ ] `nox fix` — check last command's exit code, suggest a diagnosis
- [ ] `nox cleanup` — unnecessary file cleanup (with confirmation)
- [ ] Project-based context awareness (package.json/go.mod detection)
- [ ] Project context file support (`NOX.md` — similar to ovh/shai's `SHAI.md`)

### Phase 4 — Polish / trust
- [ ] `nox explain <command>`
- [ ] Learning commit style from git log (Conventional Commits vs. free-form)
- [ ] Model routing/escalation (simple task → small model, complex → large model)
- [ ] `--verbose` transparency mode (model, token, timing info)
- [ ] Offline fallback (fall back to Ollama if no key/internet)
- [ ] Zero-config first experience (free/rate-limited default provider)

### Phase 5 — Advanced
- [ ] Audit log + `nox undo`
- [ ] Session flag (optional multi-turn conversation context, `--session <name>`)
- [ ] Config v1: profile-based (`fast`/`quality`) + provider layer separation
- [ ] Backlog commands: `pr`, `branch`, `conflict`, `changelog`, `deps`, `audit`,
      `test-fail`, `ci-log`, port/process queries, `kill`, `man` summaries

## 7. Naming Decision

- ~~`shai`~~ — conflicts with `ovh/shai` (Rust, 600+ stars, active), dropped.
- ~~`wisp` / `shx` / `aish` / `nsh`~~ — considered in round two, philosophy-aligned
  but dropped after a "short Jarvis-style name" request.
- ~~`otto` / `juno` / `ivy` / `axl` / `kai`~~ — round three, `ivy` dropped for
  being too common.
- ~~`tars`~~ — too crowded a name space (multiple active projects), dropped.
- **`nox`** ✅ — final decision. Short (3 letters), fluent (`nox commit`,
  `nox ask`, `nox on`), matches the "quiet background helper" philosophy.

**Open item:** the `nox` package name hasn't been checked for conflicts on
npm / crates.io / PyPI / Homebrew yet — must be verified before Phase 0 starts.
