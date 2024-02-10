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
	RegisterAdapter("sonicballroom", sonicballroomUrl, "Live music bar in Cologne", sonicballroomAdapter)
}

const (
	sonicballroomUrl      = "https://www.sonic-ballroom.de/"
	sonicballroomSrc      = "https://api.loveyourartist.com/v1/store/events?filter[profiles][$eq]=5ed76651ff8ee32c0749e103&filter[endAt][$gte]=%d"
	sonicballroomEventUrl = "https://loveyourartist.com/de/events/%s"
)

type sonicballroomEvent struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"publishedAt"`
	StartAt     time.Time `json:"startAt"`
	EndAt       time.Time `json:"endAt"`
	Slug        string    `json:"slug"`
}

func sonicballroomAdapter(w io.Writer) error {
	now := time.Now()
	res, err := http.Get(fmt.Sprintf(sonicballroomSrc, now.Truncate(240*time.Hour).UnixMilli()))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var events []sonicballroomEvent
	err = json.NewDecoder(res.Body).Decode(&events)
	if err != nil {
		return err
	}

	cal := ics.NewCalendarFor("github.com/cbix/ics")
	cal.SetName("Sonic Ballroom")
	cal.SetUrl(sonicballroomUrl)
	cal.SetTimezoneId("Europe/Berlin")
	lastModified := time.Time{}
	for _, ev := range events {
		event := cal.AddEvent(fmt.Sprintf("%s@loveyourartist.com", ev.Id))
		event.SetDtStampTime(now)
		event.SetSummary(ev.Name)
		event.SetLocation("Sonic Ballroom")
		event.SetStartAt(ev.StartAt)
		event.SetEndAt(ev.EndAt)
		event.SetURL("https://loveyourartist.com/de/events/" + ev.Slug)
		event.SetCreatedTime(ev.PublishedAt)
		event.SetModifiedAt(ev.PublishedAt)
		if ev.PublishedAt.After(lastModified) {
			lastModified = ev.PublishedAt
		}
	}
	if !lastModified.IsZero() {
		cal.SetLastModified(lastModified)
	}
	return cal.SerializeTo(w)
}
