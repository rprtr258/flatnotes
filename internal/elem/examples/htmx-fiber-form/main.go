package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/rprtr258/flatnotes/internal/elem"
	"github.com/rprtr258/flatnotes/internal/elem/attrs"
	"github.com/rprtr258/flatnotes/internal/elem/examples/htmx"
)

func main() {
	app := fiber.New()

	// Define a simple structure to hold our form data
	app.Post("/submit-form", func(c *fiber.Ctx) error {
		// Capture the form data
		formDataName := c.FormValue("name")
		formDataEmail := c.FormValue("email")

		// Send a response with the captured data
		return c.SendString(fmt.Sprintf("Name: %s, Email: %s", formDataName, formDataEmail))
	})

	app.Get("/", func(c *fiber.Ctx) error {
		pageContent := elem.HTML(nil,
			elem.Head(nil,
				elem.Script(attrs.Props{attrs.Src: "https://unpkg.com/htmx.org@1.9.6"}),
			),
			elem.Body(nil,
				elem.H1(nil, elem.Text("Simple Form App")),
				elem.Form(attrs.Props{
					attrs.Action: "/submit-form",
					attrs.Method: "POST",
					htmx.HXPost:  "/submit-form",
					htmx.HXSwap:  "outerHTML",
				},
					elem.Label(attrs.Props{attrs.For: "name"}, elem.Text("Name: ")),
					elem.Input(attrs.Props{attrs.Type: "text", attrs.Name: "name", attrs.ID: "name"}),
					elem.Br(nil),
					elem.Label(attrs.Props{attrs.For: "email"}, elem.Text("Email: ")),
					elem.Input(attrs.Props{attrs.Type: "email", attrs.Name: "email", attrs.ID: "email"}),
					elem.Br(nil),
					elem.Input(attrs.Props{attrs.Type: "submit", attrs.Value: "Submit"}),
				),
				elem.Div(attrs.Props{attrs.ID: "response"}, elem.Text("")),
			),
		)

		// Specify that the response content type is HTML before sending the response
		c.Type("html")
		return c.SendString(elem.Render(pageContent))
	})

	app.Listen(":3000")
}
