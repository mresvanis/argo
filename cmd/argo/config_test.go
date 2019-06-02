package main

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	var tests = []struct {
		json string
		cfg  *Config
		err  error
	}{
		{
			`{"host":"http://localhost:9200","paths":["./some.log"]}`,
			&Config{
				Host:             "http://localhost:9200",
				Paths:            []string{"./some.log"},
				DispatchInterval: 5,
				Timeout:          10,
				DeadTime:         "24h",
				BufferSize:       2048,
				dispatchInterval: time.Duration(5) * time.Second,
				timeout:          time.Duration(10) * time.Second,
				deadtime:         time.Duration(86400) * time.Second,
			},
			nil,
		}, {
			`{"host":"http://localhost:9200","paths":["./some.log"],"timeout":15}`,
			&Config{
				Host:             "http://localhost:9200",
				Paths:            []string{"./some.log"},
				DispatchInterval: 5,
				Timeout:          15,
				DeadTime:         "24h",
				BufferSize:       2048,
				dispatchInterval: time.Duration(5) * time.Second,
				timeout:          time.Duration(15) * time.Second,
				deadtime:         time.Duration(86400) * time.Second,
			},
			nil,
		}, {
			`{"host":"http://localhost:9200","paths":["./some.log"],"dead_time":"12h"}`,
			&Config{
				Host:             "http://localhost:9200",
				Paths:            []string{"./some.log"},
				DispatchInterval: 5,
				Timeout:          10,
				DeadTime:         "12h",
				BufferSize:       2048,
				dispatchInterval: time.Duration(5) * time.Second,
				timeout:          time.Duration(10) * time.Second,
				deadtime:         time.Duration(43200) * time.Second,
			},
			nil,
		}, {
			`{"host":"http://localhost:9200","paths":["./some.log"],"buffer_size":100}`,
			&Config{
				Host:             "http://localhost:9200",
				Paths:            []string{"./some.log"},
				DispatchInterval: 5,
				Timeout:          10,
				DeadTime:         "24h",
				BufferSize:       100,
				dispatchInterval: time.Duration(5) * time.Second,
				timeout:          time.Duration(10) * time.Second,
				deadtime:         time.Duration(86400) * time.Second,
			},
			nil,
		}, {
			`{"host":"http://localhost:9200","paths":["./some.log"],"dispatch_interval":4}`,
			&Config{
				Host:             "http://localhost:9200",
				Paths:            []string{"./some.log"},
				DispatchInterval: 4,
				Timeout:          10,
				DeadTime:         "24h",
				BufferSize:       2048,
				dispatchInterval: time.Duration(4) * time.Second,
				timeout:          time.Duration(10) * time.Second,
				deadtime:         time.Duration(86400) * time.Second,
			},
			nil,
		}, {
			`{"host":"http://localhost:9200"}`,
			nil,
			errors.New("no paths defined"),
		}, {
			`{"paths":["some.log"]}`,
			nil,
			errors.New("host not defined"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.json, func(t *testing.T) {
			config, err := ParseConfig(strings.NewReader(tt.json))

			assertEq(config, tt.cfg, t)
			assertEq(err, tt.err, t)
		})
	}
}
