package commands

import (
	"io"
	"os"
)

// maxPipeChars caps how much piped stdin gets sent to the model as context.
const maxPipeChars = 8000

// readPipedInput reads stdin's content when it's piped/redirected (not an
// interactive terminal). The second return value reports whether stdin was
// piped at all, regardless of whether any content was read.
func readPipedInput() (string, bool) {
	if isInteractive() {
		return "", false
	}

	data, err := io.ReadAll(io.LimitReader(os.Stdin, maxPipeChars+1))
	if err != nil {
		return "", true
	}

	s := string(data)
	if len(s) > maxPipeChars {
		s = s[:maxPipeChars] + "\n... (piped input truncated)"
	}
	return s, true
}
