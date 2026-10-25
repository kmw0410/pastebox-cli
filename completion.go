package main

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed packaging/completions/_pb
var zshCompletion string

//go:embed packaging/completions/pb.bash
var bashCompletion string

//go:embed packaging/completions/pb.fish
var fishCompletion string

func (a application) runCompletion(args []string) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(a.stdout, completionUsageText)
		return 0
	}
	if len(args) != 1 {
		fmt.Fprint(a.stderr, completionUsageText)
		return 2
	}

	var script string
	switch strings.ToLower(args[0]) {
	case "zsh":
		script = zshCompletion
	case "bash":
		script = bashCompletion
	case "fish":
		script = fishCompletion
	default:
		fmt.Fprintf(a.stderr, "unsupported shell %q; expected zsh, bash, or fish\n", args[0])
		return 2
	}
	fmt.Fprint(a.stdout, script)
	return 0
}
