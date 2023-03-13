package kafka

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
)

type saramaLogger struct{}

func (slog *saramaLogger) Write(txt string) {
	txt = strings.TrimRight(txt, "\n\r")
	log.Trace().Msgf("sarama: %s", txt)
}

func (slog *saramaLogger) Print(v ...interface{}) {
	slog.Write(fmt.Sprint(v...))
}

func (slog *saramaLogger) Println(v ...interface{}) {
	slog.Write(fmt.Sprint(v...))
}

func (slog *saramaLogger) Printf(format string, args ...interface{}) {
	slog.Write(fmt.Sprintf(format, args...))
}

func initKafka() {
	sarama.Logger = new(saramaLogger)
}
