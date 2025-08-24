package main

import (
	"flag"
	"log"

	"github.com/fiffeek/hyprdynamicmonitors/internal/config"
	"github.com/fiffeek/hyprdynamicmonitors/internal/detectors"
	"github.com/fiffeek/hyprdynamicmonitors/internal/generators"
	"github.com/fiffeek/hyprdynamicmonitors/internal/hypr"
	"github.com/fiffeek/hyprdynamicmonitors/internal/matchers"
	"github.com/fiffeek/hyprdynamicmonitors/internal/service"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	var (
		configPath = flag.String("config", "$HOME/.config/hyprdynamicmonitors/config.toml", "Path to configuration file")
		dryRun     = flag.Bool("dry-run", false, "Show what would be done without making changes")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
		showVer    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showVer {
		log.Printf("hypr-dynmon %s (commit: %s, built: %s)", version, commit, buildDate)
		return
	}

	if !*verbose {
		log.SetFlags(0)
	}

	log.Printf("Starting Hyprland Dynamic Monitor Manager v%s", version)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	hyprIPC, err := hypr.NewIPC(*verbose)
	if err != nil {
		log.Fatalf("Failed to initialize Hyprland IPC: %v", err)
	}

	monitorDetector, err := detectors.NewMonitorDetector(hyprIPC)
	if err != nil {
		log.Fatalf("Failed to initialize MonitorDetector: %v", err)
	}

	powerDetector, err := detectors.NewPowerDetector(*verbose)
	if err != nil {
		log.Fatalf("Failed to initialize PowerDetector with file path %v", err)
	}

	matcher := matchers.NewMatcher(cfg, *verbose)

	generator := generators.NewConfigGenerator(&generators.GeneratorConfig{
		Verbose: *verbose,
	})

	svc := service.NewService(cfg, monitorDetector, powerDetector, &service.Config{
		DryRun:  *dryRun,
		Verbose: *verbose,
	}, matcher, generator)

	if err := svc.Run(); err != nil {
		log.Fatalf("Service failed: %v", err)
	}
}
