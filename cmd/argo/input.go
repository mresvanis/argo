package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"time"

	"golang.org/x/xerrors"

	"github.com/mresvanis/argo/pkg/registry"
	"github.com/mresvanis/argo/pkg/util"
)

const (
	batchSize = 128
)

// Input starts watching the specified file.
type Input interface {
	// Start spawns a read loop and outputs batches of Events while waiting for acks.
	Start(chan<- []Event, <-chan Ack)

	// Stop terminates the read loop.
	Stop()
}

type FileInput struct {
	config *Config
	reg    registry.Registrar

	path   string
	offset int64

	file *os.File
	log  *log.Logger
	term chan struct{}

	events []Event

	reader       *bufio.Reader
	buffer       *bytes.Buffer
	lastSendTime time.Time
	lastReadTime time.Time
	readTimeout  time.Duration
}

func NewFileInput(cfg *Config, path string, reg registry.Registrar) Input {
	fi := new(FileInput)

	fi.config = cfg
	fi.path = path
	fi.reg = reg

	fi.log = log.New(os.Stderr, "[file] ", log.LstdFlags)
	fi.term = make(chan struct{})

	return fi
}

func (fi *FileInput) Start(output chan<- []Event, ack <-chan Ack) {
	var line uint64 = 0

	err := fi.setup()
	if err != nil {
		fi.log.Printf("%s; %s", fi.path, err.Error())
		return
	}
	defer fi.file.Close()

	for {
		if len(fi.events) > 0 && fi.shouldDispatch() {
			fi.dispatch(output, ack)
		}
		if fi.stopped() {
			break
		}

		text, bytesread, err := util.Readline(fi.reader, fi.buffer, fi.readTimeout)
		if xerrors.Is(err, io.EOF) {
			if fi.isFileDead() {
				fi.log.Printf("stopped watching dead file %s", fi.path)
				break
			}

			if util.IsFileTruncated(fi.file, fi.offset) {
				fi.log.Printf("file %s truncated, seeking from start", fi.path)
				fi.resetFileOffset()
				line = 0
			}

			continue
		}
		if err != nil {
			fi.log.Printf("unexpected state reading from %s, %s", fi.path, err)
			break
		}

		fi.lastReadTime = time.Now()
		line++
		fi.events = append(fi.events, NewEvent(&fi.path, line, fi.offset, text))
		fi.offset += int64(bytesread)

		if len(fi.events) >= batchSize || fi.shouldDispatch() {
			fi.dispatch(output, ack)
		}
	}

	fi.log.Printf("terminated input for %s", fi.path)
}

func (fi *FileInput) Stop() {
	fi.log.Printf("terminating input for %s", fi.path)
	fi.term <- struct{}{}
}

func (fi *FileInput) dispatch(output chan<- []Event, ack <-chan Ack) {
	output <- fi.events

	err := fi.waitForAck(ack)
	if xerrors.Is(err, registry.ErrUpdate) {
		fi.log.Printf("%s; %s", fi.path, err.Error())

	} else if err != nil {
		fi.log.Printf("%s; %s", fi.path, err.Error())
		fi.lastSendTime = time.Now()
		return
	}

	fi.events = []Event{}
	fi.lastSendTime = time.Now()
}

func (fi *FileInput) isFileDead() bool {
	return time.Since(fi.lastReadTime) >= fi.config.deadtime
}

func (fi *FileInput) resetFileOffset() {
	fi.file.Seek(0, os.SEEK_SET)
	fi.offset = 0
}

func (fi *FileInput) setFileOffset() {
	offset, _ := fi.file.Seek(0, os.SEEK_CUR)
	fi.offset, _ = fi.reg.GetOffset(fi.path)

	if fi.offset > 0 {
		fi.log.Printf("%s position:%d (offset snapshot:%d)", fi.path, fi.offset, offset)
		fi.file.Seek(fi.offset, os.SEEK_SET)
		return
	}

	fi.log.Printf("%s (offset snapshot:%d)", fi.path, offset)
	fi.file.Seek(0, os.SEEK_SET)
	fi.offset = 0
}

func (fi *FileInput) setup() error {
	file, err := os.Open(fi.path)
	if err != nil {
		return xerrors.Errorf("could not open file: %w", err)
	}
	fi.file = file

	fi.setFileOffset()

	_, err = fi.file.Stat()
	if err != nil {
		return xerrors.Errorf("cound not stat file: %w", err)
	}

	fi.reader = bufio.NewReaderSize(fi.file, 16<<10) // 16kb buffer by default
	fi.buffer = new(bytes.Buffer)
	fi.lastReadTime = time.Now()
	fi.lastSendTime = fi.lastReadTime
	fi.readTimeout = 10 * time.Second

	return nil
}

func (fi *FileInput) shouldDispatch() bool {
	return time.Now().Sub(fi.lastSendTime) >= fi.config.dispatchInterval
}

func (fi *FileInput) stopped() bool {
	select {
	case <-fi.term:
		return true
	default:
	}
	return false
}

func (fi *FileInput) waitForAck(ackCh <-chan Ack) error {
	select {
	case ack, ok := <-ackCh:
		// TODO: ERROR HANDLING
		if !ok {
			return xerrors.Errorf("ack channel error")
		}
		e := ack.Event()
		fi.log.Printf("%s; ack received for offset %d and hasErrors: %t", *e.Source, e.Offset, ack.HasError())

		if ack.HasError() {
			return xerrors.Errorf("%s; could not dispatch batch with offset %d", *e.Source, e.Offset)
		}

		err := fi.reg.UpdateOffset(*e.Source, e.Offset+e.GetLength()+1)
		if err != nil {
			return xerrors.Errorf("could not update registry for offset %d: %w", e.Offset, err)
		}

		return nil
	}
}
