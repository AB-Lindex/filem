package main

import (
	stdlog "log"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		NoColor:    true,
		TimeFormat: time.DateTime,
	})

	//zlog := zerolog.New(os.Stdout)

	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)
}
