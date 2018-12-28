package api

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/keydotcat/keycatd/util"
)

var gEventSourceHeaders []byte

func init() {
	buf := util.BufPool.Get()
	defer util.BufPool.Put(buf)
	headers := []string{
		"HTTP/1.1 200 OK",
		"Content-Type: text/event-stream",
		"Vary: Accept-Encoding",
		"Cache-Control: no-cache",
		"Connection: keep-alive",
		"",
	}
	for _, header := range headers {
		buf.Write([]byte(fmt.Sprintf("%s\r\n", header)))
	}
	gEventSourceHeaders = append(gEventSourceHeaders, buf.Bytes()...)
}

func (ah apiHandler) eventSourceRoot(w http.ResponseWriter, r *http.Request) error {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	if len(head) == 0 {
		return ah.eventSourceSubscribe(w, r)
	}
	return util.NewErrorFrom(ErrNotFound)
}

func (ah apiHandler) makeEventSourceSender(w http.ResponseWriter) (*eventSourceSender, error) {
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return nil, util.NewErrorFrom(err)
	}
	_, err = conn.Write(gEventSourceHeaders)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &eventSourceSender{conn}, nil
}

type eventSourceSender struct {
	conn net.Conn
}

func (e eventSourceSender) sendPayload(msg string) error {
	payload := fmt.Sprintf("id: %d\nevent: message\ndata: %s\n\n", time.Now().UTC().UnixNano(), msg)
	e.conn.SetWriteDeadline(time.Now().Add(time.Second))
	_, err := e.conn.Write([]byte(payload))
	return err
}

func (e eventSourceSender) sendPing() error {
	return e.sendPayload("{\"action\": \"ping\"}")
}

func (e eventSourceSender) sengMessage(msg []byte) error {
	return e.sendPayload(string(msg))
}

// /eventsource
func (ah apiHandler) eventSourceSubscribe(w http.ResponseWriter, r *http.Request) error {
	ess, err := ah.makeEventSourceSender(w)
	if err != nil {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return nil
	}
	return ah.broadcastEventListenLoop(r, ess.sendPing, ess.sengMessage)
}
