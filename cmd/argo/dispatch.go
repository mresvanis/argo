package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/mresvanis/argo/pkg/util"
)

type Dispatcher interface {
	// Setup initializes the dispatcher or returns an error if it fails.
	Setup() error

	// Send dispatches events and returns an acknowledgement or an error.
	Send([]Event) (Ack, error)
}

type Elasticsearch struct {
	hosts  []string
	client *elasticsearch.Client
	log    *log.Logger
}

func NewElasticsearchDispatcher(hosts []string) *Elasticsearch {
	es := new(Elasticsearch)

	es.hosts = hosts
	es.log = log.New(os.Stderr, fmt.Sprintf("[es] %s ", es.hosts), log.LstdFlags)

	return es
}

type bulkResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			ID     string `json:"_id"`
			Result string `json:"result"`
			Status int    `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
				Cause  struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"caused_by"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}

func (es *Elasticsearch) Setup() error {
	cfg := elasticsearch.Config{Addresses: es.hosts}

	var err error
	es.client, err = elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to setup client, %s", err)
	}

	return nil
}

func (es *Elasticsearch) Send(events []Event) (Ack, error) {
	if len(events) <= 0 {
		return Ack{}, fmt.Errorf("no events given")
	}

	numErrors := 0
	numIndexed := 0
	lastEvent := events[len(events)-1]

	start := time.Now().UTC()

	buf, err := prepareBulkPayload(events)
	if err != nil {
		return Ack{}, err
	}

	res, err := es.client.Bulk(bytes.NewReader(buf.Bytes()), es.client.Bulk.WithIndex(indexName))
	if err != nil {
		return Ack{}, fmt.Errorf("failed to index events with last offset %d, %s", lastEvent.Offset, err)
	}

	withErrors := false
	if !res.IsError() {
		var blk *bulkResponse
		if err := json.NewDecoder(res.Body).Decode(&blk); err != nil {
			es.log.Printf("failed to parse response body: %s", err.Error())

		} else {
			for _, d := range blk.Items {
				if d.Index.Status > 201 {
					numErrors++
					es.log.Printf("error: [%d]: %s: %s: %s: %s",
						d.Index.Status,
						d.Index.Error.Type,
						d.Index.Error.Reason,
						d.Index.Error.Cause.Type,
						d.Index.Error.Cause.Reason,
					)
					continue
				}

				numIndexed++
			}
		}
	} else {
		numErrors = len(events)
		withErrors = true

		var raw map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
			es.log.Printf("failed to parse response body, %s", err.Error())

		} else {
			es.log.Printf("error: [%d] %s: %s",
				res.StatusCode,
				raw["error"].(map[string]interface{})["type"],
				raw["error"].(map[string]interface{})["reason"],
			)
		}
	}

	dur := time.Since(start)
	es.log.Printf("indexed [%d] documents with [%d] errors in %s (%.0f docs/sec)",
		numIndexed,
		numErrors,
		dur.Truncate(time.Millisecond),
		1000.0/float64(dur/time.Millisecond)*float64(numIndexed),
	)

	return NewAck(lastEvent, withErrors), nil
}

func prepareBulkPayload(events []Event) (bytes.Buffer, error) {
	var buf bytes.Buffer
	for _, event := range events {
		meta := []byte(fmt.Sprintf(`{"index":{"_id":"%s"}}%s`, util.GenerateRandomString(32), "\n"))

		doc, err := json.Marshal(event)
		if err != nil {
			return bytes.Buffer{}, errors.New(fmt.Sprintf("cannot encode event %d, %s", event.Offset, err.Error()))
		}

		doc = append(doc, "\n"...)

		buf.Grow(len(meta) + len(doc))
		buf.Write(meta)
		buf.Write(doc)
	}
	return buf, nil
}
