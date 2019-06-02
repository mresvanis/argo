package main

import (
	"log"
	"os"
	"sync"
	"time"
)

const (
	indexName     = "argo"
	retryInterval = 5
)

// Output accepts event batches from inputs.
type Output interface {
	// Input returns the channel where it accepts incoming event
	// batches.
	Input() chan<- []Event

	// Start reads events from the input channel.
	Start()

	// Stop terminates the input loop.
	Stop()

	// Subscribe returns the channel to listen to acks to with the specific Event source.
	Subscribe(string) <-chan Ack

	// Unsubscribe removes the specified subscriber.
	Unsubscribe(string)
}

type EsOutput struct {
	sync.Mutex

	config *Config
	log    *log.Logger
	input  chan []Event
	term   chan struct{}
	es     *Elasticsearch

	subscribers map[string]chan Ack
}

func NewEsOutput(cfg *Config) Output {
	eso := new(EsOutput)

	eso.config = cfg
	eso.input = make(chan []Event, cfg.BufferSize)
	eso.term = make(chan struct{})
	eso.log = log.New(os.Stderr, "[out] ", log.LstdFlags)
	eso.subscribers = make(map[string]chan Ack)

	eso.es = NewElasticsearchDispatcher([]string{cfg.Host})

	return eso
}

func (eso *EsOutput) Input() chan<- []Event {
	return eso.input
}

func (eso *EsOutput) Start() {
	err := eso.es.Setup()
	if err != nil {
		eso.log.Printf("could not setup es client; %s", err.Error())
		return
	}

	for {
		select {
		case <-eso.term:
			eso.input = nil
			return

		case batch, ok := <-eso.input:
			if !ok {
				eso.input = nil
				// TODO: error handling
				return
			}

			for {
				ack, err := eso.es.Send(batch)
				if err == nil {
					eso.notifySubscribers(ack)
					break
				}

				eso.log.Printf(err.Error())
				time.Sleep(retryInterval * time.Second)
			}
		}
	}
}

func (eso *EsOutput) Stop() {
	eso.term <- struct{}{}
	eso.log.Printf("stopped for host %s", eso.config.Host)
}

func (eso *EsOutput) Subscribe(subID string) <-chan Ack {
	ackCh := make(chan Ack)

	eso.Lock()
	eso.subscribers[subID] = ackCh
	eso.Unlock()

	return ackCh
}

func (eso *EsOutput) Unsubscribe(subID string) {
	eso.Lock()
	delete(eso.subscribers, subID)
	eso.Unlock()
}

func (eso *EsOutput) notifySubscribers(ack Ack) {
	event := ack.Event()

	eso.Lock()
	ch, exists := eso.subscribers[*event.Source]
	if exists {
		select {
		case ch <- ack:
		default:
			eso.log.Printf("did not ack events for %s at offset %d", *event.Source, event.Offset)
		}
	}
	eso.Unlock()
}
