package main

import (
	"math/rand"
	"net/url"
	"time"

	"github.com/patdowney/dashing-publisher/dashing"
)

type convergence struct {
	points []map[string]int
}

func (j *convergence) Work(send chan dashing.Event) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			j.points = j.points[1:]
			j.points = append(j.points, map[string]int{
				"x": j.points[len(j.points)-1]["x"] + 1,
				"y": rand.Intn(50),
			})
			send <- dashing.Event{
				WidgetID: "convergence",
				Body: map[string]interface{}{
					"points": j.points,
				},
			}
		}
	}
}

func NewRandomConvergence() *convergence {
	c := &convergence{}
	//points: make([]map[string]int, 10, 10)}
	for i := 0; i < 10; i++ {
		c.points = append(c.points, map[string]int{
			"x": i,
			"y": rand.Intn(50),
		})
	}
	return c
}

func main() {
	job := NewRandomConvergence()

	//dashing.StartPublishLoop(c)

	targetURL, _ := url.Parse("http://localhost:3000/widgets/convergence")

	p := dashing.NewJobPublisher(targetURL, job)
	p.Start()

}
