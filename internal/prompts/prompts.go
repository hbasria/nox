// Package prompts holds nox's system prompts. Kept as a plain embedded text
// file so it stays easy to swap out later (e.g. merging with .crux agent
// prompts) without touching Go code.
package prompts

import _ "embed"

//go:embed system.txt
var system string

// System returns the base assistant system prompt.
func System() string {
	return system
}

// CommandGen appends task-specific instructions for turning a natural
// language request into a single runnable shell command. memoryCtx is the
// contents of the user's ~/.nox/memory.md file (system facts, notes), used
// so the model prefers tools that actually exist on this machine.
func CommandGen(memoryCtx string) string {
	return system + `
Task: translate the user's natural language request into a single POSIX-compliant shell command.
- Return ONLY the command. No explanation, no markdown code fence (` + "```" + `), no extra text.
- Chain multi-step operations with "&&" on a single line.
- If unsure, pick the most reasonable/common interpretation. Do not overthink this; decide quickly.
- Prefer commands and tools that are native to the user's OS (see memory below). For example, on macOS prefer lsof/netstat over Linux-only tools like ss/ip.
- The command itself is not a reply to the user, so write it in the shell's syntax, not in any particular human language.
- If the request has no meaningful shell command translation (small talk, a greeting, a general question), respond with a single "echo '<short reply, in the user's language>'" command instead of refusing or leaving output empty.
- If, and only if, there is a new durable fact worth remembering for next time (a tool that's now installed, a preference the user stated, a system quirk you noticed, or something the user explicitly asked you to remember), add exactly one extra line after the command, formatted as:
MEMORY: <short fact, in English>
Skip this entirely when there's nothing new and durable to note; most requests won't need it. Never invent facts that weren't stated or directly observed.

What nox remembers about this machine:
` + memoryCtx
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
