// Command skycast shows the current weather as an animated ASCII scene.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/YOUR_USERNAME/skycast/internal/config"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	cfg, err := config.Parse(args, os.Getenv, stderr)
	if errors.Is(err, config.ErrHelp) {
		fmt.Fprint(stdout, config.Usage)
		return 0
	}
	var ue *config.UsageError
	if errors.As(err, &ue) {
		fmt.Fprintf(stderr, "skycast: %v\nRun 'skycast --help' for usage.\n", err)
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "skycast: %v\n", err)
		return 1
	}
	if cfg.Version {
		fmt.Fprintf(stdout, "skycast %s\n", version)
		return 0
	}
	// NOTE: the app itself is wired up in later phases.
	fmt.Fprintln(stdout, "skycast: not implemented yet")
	return 0
}
