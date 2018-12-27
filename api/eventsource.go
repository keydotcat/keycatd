package api

import (
	"fmt"
	"net/http"

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

// /eventsource
func (ah apiHandler) eventSourceSubscribe(w http.ResponseWriter, r *http.Request) error {
	flusher := ah.eventSourceFlusher(w)
	if flusher == nil {
		return nil
	}
	return ah.broadcastEventListenLoop(r, func() error {
		if _, err := fmt.Fprintf(w, "{ping: true}\r\n"); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}, func(msg []byte) error {
		if _, err := fmt.Fprintf(w, "%s\r\n", string(msg)); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	})
}
