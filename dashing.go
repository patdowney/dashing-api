package dashing

import (
	"net/http"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
)

type dashingServer struct {
	martini *martini.Martini
	broker  *Broker

	staticDirectory string
}

func NewServer(staticDirectory string) *dashingServer {
	s := dashingServer{
		martini:         martini.New(),
		broker:          NewBroker(),
		staticDirectory: staticDirectory,
	}

	s.initDashing()

	return &s
}

func (s *dashingServer) Start() {
	// Start the event broker
	s.broker.Start()

	// Start the jobs
	for _, j := range registry {
		go j.Work(s.broker.events)
	}

	// Start Martini
	s.martini.Run()
}

func (s *dashingServer) initDashing() {
	// Setup middleware
	s.martini.Use(martini.Recovery())
	s.martini.Use(martini.Logger())
	s.martini.Use(martini.Static(s.staticDirectory))

	// Setup encoder
	s.martini.Use(func(c martini.Context, w http.ResponseWriter) {
		c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	})

	// Setup and inject event broker
	s.martini.Map(s.broker)

	// Setup routes
	r := martini.NewRouter()

	r.Get("/", GetRoot)
	r.Get("/events", GetEvents)
	r.Get("/:dashboard", GetDashboard)
	r.Post("/dashboards/:id", PostDashboardEvent)
	r.Post("/widgets/:id", PostWidgetEvent)
	r.Get("/views/:widget.html", GetWidget)

	// Add the router action
	s.martini.Action(r.Handle)
}

// Start all jobs and listen to requests.
func Start() {
	StartWithStaticDirectory("public")
}

func StartWithStaticDirectory(staticDirectory string) {
	server := NewServer(staticDirectory)

	server.Start()
}
