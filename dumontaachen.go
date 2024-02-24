package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/arran4/golang-ical"
)

func init() {
	RegisterAdapter("dumontaachen", dumontaachenUrl, "Live jazz music bar in Aachen", dumontaachenAdapter)
}

type dumontaachenTextItem struct {
	Rendered string `json:"rendered"`
}

type dumontaachenEvent struct {
	Id       int64                `json:"id"`
	Date     string               `json:"date"`
	Link     string               `json:"link"`
	Modified string               `json:"modified"`
	Title    dumontaachenTextItem `json:"title"`
	Content  dumontaachenTextItem `json:"content"`
}

const (
	dumontaachenUrl         = "https://www.dumont-aachen.de/"
	dumontaachenSrc         = "https://dumont-aachen.de/wp-json/wp/v2/posts?per_page=50"
	dumontaachenLocalFormat = "2006-01-02T15:04:05"
)

var (
	dumontaachenTZ, _     = time.LoadLocation("Europe/Berlin")
	dumontaachenStripTags = regexp.MustCompile(`<.*?>`)
)

func (t dumontaachenTextItem) PlainText() string {
	return html.UnescapeString(dumontaachenStripTags.ReplaceAllString(t.Rendered, ""))
}

func (e dumontaachenEvent) ModifiedAt() time.Time {
	t, _ := time.ParseInLocation(dumontaachenLocalFormat, e.Modified, dumontaachenTZ)
	return t
}

// the pattern seems to be post date = event start time + 3h + a few seconds
func (e dumontaachenEvent) StartTime() time.Time {
	return e.EndTime().Add(-3 * time.Hour)
}

func (e dumontaachenEvent) EndTime() time.Time {
	t, _ := time.ParseInLocation(dumontaachenLocalFormat, e.Date, dumontaachenTZ)
	return t.Truncate(time.Minute)
}

func dumontaachenAdapter(w io.Writer) error {
	now := time.Now()

	req, err := http.NewRequest("GET", dumontaachenSrc, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "ics generator/0.1")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var events []dumontaachenEvent
	err = json.NewDecoder(res.Body).Decode(&events)
	if err != nil {
		return err
	}

	cal := ics.NewCalendarFor("github.com/cbix/ics")
	cal.SetName("Dumont Aachen")
	cal.SetColor("#fdc300")
	cal.SetUrl(dumontaachenUrl)
	cal.SetTimezoneId("Europe/Berlin")
	lastModified := time.Time{}
	for _, ev := range events {
		event := cal.AddEvent(fmt.Sprintf("event-%d@dumont-aachen.de", ev.Id))
		event.SetDtStampTime(now)
		event.SetSummary(ev.Title.PlainText())
		event.SetDescription(ev.Content.PlainText())
		event.SetLocation("Dumont Aachen")
		event.SetStartAt(ev.StartTime())
		event.SetEndAt(ev.EndTime())
		event.SetURL(ev.Link)
		modifiedAt := ev.ModifiedAt()
		event.SetCreatedTime(modifiedAt)
		event.SetModifiedAt(modifiedAt)
		if modifiedAt.After(lastModified) {
			lastModified = modifiedAt
		}
	}
	if !lastModified.IsZero() {
		cal.SetLastModified(lastModified)
	}
	return cal.SerializeTo(w)
}
