package main

import (
	"math/rand"
	"net/url"
	"time"

	"github.com/patdowney/dashing-publisher/dashing"
)

type Word struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type buzzwords struct {
	words []Word
}

func (j *buzzwords) Work(send chan dashing.Event) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			for i := 0; i < len(j.words); i++ {
				if 1 < rand.Intn(3) {
					value := j.words[i].Value
					j.words[i].Value = (value + 1) % 30
				}
			}
			words := make([]Word, 0)
			for _, v := range j.words {
				words = append(words, v)
			}
			send <- dashing.Event{
				WidgetID: "buzzwords",
				Body: map[string]interface{}{
					"items": words,
				}}
		}
	}
}

func main() {
	job := &buzzwords{[]Word{
		Word{"Paradigm shift", 0},
		Word{"Leverage", 0},
		Word{"Pivoting", 0},
		Word{"Turn-key", 0},
		Word{"Streamlininess", 0},
		Word{"Exit strategy", 0},
		Word{"Synergy", 0},
		Word{"Enterprise", 0},
		Word{"Web 2.0", 0},
	}}

	targetURL, _ := url.Parse("http://localhost:3000/widgets/buzzwords")

	p := dashing.NewJobPublisher(targetURL, job)
	p.Start()
}
