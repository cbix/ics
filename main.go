package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/coreos/go-systemd/v22/activation"
)

type icsAdapter struct {
	name          string
	url           string
	description   string
	writeCalendar func(w io.Writer) error
}

func (a icsAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/calendar")
	if err := a.writeCalendar(w); err != nil {
		log.Printf("Error writing calendar %s: %v", a.name, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *icsAdapter) Path() string {
	return "/" + a.name + ".ics"
}

var adapters []icsAdapter

func RegisterAdapter(name, url, description string, writeCalendar func(w io.Writer) error) {
	adapters = append(adapters, icsAdapter{name, url, description, writeCalendar})
}

func main() {
	for _, adapter := range adapters {
		http.Handle(adapter.Path(), adapter)
	}

	listeners, err := activation.Listeners()
	if err == nil && len(listeners) >= 1 {
		wg := new(sync.WaitGroup)
		wg.Add(len(listeners))
		for _, l := range listeners {
			go func(listener net.Listener) {
				log.Println("Listening on systemd activated socket ...")
				log.Fatal(http.Serve(listener, nil))
				wg.Done()
			}(l)
		}
		wg.Wait()
	} else {
		log.Fatal(http.ListenAndServe(":8083", nil))
	}
}
