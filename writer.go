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
	started         bool
}

func NewWriter(dir string, id int, statsChannel chan time.Duration) *Writer {
	return &Writer{
		dir:             dir,
		id:              id,
		listenChannel:   make(chan struct{}, 1000),
		durationChannel: statsChannel,
		started:         false,
	}
}

func (w *Writer) run() {
	go func() {
		// random startup delay to spread the load
		time.Sleep((time.Duration)(rand.Intn(5000)) * time.Millisecond)

		w.started = true

		log.Printf("started writer %d", w.id)

		filename := filepath.Join(w.dir, fmt.Sprintf("%d", w.id))

		f, err := os.Create(filename)

		if err != nil {
			log.Printf("ERROR: writer %d failed to open its file %s", w.id, filename)
		}

		randomBuffer := make([]byte, 160)

		for i, _ := range randomBuffer {
			randomBuffer[i] = byte(rand.Intn(256))
		}

		// cleanup after us
		defer func() {
			f.Close()
			os.Remove(filename)
		}()

		writeBuffer := make([]byte, 0, 32768)

		for {
			select {
			case <-w.listenChannel:
				writeBuffer = append(writeBuffer, randomBuffer...)

				if len(writeBuffer) >= 32768 {

					start := time.Now()
					_, err := f.Write(writeBuffer)

					// truncate the write buffer
					writeBuffer = writeBuffer[:0]

					if err != nil {
						log.Printf("ERROR: writer %d failed to write frame", w.id)
					}

					took := time.Since(start)

					w.durationChannel <- took
				}
			}
		}
	}()
}

func (w *Writer) tick() {
	if w.started {
		w.listenChannel <- struct{}{}
	}
}
