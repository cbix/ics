package main

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

const (
	theateraachenApi = "https://www.theateraachen.de/de/spielplan/kalender.html?ajax=1&offset=%d"
)

func theateraachenAdapter(w io.Writer) error {
	for i := 0; i < 10; i++ {
		res, err := http.Get(fmt.Sprintf(theateraachenApi, i))
		if err != nil {
			return err
		}
		defer res.Body.Close()
	}
	return nil
}
