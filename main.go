package main

import (
	"fmt"
	"os"
	"strings"

	"nox/internal/commands"
	"nox/internal/config"
)

func usage() {
	fmt.Println(`nox — terminal-native AI asistanı

Kullanım:
  nox "doğal dil isteği"     istekten shell komutu üretir, Enter'da çalıştırır
  nox "istek" --auto         onaysız çalıştırır (tehlikeli komutlar hariç)
  nox commit                 staged diff'ten commit mesajı üretir
  nox commit --auto          onaysız commit eder`)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		return
	}

	auto := false
	rest := args[:0:0]
	for _, a := range args {
		if a == "--auto" {
			auto = true
			continue
		}
		rest = append(rest, a)
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
		err = commands.Commit(cfg, auto)
	case "help", "-h", "--help":
		usage()
		return
	default:
		err = commands.RunNL(cfg, strings.Join(rest, " "), auto)
	}

	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "nox:", err)
	os.Exit(1)
}
