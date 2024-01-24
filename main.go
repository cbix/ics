package main

import (
	"io"
	"log"
	"net/http"
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
	log.Fatal(http.ListenAndServe(":8080", nil))
}
