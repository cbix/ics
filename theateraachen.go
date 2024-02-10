package main

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

func init() {
	RegisterAdapter("theateraachen", theateraachenUrl, "Theatre in Aachen", theateraachenAdapter)
}

const (
	theateraachenUrl = "https://www.theateraachen.de/"
	//theateraachenSrc = "https://www.theateraachen.de/de/spielplan/kalender.html?ajax=1&offset=%d"
	theateraachenSrc = "https://theateraachen.reservix.de/events/%d"
)

func theateraachenAdapter(w io.Writer) error {
	n := 0
	for i := 1; i < 10; i++ {
		res, err := http.Get(fmt.Sprintf(theateraachenSrc, i))
		if err != nil {
			return err
		}
		nodes, err := html.Parse(res.Body)
		res.Body.Close()
		if err != nil {
			return err
		}
		startNode := nodes.FirstChild.FirstChild.NextSibling.FirstChild
		if startNode.FirstChild == nil {
			break
		}
		//fmt.Fprintf(w, "startNode: %+v\n", startNode)
		for node := startNode; node != nil; node = node.NextSibling {
			if node.FirstChild == nil || node.FirstChild.NextSibling == nil {
				continue
			}
			n++
			fmt.Fprintf(w, "child: %+v\n", node.FirstChild.NextSibling)
		}
	}
	fmt.Fprintf(w, "nodes: %d\n", n)
	return nil
}
