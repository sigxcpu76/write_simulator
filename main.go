package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var writersCount = 100
var writeFolder = "."

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	flag.IntVar(&writersCount, "writers", 1, "number of simultaneous writers")
	flag.StringVar(&writeFolder, "dir", ".", "write directory")

	flag.Parse()

	log.Printf("using %d writers in %s", writersCount, writeFolder)

	// create the channels
	statsChannel := make(chan time.Duration, 10000)

	// creating the writers
	writers := make([]*Writer, writersCount)
	for i := 0; i < writersCount; i++ {
		writers[i] = NewWriter(writeFolder, i, statsChannel)
		writers[i].run()
	}

	// start a 20ms ticker
	ticker := time.NewTicker(time.Millisecond * 20)
	go func() {
		for range ticker.C {
			for _, w := range writers {
				w.tick()
			}
		}
	}()

	// start a one second statistics display
	displayChannel := make(chan struct{})
	ticker2 := time.NewTicker(time.Second * 1)
	go func() {
		for range ticker2.C {
			displayChannel <- struct{}{}
		}
	}()

	values := make([]time.Duration, 1000)

	for {
		select {
		case v := <-statsChannel:
			values = append(values, v)
		case <-displayChannel:
			if len(values) == 0 {
				log.Printf("got no values this time")
			} else {
				minTime := values[0]
				maxTime := values[0]
				avg := 0

				for _, v := range values {
					if v < minTime {
						minTime = v
					} else if v > maxTime {
						maxTime = v
					}
					avg += int(v / time.Microsecond)
				}

				avg /= len(values)
				values = values[:0]

				log.Printf("min time: %v max time: %v avg time: %d µs", minTime, maxTime, avg)
			}
		}
	}

}
