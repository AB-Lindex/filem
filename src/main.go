package main

import (
	"os"
	"runtime"

	"github.com/AB-Lindex/filem/src/metrics"
	"github.com/rs/zerolog/log"
)

func main() {
	args.Load()

	// buf, _ := json.MarshalIndent(cfg, "", "  ")
	// fmt.Println(string(buf))
	metric, err := cfg.Metrics.Connect(args.DryRun)
	if err != nil {
		log.Error().Msgf("metrics error: %v", err)
		os.Exit(1)
	}

	_ = metric.Set(metrics.MetricsStart, metrics.Now(), nil)
	_ = metric.Set(metrics.VersionInfo, 1, map[string]interface{}{
		"version":    versionFunc(),
		"go_version": runtime.Version(),
	})
	_ = metric.Send()

	store, err := cfg.Storage.Connect(args.DryRun)
	if err != nil {
		log.Error().Msgf("storage error: %v", err)
		os.Exit(1)
	}

	sender, err := cfg.Message.Connect(args.DryRun)
	if err != nil {
		log.Error().Msgf("sender error: %v", err)
		os.Exit(1)
	}
	defer sender.Close()

	n := 0
	for _, folder := range cfg.Folders {
		n += folder.Process(store, sender, metric, args.DryRun)
	}
	log.Info().Msgf("%d files processed", n)

	_ = metric.Set(metrics.MetricsEnd, metrics.Now(), nil)
	err = metric.Send()
	if err != nil {
		log.Error().Msgf("metrics-error: %v", err)
	}
}
