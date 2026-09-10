//go:build linux || darwin || windows

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	defaultHome, err := defaultCodexHome()
	if err != nil {
		fmt.Fprintf(os.Stderr, "codex-reasoning-monitor: %v\n", err)
		return 1
	}

	flags := flag.NewFlagSet("codex-reasoning-monitor", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	codexHome := flags.String("codex-home", defaultHome, "Codex state directory")
	colorMode := flags.String("color", "auto", "color output: auto, always, or never")
	webEnabled := flags.Bool("web", true, "serve the local dashboard at http://127.0.0.1:5927")
	logDatabase := flags.String("db", "", "history database (default: data/log.sqlite beside the executable)")
	showVersion := flags.Bool("version", false, "print version and exit")
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: %s [options]\n\n", flags.Name())
		fmt.Fprintln(flags.Output(), "Monitor active Codex rollout files and print per-response token usage.")
		fmt.Fprintln(flags.Output(), "\nOptions:")
		flags.PrintDefaults()
	}
	if err := flags.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *showVersion {
		fmt.Println(version)
		return 0
	}

	home, err := filepath.Abs(*codexHome)
	if err != nil {
		fmt.Fprintf(os.Stderr, "codex-reasoning-monitor: resolve Codex home: %v\n", err)
		return 1
	}
	printer, err := newPrinter(os.Stdout, *colorMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "codex-reasoning-monitor: %v\n", err)
		return 2
	}
	defer printer.Close()
	monitor, err := newMonitor(home, printer, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "codex-reasoning-monitor: %v\n", err)
		return 1
	}
	defer monitor.Close()
	logPath := *logDatabase
	if logPath == "" {
		executable, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "codex-reasoning-monitor: locate executable: %v\n", err)
			return 1
		}
		logPath = filepath.Join(filepath.Dir(executable), "data", "log.sqlite")
	}
	logPath, err = filepath.Abs(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "codex-reasoning-monitor: resolve history database: %v\n", err)
		return 1
	}
	monitor.dashboard, err = newPersistentDashboard(logPath, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "codex-reasoning-monitor: open history database: %v\n", err)
		return 1
	}
	defer monitor.dashboard.Close()
	fmt.Fprintln(os.Stderr, "codex-reasoning-monitor: history database "+logPath)
	if *webEnabled {
		server, err := monitor.dashboard.serve()
		if err != nil {
			fmt.Fprintf(os.Stderr, "codex-reasoning-monitor: start dashboard: %v (use --web=false for console only)\n", err)
			return 1
		}
		defer server.Close()
		printer.dashboardNotice(os.Stderr)
	}

	// Go maps Windows Ctrl+C/Break to Interrupt, and console close/logoff to SIGTERM.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := monitor.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "codex-reasoning-monitor: %v\n", err)
		return 1
	}
	return 0
}

func defaultCodexHome() (string, error) {
	if configured := os.Getenv("CODEX_HOME"); configured != "" {
		return configured, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".codex"), nil
}
