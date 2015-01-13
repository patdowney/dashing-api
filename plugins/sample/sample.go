package main

import (
	"math/rand"
	"time"

	"github.com/patdowney/dashing-publisher/dashing"
)

type sample struct{}

func (j *sample) Work(send chan dashing.Event) {
	ticker := time.NewTicker(1 * time.Second)
	var lastValuation, lastKarma, currentValuation, currentKarma int
	for {
		select {
		case <-ticker.C:
			lastValuation, currentValuation = currentValuation, rand.Intn(100)
			lastKarma, currentKarma = currentKarma, rand.Intn(200000)
			send <- dashing.Event{
				WidgetID: "valuation",
				Body: map[string]interface{}{
					"current": currentValuation,
					"last":    lastValuation,
				}}
			send <- dashing.Event{
				WidgetID: "karma",
				Body: map[string]interface{}{
					"current": currentKarma,
					"last":    lastKarma,
				}}
			send <- dashing.Event{
				WidgetID: "synergy",
				Body: map[string]interface{}{
					"value": rand.Intn(100),
				}}
		}
	}
}

func main() {
	s := &sample{}

	dashing.StartPublishLoop(s)
}
