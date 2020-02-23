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

func debounce(interval time.Duration, input chan gpiod.LineEvent, f func(evt gpiod.LineEvent)) {
	var (
		evt gpiod.LineEvent
	)
	for {
		select {
		case evt = <-input:
			log.Println("received event")
		case <-time.After(interval):
			f(evt)
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
		gasUsage.Inc()
		if evt.Type == gpiod.LineEventRisingEdge {
			log.Println("RisingEdge")
		} else if evt.Type == gpiod.LineEventFallingEdge {
			log.Println("FallingEdge")
		}
	})

	l2, err := c.RequestLine(
		rpi.GPIO26,
		gpiod.WithBothEdges(func(evt gpiod.LineEvent) { spammyChan <- evt }),
	)
	if err != nil {
		panic(err)
	}
	defer l2.Close()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*addr, nil)
}
