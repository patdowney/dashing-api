package dashing

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-martini/martini"
	"github.com/karlseguin/gerb"
	"github.com/martini-contrib/encoder"
)

func GetRoot(w http.ResponseWriter, r *http.Request) {
	files, _ := filepath.Glob("dashboards/*.gerb")

	for _, file := range files {
		dashboard := file[11 : len(file)-5]
		if dashboard != "layout" {
			http.Redirect(w, r, "/"+dashboard, http.StatusTemporaryRedirect)
			return
		}
	}

	http.NotFound(w, r)
	return
}

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

func GetDashboard(r *http.Request, w http.ResponseWriter, params martini.Params) {
	template, err := gerb.ParseFile(true, "dashboards/"+params["dashboard"]+".gerb", "dashboards/layout.gerb")

	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	template.Render(w, map[string]interface{}{
		"dashboard":   params["dashboard"],
		"development": os.Getenv("DEV") != "",
		"request":     r,
	})
	return
}

func postEvent(r *http.Request, params martini.Params, b *Broker, target string) (int, string) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	var data map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return http.StatusBadRequest, ""
	}

	b.events <- &Event{params["id"], data, target}
	return http.StatusNoContent, ""
}

func PostDashboardEvent(r *http.Request, params martini.Params, b *Broker) (int, string) {
	return postEvent(r, params, b, "dashboards")
}

func PostWidgetEvent(r *http.Request, params martini.Params, b *Broker) (int, string) {
	return postEvent(r, params, b, "")
}

func GetWidget(w http.ResponseWriter, r *http.Request, params martini.Params) {
	template, err := gerb.ParseFile(true, "widgets/"+params["widget"]+"/"+params["widget"]+".html")

	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	template.Render(w, nil)
	return
}
