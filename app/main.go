package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

var gasUsage = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gas_usage",
	Help: "Units of gas used",
})

func init() {
	prometheus.MustRegister(gasUsage)
}

func debounce(interval time.Duration, input chan gpiod.LineEvent, cb func(evt gpiod.LineEvent)) {
	var evt gpiod.LineEvent
	timer := time.NewTimer(interval)
	for {
		select {
		case evt = <-input:
			timer.Reset(interval)
		case <-timer.C:
			if evt.Type > 0 {
				cb(evt)
			}
		}
	}
}

func main() {
	flag.Parse()

	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	spammyChan := make(chan gpiod.LineEvent, 10)

	go debounce(500*time.Millisecond, spammyChan, func(evt gpiod.LineEvent) {
		log.Println("gasUsage.Inc()")
		gasUsage.Inc()
	})

	l2, err := c.RequestLine(
		rpi.GPIO26,
		gpiod.WithBothEdges(func(evt gpiod.LineEvent) {
			if evt.Type == gpiod.LineEventRisingEdge {
				log.Println("+++")
				spammyChan <- evt
			} else if evt.Type == gpiod.LineEventFallingEdge {
				log.Println("---")
			}
		}),
	)
	if err != nil {
		panic(err)
	}
	defer l2.Close()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*addr, nil)
}
