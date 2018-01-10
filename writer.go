package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Writer struct {
	dir             string
	id              int
	listenChannel   chan struct{}
	durationChannel chan time.Duration
}

func NewWriter(dir string, id int, statsChannel chan time.Duration) *Writer {
	return &Writer{
		dir:             dir,
		id:              id,
		listenChannel:   make(chan struct{}, 1000),
		durationChannel: statsChannel,
	}
}

func (w *Writer) run() {
	go func() {
		filename := filepath.Join(w.dir, fmt.Sprintf("%d", w.id))

		f, err := os.Create(filename)

		if err != nil {
			log.Printf("ERROR: writer %d failed to open its file %s", w.id, filename)
		}

		buffer := make([]byte, 160)

		for i, _ := range buffer {
			buffer[i] = byte(rand.Intn(256))
		}

		// cleanup after us
		defer func() {
			f.Close()
			os.Remove(filename)
		}()

		for {
			select {
			case <-w.listenChannel:
				start := time.Now()
				_, err := f.Write(buffer)
				if err != nil {
					log.Printf("ERROR: writer %d failed to write frame", w.id)
				}
				took := time.Since(start)

				w.durationChannel <- took
			}
		}
	}()
}

func (w *Writer) tick() {
	w.listenChannel <- struct{}{}
}
