package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// confirmAndRun shows cmdStr to the user and runs it through the shell once
// confirmed. Dangerous commands always require typing "yes", regardless of
// auto. Non-dangerous commands run immediately when auto is set, otherwise
// wait for an Enter keypress (Ctrl+C cancels).
func confirmAndRun(cmdStr string, auto bool) error {
	fmt.Printf("$ %s\n", cmdStr)

	reader := bufio.NewReader(os.Stdin)

	if isDangerous(cmdStr) {
		fmt.Print(`Bu tehlikeli bir komut olabilir. Devam etmek için "yes" yaz: `)
		line, _ := reader.ReadString('\n')
		if strings.TrimSpace(line) != "yes" {
			fmt.Println("iptal edildi.")
			return nil
		}
	} else if !auto {
		fmt.Print("[Enter] çalıştır  [Ctrl+C] iptal ")
		if _, err := reader.ReadString('\n'); err != nil {
			return nil
		}
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
