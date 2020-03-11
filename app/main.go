package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

var odometerGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: getEnv("GAUGE_NAME", "odometer"),
	Help: getEnv("GAUGE_DESCRIPTION", "Number of times the odometer has cycled."),
})

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	prometheus.MustRegister(odometerGauge)
}

const (
	LOW  = 1 << iota
	HIGH = 1 << iota
)

func main() {
	flag.Parse()

	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	pastState := LOW
	pastDate := time.Now()

	line, err := c.RequestLine(
		rpi.GPIO26,
		gpiod.WithBothEdges(func(evt gpiod.LineEvent) {
			date := time.Now()
			diff := date.Sub(pastDate)

			var state int

			if evt.Type == gpiod.LineEventRisingEdge {
				log.Println("+++")
				state = HIGH
			} else if evt.Type == gpiod.LineEventFallingEdge {
				log.Println("---")
				state = LOW
			}

			if diff.Milliseconds() > 500 && pastState == HIGH && state == LOW {
				log.Println("odometerGauge.Inc()")
				odometerGauge.Inc()
			}

			pastState = state
			pastDate = date
		}),
	)
	if err != nil {
		panic(err)
	}
	defer line.Close()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*addr, nil)
}
