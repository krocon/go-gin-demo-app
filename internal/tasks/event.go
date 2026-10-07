package tasks

// Event types, used in the "type" JSON field and as SSE event name.
const (
	EventCreated = "created"
	EventUpdated = "updated"
	EventDeleted = "deleted"
)

// Event is something that happened to a task. Sent to live clients as JSON, e.g.
//
//	{"type":"created","task":{"id":4,"title":"Try Gin","done":false}}
//	{"type":"deleted","id":4}
//
// Go has no sealed classes; one struct with a "type" discriminator and
// `omitempty` fields gives the same JSON shape.
type Event struct {
	Type string `json:"type"`
	Task *Task  `json:"task,omitempty"`
	ID   int64  `json:"id,omitempty"`
}

func Created(t Task) Event   { return Event{Type: EventCreated, Task: &t} }
func Updated(t Task) Event   { return Event{Type: EventUpdated, Task: &t} }
func Deleted(id int64) Event { return Event{Type: EventDeleted, ID: id} }
