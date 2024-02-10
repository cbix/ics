package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

func init() {
	RegisterAdapter("theateraachen", theateraachenUrl, "Theatre in Aachen (beta)", theateraachenAdapter)
}

type theateraachenData struct {
	Ecommerce theateraachenEcommerce `json:"ecommerce"`
}

type theateraachenEcommerce struct {
	Impressions []theateraachenImpression `json:"impressions"`
}

type theateraachenImpression struct {
	Id       string               `json:"id"`
	Name     string               `json:"name"`
	Variant  theateraachenVariant `json:"variant"`
	Position int                  `json:"position"`
}

type theateraachenVariant struct {
	DateTime  string `json:"dateTime"`
	VenueName string `json:"venueName"`
	GenreName string `json:"genreName"`
}

func (i theateraachenImpression) StartTime() time.Time {
	t, _ := time.ParseInLocation(theateraachenLocalFormat, i.Variant.DateTime, theateraachenTZ)
	return t
}

func (i theateraachenImpression) EndTime() time.Time {
	return i.StartTime().Add(2 * time.Hour)
}

const (
	theateraachenUrl         = "https://www.theateraachen.de/"
	theateraachenSrc         = "https://theateraachen.reservix.de/events/%d"
	theateraachenJsonPrefix  = "var impressionData = "
	theateraachenLocalFormat = "2006-01-02 15:04:05"
)

var theateraachenTZ, _ = time.LoadLocation("Europe/Berlin")

func theateraachenAdapter(w io.Writer) error {
	cal := ics.NewCalendarFor("github.com/cbix/ics")
	cal.SetName("Theater Aachen")
	cal.SetUrl(theateraachenUrl)
	cal.SetTimezoneId("Europe/Berlin")
	now := time.Now()
	pos := 0
page:
	for i := 1; i < 10; i++ {
		req, err := http.NewRequest("GET", fmt.Sprintf(theateraachenSrc, i), nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 Gecko/20100101 Firefox/122.0")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(res.Body)
		var jsonData []byte
		for scanner.Scan() {
			l := strings.Trim(scanner.Text(), " ;")
			if strings.HasPrefix(l, theateraachenJsonPrefix) {
				jsonData = []byte(strings.TrimPrefix(l, theateraachenJsonPrefix))
				break
			}
		}
		if len(jsonData) == 0 {
			continue
		}
		var data theateraachenData
		err = json.Unmarshal(jsonData, &data)
		if err != nil {
			return err
		}
		if len(data.Ecommerce.Impressions) == 0 {
			break
		}
		for _, ev := range data.Ecommerce.Impressions {
			p := ev.Position
			if p <= pos {
				break page
			}
			pos = p
			event := cal.AddEvent(fmt.Sprintf("%s@theateraachen.reservix.de", ev.Id))
			event.SetDtStampTime(now)
			event.SetSummary(ev.Name)
			event.SetDescription(ev.Variant.GenreName)
			event.SetLocation(ev.Variant.VenueName)
			event.SetStartAt(ev.StartTime())
			event.SetEndAt(ev.EndTime())
		}
	}
	return cal.SerializeTo(w)
}
