package main

import (
	"github.com/AB-Lindex/filem/src/metrics"
	"github.com/AB-Lindex/filem/src/metrics/prompush"
)

type metricsConfig struct {
	PromPush *prompush.Settings `yaml:"prompush"`
}

type metricsFallback struct{}

func (mf *metricsFallback) Set(name string, value interface{}, keys map[string]interface{}) error {
	return nil
}

func (mf *metricsFallback) Send() error {
	return nil
}

var fallback *metricsFallback

func (mc *metricsConfig) Connect(dryRun bool) (metrics.Target, error) {
	if mc.PromPush != nil {
		return mc.PromPush.GetHandler()
	}
	// if no metrics is defined, return a fallback "dud"
	return fallback, nil
}
