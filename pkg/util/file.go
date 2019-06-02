package util

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"time"

	"golang.org/x/xerrors"
)

func IsFileTruncated(file *os.File, offset int64) bool {
	info, _ := file.Stat()
	return info.Size() < offset
}

func Readline(r *bufio.Reader, b *bytes.Buffer, t time.Duration) (*string, int, error) {
	var isPartial bool = true
	var newlineLength int = 1

	startTime := time.Now()

	for {
		segment, err := r.ReadBytes('\n')

		if segment != nil && len(segment) > 0 {
			if segment[len(segment)-1] == '\n' {
				isPartial = false

				if len(segment) > 1 && segment[len(segment)-2] == '\r' {
					newlineLength++
				}
			}

			b.Write(segment)
		}

		if err != nil {
			if err == io.EOF && isPartial {
				time.Sleep(1 * time.Second)

				if time.Since(startTime) > t {
					return nil, 0, xerrors.Errorf("read timeout: %w", err)
				}
				continue
			}

			return nil, 0, xerrors.Errorf("read error: %w", err)
		}

		if !isPartial {
			bufferSize := b.Len()
			str := new(string)
			*str = b.String()[:bufferSize-newlineLength]
			b.Reset()
			return str, bufferSize, nil
		}
	}
}
