package dashing

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
)

func GetEvents(w http.ResponseWriter, r *http.Request, e encoder.Encoder, b *Broker) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	c, ok := w.(http.CloseNotifier)
	if !ok {
		http.Error(w, "Close notification unsupported!", http.StatusInternalServerError)
		return
	}

	// Create a new channel, over which the broker can
	// send this client events.
	events := make(chan *Event)

	// Add this client to the map of those that should
	// receive updates
	b.newClients <- events

	// Remove this client from the map of attached clients
	// when the handler exits.
	defer func() {
		b.defunctClients <- events
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	closer := c.CloseNotify()

	for {
		select {
		case event := <-events:
			data := event.Body
			data["id"] = event.WidgetID
			data["updatedAt"] = int32(time.Now().Unix())
			if event.Target != "" {
				fmt.Fprintf(w, "event: %s\n", event.Target)
			}
			fmt.Fprintf(w, "data: %s\n\n", encoder.Must(e.Encode(data)))
			f.Flush()
		case <-closer:
			log.Println("Closing connection")
			return
		}
	}
}

func postEvent(r *http.Request, params martini.Params, b *Broker, target string) (int, string) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	var data map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return http.StatusBadRequest, ""
	}

	b.events <- &Event{
		WidgetID: params["id"],
		Body:     data,
		Target:   target}
	return http.StatusAccepted, ""
}

func PostDashboardEvent(r *http.Request, params martini.Params, b *Broker) (int, string) {
	return postEvent(r, params, b, "dashboards")
}

func PostWidgetEvent(r *http.Request, params martini.Params, b *Broker) (int, string) {
	return postEvent(r, params, b, "")
}
