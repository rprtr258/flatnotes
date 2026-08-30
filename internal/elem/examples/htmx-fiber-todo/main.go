package main

import (
	"sort"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/rprtr258/fun"

	"github.com/rprtr258/flatnotes/internal/elem"
	"github.com/rprtr258/flatnotes/internal/elem/attrs"
	"github.com/rprtr258/flatnotes/internal/elem/examples/htmx"
	"github.com/rprtr258/flatnotes/internal/elem/styles"
)

// Todo model
type Todo struct {
	Title string
	Done  bool
}

// ID -> Todo
var todos = map[int]Todo{
	1: {Title: "First task", Done: false},
	2: {Title: "Second task", Done: true},
}

func createTodoNode(id int, todo Todo) elem.Node {
	styleSpan := styles.Props{
		styles.TextDecoration: fun.IF(todo.Done, "line-through", "none"),
	}

	return elem.Li(attrs.Props{attrs.ID: "todo-" + strconv.Itoa(id)},
		elem.Input(attrs.Props{
			attrs.Type:    "checkbox",
			attrs.Checked: strconv.FormatBool(todo.Done),
			htmx.HXPost:   "/toggle/" + strconv.Itoa(id),
			htmx.HXTarget: "#todo-" + strconv.Itoa(id),
		}),
		elem.Span(attrs.Props{attrs.Style: styleSpan.ToInline()},
			elem.Text(todo.Title),
		),
	)
}

func renderTodos(todos map[int]Todo) string {
	inputButtonStyle := styles.Props{
		styles.Width:           "100%",
		styles.Padding:         "10px",
		styles.MarginBottom:    "10px",
		styles.Border:          "1px solid #ccc",
		styles.BorderRadius:    "4px",
		styles.BackgroundColor: "#F9F9F9",
		styles.BoxSizing:       "border-box",
	}

	buttonStyle := styles.Props{
		styles.BackgroundColor: "#007BFF",
		styles.Color:           "white",
		styles.BorderStyle:     "none",
		styles.BorderRadius:    "4px",
		styles.Cursor:          "pointer",
		styles.Width:           "100%",
		styles.Padding:         "8px 12px",
		styles.FontSize:        "14px",
		styles.Height:          "36px",
		styles.MarginRight:     "10px",
	}

	listContainerStyle := styles.Props{
		styles.ListStyleType: "none",
		styles.Padding:       "0",
		styles.Width:         "100%",
	}

	centerContainerStyle := styles.Props{
		styles.MaxWidth:        "300px",
		styles.Margin:          "40px auto",
		styles.Padding:         "20px",
		styles.Border:          "1px solid #ccc",
		styles.BoxShadow:       "0px 0px 10px rgba(0,0,0,0.1)",
		styles.BackgroundColor: "#F9F9F9",
	}

	ids := fun.Keys(todos)
	sort.Ints(ids)

	htmlContent := elem.HTML(nil,
		elem.Head(nil,
			elem.Script(attrs.Props{attrs.Src: "https://unpkg.com/htmx.org"}),
		),
		elem.Div(
			attrs.Props{attrs.Style: centerContainerStyle.ToInline()},
			elem.H1(nil, elem.Text("Todo List")),
			elem.Form(
				attrs.Props{attrs.Method: "post", attrs.Action: "/add"},
				elem.Input(
					attrs.Props{
						attrs.Type:        "text",
						attrs.Name:        "newTodo",
						attrs.Placeholder: "Add new task...",
						attrs.Style:       inputButtonStyle.ToInline(),
						attrs.Autofocus:   "true",
					},
				),
				elem.Button(
					attrs.Props{
						attrs.Type:  "submit",
						attrs.Style: buttonStyle.ToInline(),
					},
					elem.Text("Add"),
				),
			),
			elem.Ul(
				attrs.Props{attrs.Style: listContainerStyle.ToInline()},
				fun.Map[elem.Node](func(id int) elem.Node {
					return createTodoNode(id, todos[id])
				}, ids...)...,
			),
		),
	)

	return elem.Render(htmlContent)
}

func main() {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(renderTodos(todos))
	})
	app.Post("/toggle/:id", func(c *fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))
		updatedTodo := todos[id]
		updatedTodo.Done = !updatedTodo.Done
		todos[id] = updatedTodo
		return c.Type("html").SendString(elem.Render(createTodoNode(id, updatedTodo)))
	})
	app.Post("/add", func(c *fiber.Ctx) error {
		if newTitle := c.FormValue("newTodo"); newTitle != "" {
			todos[len(todos)+1] = Todo{Title: utils.CopyString(newTitle), Done: false}
		}

		return c.Redirect("/")
	})
	app.Listen(":3000")
}
