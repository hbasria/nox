package main

import (
	"fmt"
	"os"
	"strings"

	"nox/internal/commands"
	"nox/internal/config"
)

func usage() {
	fmt.Println(`nox — terminal-native AI assistant

Usage:
  nox "natural language request"   generates a shell command, runs it on Enter
  nox "request" --auto             runs without confirmation (except dangerous commands)
  nox commit                       generates a commit message from the staged diff
  nox commit --auto                commits without confirmation

Flags:
  --auto                           skip confirmation (except dangerous commands)
  --verbose                        print the raw request/response sent to the model`)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		return
	}

	auto := false
	verbose := false
	rest := args[:0:0]
	for _, a := range args {
		switch a {
		case "--auto":
			auto = true
		case "--verbose":
			verbose = true
		default:
			rest = append(rest, a)
		}
	}

	if len(rest) == 0 {
		usage()
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}

	switch rest[0] {
	case "commit":
		err = commands.Commit(cfg, auto, verbose)
	case "help", "-h", "--help":
		usage()
		return
	default:
		err = commands.RunNL(cfg, strings.Join(rest, " "), auto, verbose)
	}

	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "nox:", err)
	os.Exit(1)
}
