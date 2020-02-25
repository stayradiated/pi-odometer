package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
			if evt.Type == gpiod.LineEventRisingEdge || evt.Type == gpiod.LineEventFallingEdge {
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

	// capture exit signals to ensure pin is reverted to input on exit.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	spammyChan := make(chan gpiod.LineEvent, 10)

	go debounce(1000*time.Millisecond, spammyChan, func(evt gpiod.LineEvent) {
		if evt.Type == gpiod.LineEventRisingEdge {
			log.Println("Incrementing gas usage")
			gasUsage.Inc()
		}
	})

	l2, err := c.RequestLine(
		rpi.GPIO2,
		gpiod.WithBothEdges(func(evt gpiod.LineEvent) {
			log.Println(evt.Type)
			spammyChan <- evt
		}),
	)
	if err != nil {
		panic(err)
	}
	defer l2.Close()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*addr, nil)
}
