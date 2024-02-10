package infra

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	elem "github.com/chasefleming/elem-go"
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

func handlerPage(content elem.Node) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)

		var sb strings.Builder
		page(content).RenderTo(&sb, elem.RenderOptions{})
		return c.SendString(sb.String())
	}
}

type handlers struct {
	app *internal.App
}

// Get a list of all indexed tags
func (a handlers) listTags(c *fiber.Ctx) error {
	tags, err := a.app.GetTags()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("get tags: %w", err).Error())
	}

	return c.JSON(tags.List())
}

// Get a specific note
func (a handlers) getNote(c *fiber.Ctx) error {
	title, err := url.QueryUnescape(c.Params("title"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
	}

	res, err := a.app.GetNote(title)
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
}

// Perform a full text search on all notes
func (a handlers) searchNotes(c *fiber.Ctx) error {
	term := c.Query("term")
	sort := fun.
		Switch(c.Query("sort"), internal.SortScore).
		Case("score", internal.SortScore).
		Case("title", internal.SortTitle).
		Case("lastModified", internal.SortLastModified).
		End()
	order := fun.
		Switch(c.Query("order"), internal.OrderDesc).
		Case("desc", internal.OrderDesc).
		Case("asc", internal.OrderAsc).
		End()
	limit := c.QueryInt("limit", 0)

	res, err := a.app.Search(term, sort, order, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("search: %w", err).Error())
	}

	return c.JSON(res)
}

// Create a new note
func (a handlers) createNote(c *fiber.Ctx) error {
	var data internal.NotePostModel
	if err := c.BodyParser(&data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	data.Title = strings.TrimSpace(data.Title)

	res, err := a.app.CreateNote(data)
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
}

func (a handlers) updateNote(c *fiber.Ctx) error {
	title, err := url.QueryUnescape(c.Params("title"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
	}
	title = strings.TrimSpace(title)

	var new_data internal.NotePatchModel
	if err := c.BodyParser(&new_data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	res, err := a.app.UpdateNote(title, new_data)
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
}

func (a handlers) deleteNote(c *fiber.Ctx) error {
	title, err := url.QueryUnescape(c.Params("title"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
	}

	if err := a.app.DeleteNote(title); err != nil {
		// except InvalidTitleError:
		//     return invalid_title_response
		// except FileNotFoundError:
		//     return note_not_found_response
		return err
	}

	return nil
}

func setupApp(app *fiber.App, cfg config.Config, flatnotes internal.App) {
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
		if fun.Contains(cfg.AuthType, config.AuthTypeNone, config.AuthTypeReadOnly) {
			return c.Next()
		}

		if err := func() error {
			authorizationHeaders := c.GetReqHeaders()[fiber.HeaderAuthorization]
			if len(authorizationHeaders) != 1 {
				return errors.New("missing Authorization header")
			}

			token, ok := strings.CutPrefix(authorizationHeaders[0], "Bearer ")
			if !ok {
				return errors.New("invalid token in Authorization header")
			}

			if err := ValidateToken(cfg, token); err != nil {
				return fmt.Errorf("validate token: %w", err)
			}

			return nil
		}(); err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		return c.Next()
	}

	a := handlers{app: &flatnotes}
	app.Get("/api/notes/:title", authenticate, a.getNote)
	app.Get("/api/tags", authenticate, a.listTags)
	app.Get("/api/search", authenticate, a.searchNotes)

	app.Get("/", func(c *fiber.Ctx) error {
		tags, err := a.app.GetTags()
		if err != nil {
			return err
		}

		notes, err := a.app.Search("*", internal.SortScore, internal.OrderDesc, 5)
		if err != nil {
			return err
		}

		return handlerPage(viewHome(
			cfg.AuthType,
			notes,
			tags.List(),
		))(c)
	})
	app.Get("/login", handlerPage(viewLogin()))
	app.Get("/search", func(c *fiber.Ctx) error {
		term := c.Query("term")
		sort := fun.
			Switch(c.Query("sort"), internal.SortScore).
			Case("score", internal.SortScore).
			Case("title", internal.SortTitle).
			Case("lastModified", internal.SortLastModified).
			End()
		order := fun.
			Switch(c.Query("order"), internal.OrderDesc).
			Case("desc", internal.OrderDesc).
			Case("asc", internal.OrderAsc).
			End()
		limit := c.QueryInt("limit", 0)

		res, err := a.app.Search(term, sort, order, limit)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("search: %w", err).Error())
		}

		return handlerPage(viewSearch(
			cfg.AuthType,
			fun.Map[searchResult](func(note internal.SearchResultModel) searchResult {
				return newSearchResult(
					note.SearchResult.Title,
					note.LastModified,
					note.SearchResult.Score,
					note.TitleHighlights,
					note.ContentHighlights,
					note.TagMatches,
				)
			}, res...),
		))(c)
	})
	app.Get("/new", handlerPage(viewNote(internal.NoteContentResponseModel{})))
	app.Get("/note/:title", func(c *fiber.Ctx) error {
		title, err := url.QueryUnescape(c.Params("title"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
		}

		res, err := a.app.GetNote(title)
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

		return handlerPage(viewNote(res))(c)
	})

	if cfg.AuthType != config.AuthTypeReadOnly {
		if cfg.AuthType != config.AuthTypeNone {
			app.Post("/api/token",
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

		app.Post("/api/notes", authenticate, a.createNote)
		app.Patch("/api/notes/:title", authenticate, a.updateNote)
		app.Delete("/api/notes/:title", authenticate, a.deleteNote)
	}

	if os.Getenv("DEBUG") != "" {
		app.Get("/api/debug/config", func(c *fiber.Ctx) error {
			return c.JSON(internal.ConfigModel{
				AuthType: string(cfg.AuthType),
			})
		})

		app.Get("/api/debug/index", func(c *fiber.Ctx) error {
			return c.JSON(flatnotes.Index)
		})
	}

	// app.Static("/", "./flatnotes/dist")
	// app.Static("/static", filepath.Join(cfg.DataPath, "static"))
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

			return err
		},
	})
	app.Use(fiberzerolog.New())
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

	setupApp(app, cfg, appLogic)

	go func() {
		<-ctx.Done()
		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Err(err).Msg("shutdown")
		}
	}()

	return app.Listen(":8080")
}
