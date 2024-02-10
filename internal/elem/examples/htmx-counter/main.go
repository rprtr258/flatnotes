package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/rprtr258/flatnotes/internal/elem"
	"github.com/rprtr258/flatnotes/internal/elem/attrs"
	"github.com/rprtr258/flatnotes/internal/elem/examples/htmx"
	"github.com/rprtr258/flatnotes/internal/elem/styles"
)

var counter int

func generateCounterContent() string {
	return elem.Render(elem.HTML(nil,
		elem.Head(nil,
			elem.Script(attrs.Props{attrs.Src: "https://unpkg.com/htmx.org@1.6.1"}),
		),
		elem.Body(nil,
			elem.Button(attrs.Props{
				htmx.HXPost:   "/increment",
				htmx.HXTarget: "#counter-div",
				htmx.HXSwap:   "innerText",
				attrs.Style: styles.Props{
					styles.BackgroundColor: "blue",
					styles.Color:           "white",
				}.ToInline(),
			}, elem.Text("Increment")),
			elem.Div(attrs.Props{attrs.ID: "counter-div"}, elem.Text(fmt.Sprintf("%d", counter))),
		),
	))
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content := generateCounterContent()
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	})
	http.HandleFunc("/increment", func(w http.ResponseWriter, r *http.Request) {
		counter++
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(strconv.Itoa(counter)))
	})

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err.Error())
	}
}
