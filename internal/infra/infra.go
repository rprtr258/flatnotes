package infra

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/rprtr258/fun"
	"github.com/rs/zerolog/log"

	"github.com/rprtr258/flatnotes/internal"
	"github.com/rprtr258/flatnotes/internal/config"
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

func authenticate(cfg config.Config) func(*fiber.Ctx) error {
	if fun.Contains(cfg.AuthType, config.AuthTypeNone, config.AuthTypeReadOnly) {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		authorizationHeaders := c.GetReqHeaders()[fiber.HeaderAuthorization]
		if err := func() error {
			if len(authorizationHeaders) != 1 {
				return errors.New("missing Authorization header")
			}

			token, ok := strings.CutPrefix(authorizationHeaders[0], "Bearer ")
			if !ok {
				return errors.New("invalid token in Authorization header")
			}

			if err := validateToken(cfg, token); err != nil {
				return fmt.Errorf("validate token: %w", err)
			}

			return nil
		}(); err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		return c.Next()
	}
}

func setupApp(fapp *fiber.App, cfg config.Config, app internal.App) {
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

	authenticate := authenticate(cfg)

	root := func(c *fiber.Ctx) error {
		html, err := os.ReadFile("flatnotes/dist/index.html")
		if err != nil {
			return fmt.Errorf("read index.html: %w", err)
		}

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.Send(html)
	}
	fapp.Get("/", root)
	fapp.Get("/login", root)
	fapp.Get("/search", root)
	fapp.Get("/new", root)
	fapp.Get("/note/:title", root)

	// Get a specific note.
	fapp.Get("/api/notes/:title", authenticate, func(c *fiber.Ctx) error {
		title, err := url.QueryUnescape(c.Params("title"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
		}

		res, err := app.GetNote(title)
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

	if cfg.AuthType != config.AuthTypeReadOnly {
		if cfg.AuthType != config.AuthTypeNone {
			fapp.Post("/api/token",
				func(c *fiber.Ctx) error {
					var data internal.LoginModel
					if err := c.BodyParser(&data); err != nil {
						return fiber.NewError(fiber.StatusBadRequest, err.Error())
					}

					res, err := Authenticate(cfg, data, &last_used_totp)
					if err != nil {
						return fiber.NewError(fiber.StatusUnauthorized, err.Error())
					}

					return c.JSON(res)
				})
		}

		// Create a new note.
		fapp.Post("/api/notes", authenticate, func(c *fiber.Ctx) error {
			var data struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			if err := c.BodyParser(&data); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}
			data.Title = strings.TrimSpace(data.Title)

			res, err := app.CreateNote(data.Title, data.Content)
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

		fapp.Patch("/api/notes/:title", authenticate, func(c *fiber.Ctx) error {
			title, err := url.QueryUnescape(c.Params("title"))
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
			}
			title = strings.TrimSpace(title)

			var new_data internal.NotePatchModel
			if err := c.BodyParser(&new_data); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}

			res, err := app.UpdateNote(title, new_data)
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

		fapp.Delete("/api/notes/:title", authenticate, func(c *fiber.Ctx) error {
			title, err := url.QueryUnescape(c.Params("title"))
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
			}

			if err := app.DeleteNote(title); err != nil {
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
	fapp.Get("/api/tags", authenticate, func(c *fiber.Ctx) error {
		tags, err := app.GetTags()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("get tags: %w", err).Error())
		}

		return c.JSON(tags.List())
	})

	// Perform a full text search on all notes.
	fapp.Get("/api/search", authenticate, func(c *fiber.Ctx) error {
		term := c.Query("term")
		sort := fun.
			Switch[internal.Sort, string](c.Query("sort"), internal.SortScore).
			Case(internal.SortScore, "score").
			Case(internal.SortTitle, "title").
			Case(internal.SortLastModified, "lastModified").
			End()
		order := fun.
			Switch[internal.Order, string](c.Query("order"), internal.OrderDesc).
			Case(internal.OrderDesc, "desc").
			Case(internal.OrderAsc, "asc").
			End()
		limit := c.QueryInt("limit", 0)

		res, err := app.Search(term, sort, order, limit)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("search: %w", err).Error())
		}

		return c.JSON(res)
	})

	// TODO: move config to debug
	// TODO: hardcode auth type in frontend
	fapp.Get("/api/config", func(c *fiber.Ctx) error {
		return c.JSON(internal.ConfigModel{
			AuthType: string(cfg.AuthType),
		})
	})

	if os.Getenv("DEBUG") != "" {
		fapp.Get("/api/debug/index", func(c *fiber.Ctx) error {
			return c.JSON(app.Index)
		})
	}

	fapp.Static("/", "./flatnotes/dist")
	fapp.Static("/static", filepath.Join(cfg.DataPath, "static"))
}

func Run(ctx context.Context, cfg config.Config) error {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*fiber.Error); ok {
				if e.Code == fiber.StatusNotFound {
					return c.Redirect("/")
				}

				return c.Status(e.Code).SendString(e.Message)
			}

			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		},
	})
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &log.Logger,
	}))
	// app.Use(swagger.New(swagger.Config{
	// 	BasePath: "/",
	// 	FilePath: "./swagger.json", // FUCK YOU I DONT WANT TO WRITE COMMENTS AND GENERATE SHIT
	// 	Path:     "docs",
	// 	Title:    "Fiber API documentation",
	// }))

	appLogic, err := internal.New(cfg.DataPath)
	if err != nil {
		return fmt.Errorf("NewFlatnotes: %w", err)
	}
	// defer flatnotes.index.Close()

	setupApp(app, cfg, appLogic)

	go func() {
		<-ctx.Done()
		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Err(err).Msg("shutdown")
		}
	}()

	return app.Listen(":8080")
}
