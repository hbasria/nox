// Package prompts holds nox's system prompts. Kept as a plain embedded text
// file so it stays easy to swap out later (e.g. merging with .crux agent
// prompts) without touching Go code.
package prompts

import (
	_ "embed"
	"runtime"
)

//go:embed system.txt
var system string

//go:embed quirks_darwin.txt
var quirksDarwin string

// System returns the base assistant system prompt.
func System() string {
	return system
}

// platformQuirks returns a curated list of known command-line
// incompatibilities for the current OS (e.g. GNU vs BSD flag differences).
// Unlike memory.md, this is built into nox rather than learned per-machine,
// since it's true for every user on this OS, not just this one.
func platformQuirks() string {
	switch runtime.GOOS {
	case "darwin":
		return quirksDarwin
	default:
		return ""
	}
}

// CommandGen appends task-specific instructions for turning a natural
// language request into a single runnable shell command. memoryCtx is the
// contents of the user's ~/.nox/memory.md file (system facts, notes), used
// so the model prefers tools that actually exist on this machine. When
// formatted is true, the command should shape its own output into clean,
// human-readable columns instead of a tool's raw dump.
func CommandGen(memoryCtx string, formatted bool) string {
	p := system + `
Task: translate the user's natural language request into a single POSIX-compliant shell command.
- Return ONLY the command. No explanation, no markdown code fence (` + "```" + `), no extra text.
- Chain multi-step operations with "&&" on a single line.
- If unsure, pick the most reasonable/common interpretation. Do not overthink this; decide quickly.
- Prefer commands and tools that are native to the user's OS (see memory below). For example, on macOS prefer lsof/netstat over Linux-only tools like ss/ip.
- The command itself is not a reply to the user, so write it in the shell's syntax, not in any particular human language.
- If the request asks for "the most/least/highest/lowest X" or "which one is Y-est", generate a command that sorts by the relevant metric and limits output to just the top result. Don't return a raw, unsorted, multi-row listing (e.g. plain "top") for these — the answer must be unambiguous from the output alone.
- If the request has no meaningful shell command translation (small talk, a greeting, a general question), respond with a single "echo '<short reply, in the user's language>'" command instead of refusing or leaving output empty.`

	if formatted {
		p += `
- Shape the command's own output into clean, human-readable columns instead of a raw tool dump. Names and paths often contain spaces (e.g. "Zen Browser.app"), which silently breaks naive column-based printing/sorting — use this exact worked pattern as your template, adapting the ps fields/metric/sort direction/row count to the actual request:
ps -eo rss,comm | sort -k1,1nr | head -n 10 | awk '{printf "%8.1f MB  ", $1/1024; $1=""; print substr($0,2)}'
Why this shape works: the sort key (a plain number) comes first and is sorted with "-k1,1" (bounded to just that field, so later spaces don't break it); the name is printed as "the rest of the line" (substr($0,2) after blanking $1), never a single positional field like $2 or $3 alone, which would truncate multi-word names. Only include fields the request actually needs (e.g. don't add pid if it's irrelevant). Limit rows to a reasonable count (e.g. top 10) unless the user asked for everything.`
	}

	p += `
- If, and only if, there is a new durable fact worth remembering for next time (a tool that's now installed, a preference the user stated, a system quirk you noticed, or something the user explicitly asked you to remember), add exactly one extra line after the command, formatted as:
MEMORY: <short fact, in English>
Skip this entirely when there's nothing new and durable to note; most requests won't need it, and never repeat a fact that's already listed under "What nox remembers" below. Never invent facts that weren't stated or directly observed.
`

	if quirks := platformQuirks(); quirks != "" {
		p += "\nKnown platform quirks:\n" + quirks
	}

	return p + "\nWhat nox remembers about this machine:\n" + memoryCtx
}

// CommitMsg appends instructions for generating a commit message from a
// staged git diff.
func CommitMsg() string {
	return system + `
Task: generate a short, focused commit message from the given "git diff --staged" output.
- Return ONLY the commit message. No quotes, no explanation, no markdown.
- First line should be a ~50-72 character summary; add a short body after a blank line if needed.
- Match the diff's language/style if it strongly implies one; default to English if unsure.`
}

// InstallCmd appends instructions for suggesting a shell command that
// installs a missing binary, using memoryCtx for OS/package-manager facts.
func InstallCmd(memoryCtx string) string {
	return system + `
Task: the user is missing a command-line tool. Give the single shell command that installs it on this machine.
- Return ONLY the install command. No explanation, no markdown code fence, no extra text.
- Use the package manager noted in memory below; if none is noted, pick the most standard option for the OS.
- If truly unsure, return the most common install command for that tool on this OS.

What nox remembers about this machine:
` + memoryCtx
}
