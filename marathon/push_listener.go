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
	EventType  string    `json:"eventType"`
	Timestamp  time.Time `json:"timestamp"`
	SlaveID    string    `json:"slaveId"`
	TaskID     string    `json:"taskId"`
	TaskStatus string    `json:"taskStatus"`
	AppID      string    `json:"appId"`
	Host       string    `json:"host"`
	Ports      []int     `json:"ports"`
	Version    string    `json:"version"`
}

type Listener struct {
	events       chan Event
	host         string
	internalPort string // Internal/external ports are relative
	externalPort string // to the container this process runs in.
}

func NewListener(host string, internalPort, externalPort string) *Listener {
	listener := &Listener{
		events:       make(chan Event),
		host:         host,
		internalPort: internalPort,
		externalPort: externalPort,
	}
	http.HandleFunc("/push-listener", listener.handler)
	go http.ListenAndServe(":"+internalPort, nil)

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

func (l *Listener) unsubscribe(marathonHost, callback string) error {
	marathonURL := url.URL{Scheme: "http", Host: marathonHost, Path: "/v2/eventSubscriptions"}
	q := marathonURL.Query()
	q.Set("callbackUrl", callback)
	marathonURL.RawQuery = q.Encode()

	req, err := http.NewRequest("DELETE", marathonURL.String(), nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errmsg := "Bad status code while unsubscribing marathon event url: %s -- %s"
		return fmt.Errorf(errmsg, res.Status, callback)
	}

	return nil
}

func (l *Listener) unsubscribeStatCallbacks(marathonHost string) error {
	marathonURL := url.URL{Scheme: "http", Host: marathonHost, Path: "/v2/eventSubscriptions"}

	res, err := http.Get(marathonURL.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad status code while unsubscribing marathon event urls: %s", res.Status)
	}

	var data struct {
		Callbacks []string `json:"callbackUrls"`
	}
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	for _, callback := range data.Callbacks {
		if !strings.Contains(callback, "/ms-push-listener") { // Ignore non-marathon-stat urls
			continue
		}

		if err = l.unsubscribe(marathonHost, callback); err != nil {
			return err
		}
	}

	return nil
}

func (l *Listener) Subscribe(marathonHost string) error {
	if err := l.unsubscribeStatCallbacks(marathonHost); err != nil {
		return err
	}

	marathonURL := url.URL{Scheme: "http", Host: marathonHost, Path: "/v2/eventSubscriptions"}
	q := marathonURL.Query()
	q.Set("callbackUrl", fmt.Sprintf("http://%s:%s/ms-push-listener", l.host, l.externalPort))
	marathonURL.RawQuery = q.Encode()

	res, err := http.Post(marathonURL.String(), "application/json", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad status code while subscribing to marathon events: %s", res.Status)
	}

	var data map[string]interface{}
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&data); err != nil {
		return err
	}


	return nil
}
