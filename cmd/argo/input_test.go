package main

import (
	"sync"
	"testing"
)

func TestFileInputStart(t *testing.T) {
	path := "./testdata/test.log"
	out := make(chan []Event)
	ack := make(chan Ack)
	exp := []Event{{&path, 1, 0, map[string]interface{}{"test": "field"}}}

	input := NewFileInput(testcfg, path, testreg)

	go input.Start(out, ack)

	events, _ := <-out
	assertEq(events, exp, t)
}

func TestFileInputStop(t *testing.T) {
	path := "./testdata/test.log"
	out := make(chan []Event)
	ack := make(chan Ack)

	input := NewFileInput(testcfg, path, testreg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		input.Start(out, ack)
	}()
	input.Stop()
	wg.Wait()
}
