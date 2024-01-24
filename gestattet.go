package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/arran4/golang-ical"
)

func init() {
	RegisterAdapter("gestattet", gestattetUrl, "Underground concerts in Aachen", gestattetAdapter)
}

type gestattetLocation struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type gestattetEvent struct {
	Id          int64             `json:"id"`
	Date        string            `json:"date"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Image       string            `json:"image"`
	Location    gestattetLocation `json:"location"`
	CreatedAt   time.Time         `json:"date_created"`
	UpdatedAt   time.Time         `json:"date_updated"`
}

const (
	gestattetUrl         = "https://gestattet.es/"
	gestattetApi         = "https://api.gestattet.es/items/Events/?fields=id,date,title,description,image,location.id,location.name,date_created,date_updated&limit=-1"
	gestattetRover       = 2
	gestattetLola        = 4
	gestattetLocalFormat = "2006-01-02T15:04:05"
)

var gestattetTZ, _ = time.LoadLocation("Europe/Berlin")

func (e gestattetEvent) StartTime() time.Time {
	t, _ := time.ParseInLocation(gestattetLocalFormat, e.Date, gestattetTZ)
	return t
}

func (e gestattetEvent) EndTime() time.Time {
	t := e.StartTime()
	if e.Location.Id == gestattetLola || e.Location.Id == gestattetRover {
		// these always end at 22:00
		end := time.Date(t.Year(), t.Month(), t.Day(), 22, 0, 0, 0, t.Location())
		if end.After(t) {
			return end
		}
	}
	return t.Add(3 * time.Hour)
}

type gestattetData struct {
	Data []gestattetEvent `json:"data"`
}

func gestattetAdapter(w io.Writer) error {
	res, err := http.Get(gestattetApi)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body := new(gestattetData)
	err = json.NewDecoder(res.Body).Decode(body)
	if err != nil {
		return err
	}

	cal := ics.NewCalendarFor("github.com/cbix/ics")
	cal.SetName("gestattet.es")
	cal.SetUrl(gestattetUrl)
	cal.SetTimezoneId("Europe/Berlin")
	lastModified := time.Time{}
	now := time.Now()
	for _, ev := range body.Data {
		event := cal.AddEvent(fmt.Sprintf("event-%d@gestattet.es", ev.Id))
		event.SetDtStampTime(now)
		event.SetSummary(ev.Title)
		event.SetDescription(ev.Description)
		event.SetLocation(ev.Location.Name)
		event.SetStartAt(ev.StartTime())
		event.SetEndAt(ev.EndTime())
		event.SetURL("https://api.gestattet.es/assets/" + ev.Image)
		if !ev.CreatedAt.IsZero() {
			event.SetCreatedTime(ev.CreatedAt)
			if ev.CreatedAt.After(lastModified) {
				lastModified = ev.CreatedAt
			}
		}
		if !ev.UpdatedAt.IsZero() {
			event.SetModifiedAt(ev.UpdatedAt)
			if ev.UpdatedAt.After(lastModified) {
				lastModified = ev.UpdatedAt
			}
		}
	}
	if !lastModified.IsZero() {
		cal.SetLastModified(lastModified)
	}
	return cal.SerializeTo(w)
}
