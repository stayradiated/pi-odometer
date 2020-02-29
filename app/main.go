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

var gasSwitchRise = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gas_switch_rise",
	Help: "Count of times the switch line has risen",
})

var gasSwitchFall = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "gas_switch_fall",
	Help: "Count of times the switch line has fallen",
})

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	prometheus.MustRegister(gasUsage)
	prometheus.MustRegister(gasSwitchRise)
	prometheus.MustRegister(gasSwitchFall)
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

	debounceGasChan := make(chan gpiod.LineEvent, 10)

	go debounce(1000*time.Millisecond, debounceGasChan, func(evt gpiod.LineEvent) {
		log.Println("gasUsage.Inc()")
		gasUsage.Inc()
	})

	line, err := c.RequestLine(
		rpi.GPIO26,
		gpiod.WithBothEdges(func(evt gpiod.LineEvent) {
			if evt.Type == gpiod.LineEventRisingEdge {
				log.Println("+++")
				gasSwitchRise.Inc()
				debounceGasChan <- evt
			} else if evt.Type == gpiod.LineEventFallingEdge {
				log.Println("---")
				gasSwitchFall.Inc()
			}
		}),
	)
	if err != nil {
		panic(err)
	}
	defer line.Close()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*addr, nil)
}
