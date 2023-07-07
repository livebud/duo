package hot

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/livebud/duo/internal/cli/pubsub"
)

// New server-sent event (SSE) server
func New(ps pubsub.Subscriber) *Server {
	return &Server{ps, time.Now}
}

type Server struct {
	ps  pubsub.Subscriber
	Now func() time.Time // Used for testing
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Take control of flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		err := fmt.Errorf("hot: response writer is not a flusher")
		http.Error(w, err.Error(), 500)
		return
	}
	// Set the appropriate response headers
	headers := w.Header()
	headers.Add(`Content-Type`, `text/event-stream`)
	headers.Add(`Cache-Control`, `no-cache`)
	headers.Add(`Connection`, `keep-alive`)
	headers.Add(`Access-Control-Allow-Origin`, "*")
	// Flush the headers
	flusher.Flush()
	// Subscribe to a specific page path or all pages
	subscription := s.ps.Subscribe("update", "create", "delete")
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-subscription.Wait():
			reload(flusher, w)
			flusher.Flush()
		}
	}
}

func reload(flusher http.Flusher, w http.ResponseWriter) {
	event := &Event{
		Data: []byte(`{"reload":true}`),
	}
	w.Write(event.Format().Bytes())
	flusher.Flush()
}

// https://html.spec.whatwg.org/multipage/server-sent-events.html#event-stream-interpretation
type Event struct {
	ID    string // id (optional)
	Type  string // event type (optional)
	Data  []byte // data
	Retry int    // retry (optional)
}

func (e *Event) Format() *bytes.Buffer {
	b := new(bytes.Buffer)
	if e.ID != "" {
		b.WriteString("id: " + e.ID + "\n")
	}
	if e.Type != "" {
		b.WriteString("event: " + e.Type + "\n")
	}
	if len(e.Data) > 0 {
		b.WriteString("data: ")
		b.Write(e.Data)
		b.WriteByte('\n')
	}
	if e.Retry > 0 {
		b.WriteString("retry: " + strconv.Itoa(e.Retry) + "\n")
	}
	b.WriteByte('\n')
	return b
}
