package main

import (
	"encoding/json"
	"errors"
	"io"
	"time"
)

// Config holds the configuration values that argo needs in order to function.
type Config struct {
	DeadTime         string   `json:"dead_time"`
	Paths            []string `json:"paths"`
	Host             string   `json:"host"`
	Timeout          int64    `json:"timeout"`
	DispatchInterval int64    `json:"dispatch_interval"`
	BufferSize       int64    `json:"buffer_size"`

	deadtime         time.Duration
	timeout          time.Duration
	dispatchInterval time.Duration
}

// ParseConfig accepts a reader from which to parse the configuration, and returns a valid
// Config or an error.
func ParseConfig(r io.Reader) (*Config, error) {
	cfg := new(Config)

	dec := json.NewDecoder(r)
	err := dec.Decode(cfg)
	if err != nil {
		return nil, err
	}

	if len(cfg.Paths) <= 0 {
		return nil, errors.New("no paths defined")
	}

	if cfg.Host == "" {
		return nil, errors.New("host not defined")
	}

	if cfg.DeadTime == "" {
		cfg.DeadTime = "24h"
	}
	cfg.deadtime, err = time.ParseDuration(cfg.DeadTime)
	if err != nil {
		return nil, err
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 10
	}
	cfg.timeout = time.Duration(cfg.Timeout) * time.Second

	if cfg.DispatchInterval == 0 {
		cfg.DispatchInterval = 5
	}
	cfg.dispatchInterval = time.Duration(cfg.DispatchInterval) * time.Second

	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 2048
	}

	return cfg, nil
}
