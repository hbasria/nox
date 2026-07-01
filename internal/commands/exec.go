package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// isInteractive reports whether stdin is an interactive terminal, as
// opposed to a pipe or redirected file.
func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// confirmAndRun shows cmdStr to the user and runs it through the shell once
// confirmed. Dangerous commands always require typing "yes" on an
// interactive terminal, regardless of auto; there's no way to confirm those
// safely when stdin isn't a terminal, so they're skipped outright. Other
// commands run immediately when auto is set or stdin isn't interactive
// (e.g. piped input), otherwise wait for an Enter keypress (Ctrl+C cancels).
//
// stdinData, if non-nil, is used as the executed command's stdin instead of
// nox's own os.Stdin (e.g. to forward piped input to the generated command).
func confirmAndRun(cmdStr string, auto bool, stdinData io.Reader) error {
	fmt.Printf("$ %s\n", cmdStr)

	interactive := isInteractive()

	if isDangerous(cmdStr) {
		if !interactive {
			fmt.Println("this looks like a dangerous command; skipping (no interactive terminal to confirm).")
			return nil
		}
		fmt.Print(`This may be a dangerous command. Type "yes" to continue: `)
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		if strings.TrimSpace(line) != "yes" {
			fmt.Println("cancelled.")
			return nil
		}
	} else if !auto && interactive {
		fmt.Print("[Enter] run  [Ctrl+C] cancel ")
		reader := bufio.NewReader(os.Stdin)
		if _, err := reader.ReadString('\n'); err != nil {
			return nil
		}
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	if stdinData != nil {
		cmd.Stdin = stdinData
	} else {
		cmd.Stdin = os.Stdin
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
