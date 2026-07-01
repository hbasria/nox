package commands

import (
	"os/exec"
	"regexp"
	"strings"
)

// envAssignment matches a leading VAR=value token (e.g. "FOO=bar cmd").
var envAssignment = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// validBinaryName matches characters that can plausibly appear in a real
// command name (as opposed to a stray fragment of awk/sed script that
// slipped past segment splitting, e.g. "$2=\"\""). This is a defense-in-depth
// safety net: even if splitSegments mis-parses something, we won't treat
// garbage as an installable "binary".
var validBinaryName = regexp.MustCompile(`^[A-Za-z0-9_.\-/]+$`)

// missingBinaries returns the set of command names referenced in cmdStr that
// don't resolve via the shell (covers PATH binaries, builtins, and keywords).
func missingBinaries(cmdStr string) []string {
	seen := map[string]bool{}
	var missing []string

	for _, segment := range splitSegments(cmdStr) {
		fields := strings.Fields(segment)
		i := 0
		for i < len(fields) && envAssignment.MatchString(fields[i]) {
			i++
		}
		if i >= len(fields) {
			continue
		}
		bin := fields[i]
		if bin == "sudo" && i+1 < len(fields) {
			bin = fields[i+1]
		}
		if bin == "" || seen[bin] || !validBinaryName.MatchString(bin) {
			continue
		}
		seen[bin] = true
		if !binaryExists(bin) {
			missing = append(missing, bin)
		}
	}
	return missing
}

// splitSegments breaks a shell command into its chained segments on &&, ||,
// ;, and | — but only when those tokens appear outside of single/double
// quotes. This matters because command generation often produces awk/sed
// one-liners whose quoted script bodies contain their own ";" and "|"
// characters (e.g. awk '{a; b}'), which must NOT be treated as shell-level
// separators.
func splitSegments(cmdStr string) []string {
	var segments []string
	var current strings.Builder
	inSingle, inDouble := false, false

	runes := []rune(cmdStr)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch {
		case c == '\'' && !inDouble:
			inSingle = !inSingle
			current.WriteRune(c)
		case c == '"' && !inSingle:
			inDouble = !inDouble
			current.WriteRune(c)
		case c == '\\' && !inSingle && i+1 < len(runes):
			current.WriteRune(c)
			i++
			current.WriteRune(runes[i])
		case !inSingle && !inDouble && c == '&' && i+1 < len(runes) && runes[i+1] == '&':
			segments = append(segments, current.String())
			current.Reset()
			i++
		case !inSingle && !inDouble && c == '|' && i+1 < len(runes) && runes[i+1] == '|':
			segments = append(segments, current.String())
			current.Reset()
			i++
		case !inSingle && !inDouble && (c == '|' || c == ';'):
			segments = append(segments, current.String())
			current.Reset()
		default:
			current.WriteRune(c)
		}
	}
	segments = append(segments, current.String())
	return segments
}

func binaryExists(bin string) bool {
	return exec.Command("sh", "-c", "command -v "+shellQuote(bin)).Run() == nil
}

// shellQuote wraps s in single quotes, safe for embedding in a `sh -c` string.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
