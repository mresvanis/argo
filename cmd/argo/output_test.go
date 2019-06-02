package main

import (
	"testing"
)

func TestEsOutputInput(t *testing.T) {
	eso := EsOutput{input: make(chan []Event, 2)}

	input := eso.Input()

	input <- []Event{{Line: 1}}

	assertEq(len(input), 1, t)
	assertEq(cap(input), 2, t)
}
