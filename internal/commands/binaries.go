package commands

import (
	"os/exec"
	"regexp"
	"strings"
)

// segmentSplit breaks a shell command into its chained segments on &&, ||, ;, |.
var segmentSplit = regexp.MustCompile(`&&|\|\||[;|]`)

// envAssignment matches a leading VAR=value token (e.g. "FOO=bar cmd").
var envAssignment = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// missingBinaries returns the set of command names referenced in cmdStr that
// don't resolve via the shell (covers PATH binaries, builtins, and keywords).
func missingBinaries(cmdStr string) []string {
	seen := map[string]bool{}
	var missing []string

	for _, segment := range segmentSplit.Split(cmdStr, -1) {
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
		if bin == "" || seen[bin] {
			continue
		}
		seen[bin] = true
		if !binaryExists(bin) {
			missing = append(missing, bin)
		}
	}
	return missing
}

func binaryExists(bin string) bool {
	return exec.Command("sh", "-c", "command -v "+shellQuote(bin)).Run() == nil
}

// shellQuote wraps s in single quotes, safe for embedding in a `sh -c` string.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
