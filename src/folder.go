package main

import (
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AB-Lindex/filem/src/message"
	"github.com/AB-Lindex/filem/src/storage"
	"github.com/ninlil/envsubst"
	"github.com/rs/zerolog/log"
)

type folderStruct struct {
	Location    string                 `yaml:"location"`
	Filter      string                 `yaml:"filter"`
	TargetName  string                 `yaml:"targetName"`
	Format      string                 `yaml:"format"`
	ContentType string                 `yaml:"contentType"`
	OnSuccess   string                 `yaml:"onSuccess"`
	Tags        map[string]interface{} `yaml:"tags"`

	regexp   *regexp.Regexp
	varNames []string

	deleteOnSuccess bool
}

func (f *folderStruct) Process(store storage.Storage, sender message.Sender, dryRun bool) (fileCount int) {

	f.OnSuccess = strings.ToLower(f.OnSuccess)
	switch f.OnSuccess {
	case "delete":
		f.deleteOnSuccess = true
	case "":
	default:

	}

	files, err := os.ReadDir(f.Location)
	if err != nil {
		log.Error().Msgf("unable to search '%s': %v", f.Location, err)
		return
	}

	rx, err := regexp.Compile(f.Filter)
	if err != nil {
		log.Error().Msgf("unable to parse filter '%s': %v", f.Filter, err)
		return
	}
	f.regexp = rx
	f.varNames = rx.SubexpNames()
	if len(f.varNames) > 0 {
		f.varNames[0] = "_filename"
	}

	log.Info().Msgf("Searching '%s'...", f.Location)
	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				log.Error().Msgf("unable to check '%s': %v", file.Name(), err)
			} else {
				fileCount += f.handleFile(store, sender, info, dryRun)
			}
		}
	}
	return
}

var dateMap = map[string]string{
	"yyyy":       "2006",
	"yy":         "06",
	"mm":         "01",
	"dd":         "02",
	"yyyy-mm":    "2006-01",
	"yy-mm":      "06-01",
	"yyyy-mm-dd": "2006-01-02",
	"yy-mm-dd":   "06-01-02",
}

func (f *folderStruct) makeTargetName(filename string, match []string) string {
	if f.TargetName == "" {
		return filename
	}

	envsubst.SetPrefix('%')
	filename, err := envsubst.ConvertString(f.TargetName, func(name string) (string, bool) {

		if i, err := strconv.Atoi(name); err == nil {
			if i >= 0 && i < len(match) {
				return match[i], true
			}
		}

		switch strings.ToLower(name) {
		case "uid", "guid", "uuid":
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), true
		}

		if dateformat, ok := dateMap[name]; ok {
			return time.Now().Format(dateformat), true
		}

		for i, param := range f.varNames {
			if strings.EqualFold(param, name) {
				return match[i], true
			}
		}

		return "", false
	})

	if err != nil {
		log.Error().Msgf("error creating blob-name from '%s': %v", f.TargetName, err)
		return ""
	}

	return filename
}

func (f *folderStruct) handleFile(store storage.Storage, sender message.Sender, fi fs.FileInfo, dryRun bool) int {
	name := fi.Name()
	match := f.regexp.FindStringSubmatch(name)
	if match == nil {
		return 0
	}

	log.Info().Msgf("processing '%s'", name)

	msg := new(msgData)

	location := path.Join(f.Location, name)

	msg.AddVar("_size", fi.Size())
	msg.AddVar("_modified", fi.ModTime())
	msg.AddVar("_location", location)

	for k, v := range f.Tags {
		msg.AddVar(k, v)
	}

	for pos, value := range match {
		msg.AddVar(f.varNames[pos], value)
	}

	blobName := f.makeTargetName(name, match)
	if blobName == "" {
		return 0
	}

	// Save object to storage
	obj, err := store.Save(blobName, location, f.ContentType, fi.Size(), dryRun)
	if err != nil {
		log.Error().Msgf("save-error: %v", err)
		return 0
	}

	formats := make(map[string]*storage.StoredObject, 1)
	formats[f.Format] = obj

	msg.AddVar("externalData", formats)

	data := msg.Render()

	fmt.Println(string(data))

	// Send the message
	id, err := sender.Send(data, nil, dryRun)
	if err != nil {
		log.Error().Msgf("send-error: %v", err)
		return 0
	}
	if id == "" {
		log.Error().Msg("send-failed: no id")
		return 0
	}
	log.Info().Msgf("message sent with id: %s", id)

	// all done
	if f.deleteOnSuccess {
		if dryRun {
			log.Info().Msgf("dry-run: '%s' would be removed", name)
		} else {
			err = os.Remove(location)
			if err != nil {
				log.Error().Msgf("unable to delete '%s': %v", name, err)
			} else {
				log.Info().Msgf("'%s' removed", name)
			}
		}
	}

	return 1
}
