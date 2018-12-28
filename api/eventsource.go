package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/keydotcat/keycatd/util"
)

func (ah apiHandler) eventSourceRoot(w http.ResponseWriter, r *http.Request) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		return ah.eventSourceSubscribe(w, r)
	}
	return util.NewErrorFrom(ErrNotFound)
}

func (ah apiHandler) eventSourceFlusher(w http.ResponseWriter) http.Flusher {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return nil
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	return flusher
}

type eventSourceSender struct {
	w http.ResponseWriter
	f http.Flusher
}

func (e eventSourceSender) sendPing() error {
	if _, err := fmt.Fprintf(e.w, "id: %d\ndata: {\"action\": \"ping\"}\n\n", time.Now().UTC().UnixNano()); err != nil {
		return err
	}
	e.f.Flush()
	return nil
}

func (e eventSourceSender) sengMessage(msg []byte) error {
	if _, err := fmt.Fprintf(e.w, "id: %d\ndata: %s\n\n", time.Now().UTC().UnixNano(), string(msg)); err != nil {
		return err
	}
	e.f.Flush()
	return nil
}

// /eventsource
func (ah apiHandler) eventSourceSubscribe(w http.ResponseWriter, r *http.Request) error {
	flusher := ah.eventSourceFlusher(w)
	if flusher == nil {
		return nil
	}
	flusher.Flush()
	ess := eventSourceSender{w, flusher}
	return ah.broadcastEventListenLoop(r, ess.sendPing, ess.sengMessage)
}
