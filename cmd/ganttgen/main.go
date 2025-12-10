package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ganttgen/internal/calendar"
	"ganttgen/internal/csvinput"
	"ganttgen/internal/renderer"
	"ganttgen/internal/scheduler"
)

func main() {
	var output string
	var holidaysPath string
	var watch bool
	flag.StringVar(&output, "o", "gantt.html", "output HTML file")
	flag.StringVar(&output, "output", "gantt.html", "output HTML file")
	flag.StringVar(&holidaysPath, "holidays", "", "optional YAML file listing YYYY-MM-DD holidays")
	flag.BoolVar(&watch, "watch", false, "watch input CSV and regenerate on changes")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: ganttgen [--output file] [--holidays file] [--watch] <input.csv>\n")
		os.Exit(1)
	}
	input := args[0]

	if err := generate(input, output, holidaysPath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("generated %s\n", output)

	if watch {
		if err := watchAndGenerate(input, output, holidaysPath); err != nil {
			fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
			os.Exit(1)
		}
	}
}

func generate(input, output, holidaysPath string) error {
	if holidaysPath != "" {
		if err := calendar.LoadHolidaysYAML(holidaysPath); err != nil {
			return fmt.Errorf("failed to load holidays: %w", err)
		}
	}

	tasks, err := csvinput.Read(input)
	if err != nil {
		return fmt.Errorf("error reading CSV: %w", err)
	}

	scheduled, err := scheduler.Schedule(tasks)
	if err != nil {
		return fmt.Errorf("error scheduling tasks: %w", err)
	}

	html, err := renderer.BuildHTML(scheduled)
	if err != nil {
		return fmt.Errorf("error rendering HTML: %w", err)
	}

	if err := os.WriteFile(output, []byte(html), 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func watchAndGenerate(input, output, holidaysPath string) error {
	info, err := os.Stat(input)
	if err != nil {
		return fmt.Errorf("stat input: %w", err)
	}
	lastMod := info.ModTime()
	lastSize := info.Size()

	fmt.Printf("watching %s for changes (Ctrl+C to stop)...\n", input)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			info, err := os.Stat(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "watch: stat failed: %v\n", err)
				continue
			}
			if info.ModTime() == lastMod && info.Size() == lastSize {
				continue
			}

			lastMod = info.ModTime()
			lastSize = info.Size()

			fmt.Printf("[%s] change detected, regenerating...\n", time.Now().Format("15:04:05"))
			if err := generate(input, output, holidaysPath); err != nil {
				fmt.Fprintf(os.Stderr, "regenerate failed: %v\n", err)
				continue
			}
			fmt.Printf("generated %s\n", output)
		case <-sigCh:
			fmt.Println("stop watching")
			return nil
		}
	}
}
