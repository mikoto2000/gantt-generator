package main

import (
	"flag"
	"fmt"
	"os"

	"ganttgen/internal/csvinput"
	"ganttgen/internal/renderer"
	"ganttgen/internal/scheduler"
)

func main() {
	var output string
	flag.StringVar(&output, "o", "gantt.html", "output HTML file")
	flag.StringVar(&output, "output", "gantt.html", "output HTML file")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: ganttgen [--output file] <input.csv>\n")
		os.Exit(1)
	}
	input := args[0]

	tasks, err := csvinput.Read(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading CSV: %v\n", err)
		os.Exit(1)
	}

	scheduled, err := scheduler.Schedule(tasks)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error scheduling tasks: %v\n", err)
		os.Exit(1)
	}

	html, err := renderer.BuildHTML(scheduled)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error rendering HTML: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(output, []byte(html), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("generated %s\n", output)
}
