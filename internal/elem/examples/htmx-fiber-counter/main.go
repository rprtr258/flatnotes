package main

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/rprtr258/flatnotes/internal/elem"
	"github.com/rprtr258/flatnotes/internal/elem/attrs"
	"github.com/rprtr258/flatnotes/internal/elem/examples/htmx"
	"github.com/rprtr258/flatnotes/internal/elem/styles"
)

func main() {
	var count int

	app := fiber.New()
	app.Post("/increment", func(c *fiber.Ctx) error {
		count++
		return c.SendString(strconv.Itoa(count))
	})
	app.Post("/decrement", func(c *fiber.Ctx) error {
		count--
		return c.SendString(strconv.Itoa(count))
	})
	app.Get("/", func(c *fiber.Ctx) error {
		html := elem.HTML(nil,
			elem.Head(nil,
				elem.Script(attrs.Props{attrs.Src: "https://unpkg.com/htmx.org@1.9.6"}),
				elem.Style(attrs.Props{attrs.Type: "text/css"},
					map[string]styles.Props{
						"button": {
							styles.Padding:         "10px 20px",
							styles.BackgroundColor: "#007BFF",
							styles.Color:           "#fff",
							styles.BorderColor:     "#007BFF",
							styles.BorderRadius:    "5px",
							styles.Margin:          "10px",
							styles.Cursor:          "pointer",
							styles.Width:           "4rem",
						},
					}),
			),
			elem.Body(attrs.Props{
				attrs.Style: styles.Props{
					styles.BackgroundColor: "#F4F4F4",
					styles.FontFamily:      "Arial, sans-serif",
					styles.Height:          "100vh",
					styles.Display:         "flex",
					styles.FlexDirection:   "column",
					styles.AlignItems:      "center",
					styles.JustifyContent:  "center",
				}.ToInline(),
			},
				elem.H1(nil, elem.Text("Counter App")),
				elem.Div(attrs.Props{attrs.ID: "count"}, elem.Text("0")),
				elem.Span(nil,
					elem.Button(attrs.Props{
						htmx.HXPost:   "/increment",
						htmx.HXTarget: "#count",
					}, elem.Text("+")),
					elem.Button(attrs.Props{
						htmx.HXPost:   "/decrement",
						htmx.HXTarget: "#count",
					}, elem.Text("-")),
				),
			),
		)

		// Specify that the response content type is HTML before sending the response
		c.Type("html")
		return c.SendString(elem.Render(html))
	})

	app.Listen(":3000")
}
