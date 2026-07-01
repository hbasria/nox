package commands

import "regexp"

// dangerPatterns catches command shapes that deserve an extra confirmation
// step even when --auto is set, per nox's "confirm by default" philosophy.
var dangerPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\brm\s+.*-[a-zA-Z]*r[a-zA-Z]*f\b`),
	regexp.MustCompile(`\brm\s+.*-[a-zA-Z]*f[a-zA-Z]*r\b`),
	regexp.MustCompile(`\bdd\s+if=`),
	regexp.MustCompile(`\bmkfs\b`),
	regexp.MustCompile(`\bgit\s+push\s+.*--force\b`),
	regexp.MustCompile(`\bgit\s+push\s+.*-f\b`),
	regexp.MustCompile(`>\s*/dev/sd[a-z]`),
}

// isDangerous reports whether cmd matches a known destructive pattern.
func isDangerous(cmd string) bool {
	for _, p := range dangerPatterns {
		if p.MatchString(cmd) {
			return true
		}
	}
	return false
}
