package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"net/http"
	"syscall"
	"time"
	"sync"

	"ganttgen/internal/calendar"
	"ganttgen/internal/csvinput"
	"ganttgen/internal/renderer"
	"ganttgen/internal/scheduler"
)

func main() {
	var output string
	var holidaysPath string
	var watch bool
	var liveReload bool
	var liveReloadPort int
	flag.StringVar(&output, "o", "gantt.html", "output HTML file")
	flag.StringVar(&output, "output", "gantt.html", "output HTML file")
	flag.StringVar(&holidaysPath, "holidays", "", "optional YAML file listing YYYY-MM-DD holidays")
	flag.BoolVar(&watch, "watch", false, "watch input CSV and regenerate on changes")
	flag.BoolVar(&liveReload, "livereload", false, "enable livereload server and inject client script")
	flag.IntVar(&liveReloadPort, "livereload-port", 35729, "port for livereload server (default 35729)")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: ganttgen [--output file] [--holidays file] [--watch] [--livereload] [--livereload-port port] <input.csv>\n")
		os.Exit(1)
	}
	input := args[0]

	var lr *liveReloader
	liveReloadURL := ""
	if liveReload {
		var err error
		lr, liveReloadURL, err = startLiveReload(liveReloadPort)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to start livereload: %v\n", err)
			os.Exit(1)
		}
		watch = true // livereload implies watch for change events
	}

	if err := generate(input, output, holidaysPath, liveReloadURL); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("generated %s\n", output)

	if watch {
		if err := watchAndGenerate(input, output, holidaysPath, liveReloadURL, lr); err != nil {
			fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
			os.Exit(1)
		}
	}
}

func generate(input, output, holidaysPath, liveReloadURL string) error {
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

	html, err := renderer.BuildHTML(scheduled, liveReloadURL)
	if err != nil {
		return fmt.Errorf("error rendering HTML: %w", err)
	}

	if err := os.WriteFile(output, []byte(html), 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func watchAndGenerate(input, output, holidaysPath, liveReloadURL string, lr *liveReloader) error {
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
			if err := generate(input, output, holidaysPath, liveReloadURL); err != nil {
				fmt.Fprintf(os.Stderr, "regenerate failed: %v\n", err)
				continue
			}
			fmt.Printf("generated %s\n", output)
			if lr != nil {
				lr.Reload()
			}
		case <-sigCh:
			fmt.Println("stop watching")
			return nil
		}
	}
}

type liveReloader struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

func startLiveReload(port int) (*liveReloader, string, error) {
	lr := &liveReloader{
		clients: make(map[chan struct{}]struct{}),
	}

	mux := http.NewServeMux()
	mux.Handle("/livereload", lr)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "livereload server error: %v\n", err)
		}
	}()

	url := fmt.Sprintf("http://%s/livereload", addr)
	fmt.Printf("livereload server listening on %s\n", url)
	return lr, url, nil
}

func (lr *liveReloader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client := make(chan struct{}, 1)
	lr.mu.Lock()
	lr.clients[client] = struct{}{}
	lr.mu.Unlock()

	// Send initial ping to establish connection
	fmt.Fprintf(w, "data: ping\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			lr.mu.Lock()
			delete(lr.clients, client)
			lr.mu.Unlock()
			return
		case <-client:
			fmt.Fprintf(w, "data: reload\n\n")
			flusher.Flush()
		}
	}
}

func (lr *liveReloader) Reload() {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	for ch := range lr.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
