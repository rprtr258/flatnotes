package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/samber/lo"

	"github.com/rprtr258/flatnotes/internal"
)

var (
	responseTitleExists = func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusConflict).JSON(map[string]string{
			"message": "Note with specified title already exists.",
		})
	}
	responseTitleInvalid = func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{
			"message": "Title contains invalid characters.",
		})
	}
	responseNoteNotFound = func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(map[string]string{
			"message": "The note cannot be found.",
		})
	}
)

func setupApp(app *fiber.App, config internal.Config, flatnotes internal.App) {
	// totp = (
	//     pyotp.TOTP(config.totp_key) if config.auth_type == AuthType.TOTP else None
	// )
	last_used_totp := ""

	// Display TOTP QR code
	// if config.auth_type == internal.AuthTypeTOTP{
	//     uri = totp.provisioning_uri(issuer_name="flatnotes", name=config.username)
	//     qr := QRCode()
	//     qr.add_data(uri)
	//     log.Println( "Scan this QR code with your TOTP app of choice e.g. Authy or Google Authenticator:",)
	//     qr.print_ascii()
	//     log.Printf("Or manually enter this key: %s\n",totp.secret)
	// }

	authenticate := func(c *fiber.Ctx) error {
		return c.Next()
	}
	if config.AuthType != internal.AuthTypeNone && config.AuthType != internal.AuthTypeReadOnly {
		authenticate = func(c *fiber.Ctx) error {
			authorizationHeaders := c.GetReqHeaders()[fiber.HeaderAuthorization]
			if len(authorizationHeaders) != 1 {
				return fiber.NewError(fiber.StatusUnauthorized, "missing Authorization header")
			}

			token, ok := strings.CutPrefix(authorizationHeaders[0], "Bearer ")
			if !ok {
				return fiber.NewError(fiber.StatusUnauthorized, "invalid token in Authorization header")
			}

			if err := internal.ValidateToken(config, token); err != nil {
				return fiber.NewError(fiber.StatusUnauthorized, fmt.Errorf("validate token: %w", err).Error())
			}

			return c.Next()
		}
	}

	root := func(c *fiber.Ctx) error {
		html, err := os.ReadFile("flatnotes/dist/index.html")
		if err != nil {
			return fmt.Errorf("read index.html: %w", err)
		}

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.Send(html)
	}
	app.Get("/", root)
	app.Get("/login", root)
	app.Get("/search", root)
	app.Get("/new", root)
	app.Get("/note/:title", root)

	// Get a specific note.
	app.Get("/api/notes/:title", authenticate, func(c *fiber.Ctx) error {
		title, err := url.QueryUnescape(c.Params("title"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
		}

		includeContent := c.QueryBool("include_content", true)

		res, err := flatnotes.GetNote(title, includeContent)
		if err != nil {
			switch err {
			case internal.ErrTitleInvalid:
				return responseTitleInvalid(c)
			case internal.ErrNotFound:
				return responseNoteNotFound(c)
			default:
				return err
			}
		}

		return c.JSON(res)
	})

	if config.AuthType != internal.AuthTypeReadOnly {
		if config.AuthType != internal.AuthTypeNone {
			app.Post("/api/token",
				func(c *fiber.Ctx) error {
					var data internal.LoginModel
					if err := c.BodyParser(&data); err != nil {
						return fiber.NewError(fiber.StatusBadRequest, err.Error())
					}

					res, err := internal.Authenticate(config, data, &last_used_totp)
					if err != nil {
						return fiber.NewError(fiber.StatusUnauthorized, err.Error())
					}

					return c.JSON(res)
				})
		}

		// Create a new note.
		app.Post("/api/notes", authenticate, func(c *fiber.Ctx) error {
			var data internal.NotePostModel
			if err := c.BodyParser(&data); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}
			data.Title = strings.TrimSpace(data.Title)

			res, err := flatnotes.CreateNote(data)
			if err != nil {
				switch err {
				case internal.ErrTitleInvalid:
					return responseTitleInvalid(c)
				case internal.ErrTitleExists:
					return responseTitleExists(c)
				default:
					return err
				}
			}

			return c.JSON(res)
		})

		app.Patch("/api/notes/:title", authenticate, func(c *fiber.Ctx) error {
			title, err := url.QueryUnescape(c.Params("title"))
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
			}
			title = strings.TrimSpace(title)

			var new_data internal.NotePatchModel
			if err := c.BodyParser(&new_data); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}

			res, err := flatnotes.UpdateNote(title, new_data)
			if err != nil {
				// except InvalidTitleError:
				//     return invalid_title_response
				// except FileExistsError:
				//     return title_exists_response
				// except FileNotFoundError:
				//     return note_not_found_response
				return err
			}

			return c.JSON(res)
		})

		app.Delete("/api/notes/:title", authenticate, func(c *fiber.Ctx) error {
			title, err := url.QueryUnescape(c.Params("title"))
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
			}

			if err := flatnotes.DeleteNote(title); err != nil {
				// except InvalidTitleError:
				//     return invalid_title_response
				// except FileNotFoundError:
				//     return note_not_found_response
				return err
			}

			return nil
		})
	}

	// Get a list of all indexed tags.
	app.Get("/api/tags", authenticate, func(c *fiber.Ctx) error {
		tags, err := flatnotes.GetTags()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("get tags: %w", err).Error())
		}

		return c.JSON([]string(lo.Keys(tags)))
	})

	// Perform a full text search on all notes.
	app.Get("/api/search", authenticate, func(c *fiber.Ctx) error {
		term := c.Query("term")
		sort := lo.
			Switch[string, internal.Sort](c.Query("sort")).
			Case("score", internal.SortScore).
			Case("title", internal.SortTitle).
			Case("lastModified", internal.SortLastModified).
			Default(internal.SortScore)
		order := lo.
			Switch[string, internal.Order](c.Query("order")).
			Case("desc", internal.OrderDesc).
			Case("asc", internal.OrderAsc).
			Default(internal.OrderDesc)
		limit := c.QueryInt("limit", 0)

		res, err := flatnotes.Search(term, sort, order, limit)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("search: %w", err).Error())
		}

		return c.JSON(res)
	})

	// TODO: move config to debug
	// TODO: hardcode auth type in frontend
	app.Get("/api/config", func(c *fiber.Ctx) error {
		return c.JSON(internal.ConfigModel{
			AuthType: config.AuthType,
		})
	})

	if os.Getenv("DEBUG") != "" {
		app.Get("/api/debug/index", func(c *fiber.Ctx) error {
			return c.JSON(flatnotes.Index)
		})
	}

	app.Static("/", "./flatnotes/dist")
	app.Static("/static", filepath.Join(config.DataPath, "static"))
}

func run(ctx context.Context) error {
	config := internal.NewConfig()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			switch e := err.(type) {
			case *fiber.Error:
				if e.Code == fiber.StatusNotFound {
					return c.Redirect("/")
				}

				return err
			default:
				return err
			}
		},
	})
	app.Use(logger.New())
	// app.Use(swagger.New(swagger.Config{
	// 	BasePath: "/",
	// 	FilePath: "./swagger.json", // FUCK YOU I DONT WANT TO WRITE COMMENTS AND GENERATE SHIT
	// 	Path:     "docs",
	// 	Title:    "Fiber API documentation",
	// }))

	appLogic, err := internal.New(config.DataPath)
	if err != nil {
		return fmt.Errorf("NewFlatnotes: %w", err)
	}
	// defer flatnotes.index.Close()

	setupApp(app, config, appLogic)

	go func() {
		<-ctx.Done()
		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Println("shutdown", err.Error())
		}
	}()

	return app.Listen(":8080")
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.SetFlags(log.Lshortfile | log.Flags())
	if err := run(ctx); err != nil {
		log.Fatalf("app stopped: %s", err.Error())
	}
}
