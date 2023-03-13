package main

import (
	"bytes"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/ninlil/envsubst"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type argStruct struct {
	Config  string `arg:"-f,--file" default:"filem.yaml" help:"Configfile"`
	DryRun  bool   `arg:"--dry-run" help:"Perform a dry-run / what-if"`
	Verbose bool   `arg:"-v" help:"Verbose logging"`
}

func (argStruct) Version() string {
	return "filem v?"
}

type configStruct struct {
	Folders []folderStruct `yaml:"folders"`
	Storage storageConfig  `yaml:"storage"`
	Message messageConfig  `yaml:"message"`
}

var args argStruct
var cfg configStruct

func printErr(err error, prefixes ...string) bool {
	if err == nil {
		return false
	}

	prefix := strings.Join(prefixes, " - ")
	log.Error().Msgf("error in %s: %v", prefix, err)

	return true
}

func (a *argStruct) Load() {
	arg.MustParse(a)

	f, err := os.Open(args.Config)
	if printErr(err, "read config") {
		os.Exit(1)
	}
	defer f.Close()

	var buf bytes.Buffer
	_ = envsubst.Convert(f, &buf, nil)

	dec := yaml.NewDecoder(bytes.NewReader(buf.Bytes()))
	dec.KnownFields(true)
	err = dec.Decode(&cfg)
	if printErr(err, "parse config") {
		os.Exit(1)
	}

	if !args.Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

}
