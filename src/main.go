package main

import (
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	args.Load()

	// buf, _ := json.MarshalIndent(cfg, "", "  ")
	// fmt.Println(string(buf))

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
		n += folder.Process(store, sender, args.DryRun)
	}
	log.Info().Msgf("%d files processed", n)
}
