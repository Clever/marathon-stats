package marathon

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Event struct {
	EventType  string    `json:eventType`
	Timestamp  time.Time `json:timestamp`
	SlaveID    string    `json:slaveId`
	TaskID     string    `json:taskId`
	TaskStatus string    `json:taskStatus`
	AppID      string    `json:appId`
	Host       string    `json:host`
	Ports      []int     `json:ports`
	Version    string    `json:version`
}

type Listener struct {
	events chan Event
	host   string
	port   string
}

func NewListener(host string, port string) *Listener {
	listener := &Listener{
		events: make(chan Event),
		host:   host,
		port:   port,
	}
	http.HandleFunc("/push-listener", listener.handler)
	go http.ListenAndServe(":"+port, nil)

	return listener
}

func (l *Listener) handler(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)

	var event Event
	if err := decoder.Decode(&event); err != nil {
		log.Fatal(err)
	}

	if event.EventType == "status_update_event" { // We only care about container change events
		l.events <- event
	}

	res.Write([]byte("Thanks.")) // Marathon ignores replies.  Just being polite.
}

func (l *Listener) Events() <-chan Event {
	return l.events
}

func (l *Listener) Subscribe(marathonHost string) error {
	marathonURL := url.URL{Scheme: "http", Host: marathonHost, Path: "/v2/eventSubscriptions"}
	q := marathonURL.Query()
	q.Set("callbackUrl", fmt.Sprintf("http://%s:%s/push-listener", l.host, l.port))
	marathonURL.RawQuery = q.Encode()

	res, err := http.Post(marathonURL.String(), "application/json", strings.NewReader(""))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var data map[string]interface{}
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad status code while subscribing to marathon events: " + res.Status)
	}

	return nil
}
