package dashing

// An Event contains the widget ID, a body of data,
// and an optional target (only "dashboard" for now).
type Event struct {
	WidgetID string
	Body     map[string]interface{}
	Target   string
}

// An eventCache stores the latest event for each key, so that new clients can
// catch up.
type eventCache map[string]*Event
