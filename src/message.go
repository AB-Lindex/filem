package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"

	"github.com/AB-Lindex/filem/src/helpers"
	"github.com/AB-Lindex/filem/src/message"
	"github.com/AB-Lindex/filem/src/message/kafka"
)

type messageConfig struct {
	Format string        `yaml:"format"`
	Kafka  *kafka.Config `yaml:"kafka"`
}

type msgData struct {
	vars map[string]interface{}
}

var (
	errNoReceiver = helpers.StringError("no receiver available")
)

func (cfg *messageConfig) Connect(dryRun bool) (message.Sender, error) {

	switch true {
	case cfg.Kafka != nil:
		return cfg.Kafka.Connect(dryRun)

	default:
		return nil, errNoReceiver
	}
}

func (msg *msgData) AddVar(key string, value interface{}) {
	if msg.vars == nil {
		msg.vars = make(map[string]interface{})
	}
	msg.vars[key] = value
}

func (msg *msgData) Render() []byte {

	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)

	enc := json.NewEncoder(wr)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(msg.vars)
	wr.Flush()

	return buf.Bytes()
}

func (msg *msgData) DumpVars() {

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(msg.vars)

	// buf, _ := json.MarshalIndent(msg.vars, "", "  ")
	// fmt.Println(string(buf))
}
