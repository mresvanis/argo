package main

import (
	"encoding/json"
)

type Event struct {
	Source *string                `json:"source,omitempty"`
	Line   uint64                 `json:"line,omitempty"`
	Offset int64                  `json:"offset,omitempty"`
	Text   map[string]interface{} `json:"text,omitempty"`
}

func NewEvent(source *string, line uint64, offset int64, text *string) Event {
	e := Event{}

	e.Source = source
	e.Line = line
	e.Offset = offset

	var jtext map[string]interface{}
	json.Unmarshal([]byte(*text), &jtext)
	e.Text = jtext

	return e
}

func (e *Event) GetLength() int64 {
	jtext, _ := json.Marshal(e.Text)
	return int64(len(string(jtext)))
}
