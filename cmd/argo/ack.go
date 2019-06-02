package main

type Ack struct {
	event    Event
	hasError bool
}

func NewAck(event Event, hasError bool) Ack {
	return Ack{event, hasError}
}

func (ack Ack) Event() Event {
	return ack.event
}

func (ack Ack) HasError() bool {
	return ack.hasError
}
