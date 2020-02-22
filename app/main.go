package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	l2, err := c.RequestLine(
		rpi.GPIO2,
		gpiod.WithRisingEdge(func(evt gpiod.LineEvent) { gasUsage.Inc() }),
	)
	if err != nil {
		panic(err)
	}
	defer l2.Close()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*addr, nil)
}
