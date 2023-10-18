package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rprtr258/flatnotes/internal"
	"github.com/samber/lo"
)

func setupApp(app *fiber.App, config internal.Config, flatnotes internal.Flatnotes) {
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
			token, ok := strings.CutPrefix(c.GetReqHeaders()["Authorization"], "Bearer ")
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

		include_content := c.QueryBool("include_content", true)

		// try:
		note, err := internal.NewNote(config.DataPath, title, false)
		if err != nil {
			return fmt.Errorf("get note %q: %w", title, err)
		}

		modtime, err := note.LastModified()
		if err != nil {
			return fmt.Errorf("get last modified time %q: %w", title, err)
		}

		noteHeader := internal.NoteResponseModel{
			Title:        note.Title,
			LastModified: modtime.Unix(),
		}

		if include_content {
			content, err := note.GetContent()
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}

			return c.JSON(internal.NoteContentResponseModel{
				NoteResponseModel: noteHeader,
				Content:           lo.ToPtr(string(content)),
			})
		}

		return c.JSON(noteHeader)
		// except InvalidTitleError:
		//     return invalid_title_response
		// except FileNotFoundError:
		//     return note_not_found_response
	})

	if config.AuthType != internal.AuthTypeReadOnly {
		if config.AuthType != internal.AuthTypeNone {
			app.Post("/api/token",
				func(c *fiber.Ctx) error {
					var data internal.LoginModel
					_ = c.BodyParser(&data)

					username_correct := config.Username == data.Username

					expected_password := config.Password
					var current_totp string
					if config.AuthType == internal.AuthTypeTOTP {
						current_totp = "" // totp.now()
						// expected_password += current_totp
					}
					password_correct := expected_password == data.Password

					if !username_correct || !password_correct ||
						// Prevent TOTP from being reused
						config.AuthType == internal.AuthTypeTOTP && last_used_totp != "" && current_totp == last_used_totp {
						return fiber.NewError(fiber.StatusBadRequest, "Incorrect login credentials.")
					}

					access_token, err := internal.CreateAccessToken(config, config.Username)
					if err != nil {
						return fiber.NewError(fiber.StatusUnauthorized, fmt.Sprintf("create access token: %s", err.Error()))
					}

					if config.AuthType == internal.AuthTypeTOTP {
						last_used_totp = current_totp
					}
					return c.JSON(internal.TokenModel{
						AccessToken: access_token,
						TokenType:   "bearer",
					})
				})
		}

		// Create a new note.
		app.Post("/api/notes", authenticate, func(c *fiber.Ctx) error {
			var data internal.NotePostModel
			if err := c.BodyParser(&data); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}

			//         try:
			note, err := internal.NewNote(config.DataPath, data.Title, true)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("new note: %w", err).Error())
			}

			if err := note.SetContent([]byte(data.Content)); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("set note content: %w", err).Error())
			}

			lastModified, err := note.LastModified()
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("get last modified time: %w", err).Error())
			}

			return c.JSON(internal.NoteContentResponseModel{
				NoteResponseModel: internal.NoteResponseModel{
					Title:        note.Title,
					LastModified: lastModified.Unix(),
				},
				Content: &data.Content,
			})
			//         except InvalidTitleError:
			//             return invalid_title_response
			//         except FileExistsError:
			//             return title_exists_response
		})

		app.Patch("/api/notes/:title", authenticate, func(c *fiber.Ctx) error {
			title, err := url.QueryUnescape(c.Params("title"))
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
			}

			var new_data internal.NotePatchModel
			if err := c.BodyParser(&new_data); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}

			// try:
			note, err := internal.NewNote(config.DataPath, title, false)
			if err != nil {
				return fmt.Errorf("get note %q: %w", title, err)
			}

			if new_data.NewTitle != nil {
				if err := note.SetTitle(*new_data.NewTitle); err != nil {
					return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("set note %q title to %q: %w", title, *new_data.NewTitle, err).Error())
				}
			}
			if new_data.NewContent != nil {
				if err := note.SetContent([]byte(*new_data.NewContent)); err != nil {
					return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("set note %q content: %w", title, err).Error())
				}
			}

			doc, err := note.Document()
			if err != nil {
				return fmt.Errorf("get note data %q: %w", title, err)
			}

			return c.JSON(internal.NoteContentResponseModel{
				NoteResponseModel: internal.NoteResponseModel{
					Title:        note.Title,
					LastModified: doc.Modtime.Unix(),
				},
				Content: lo.ToPtr(doc.Content),
			})
			// except InvalidTitleError:
			//     return invalid_title_response
			// except FileExistsError:
			//     return title_exists_response
			// except FileNotFoundError:
			//     return note_not_found_response
		})

		app.Delete("/api/notes/:title", authenticate, func(c *fiber.Ctx) error {
			title, err := url.QueryUnescape(c.Params("title"))
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
			}

			note, _ := internal.NewNote(config.DataPath, title, false)
			return note.Delete()
			// except InvalidTitleError:
			//     return invalid_title_response
			// except FileNotFoundError:
			//     return note_not_found_response
		})
	}

	// Get a list of all indexed tags.
	app.Get("/api/tags", authenticate, func(c *fiber.Ctx) error {
		tags, err := flatnotes.GetTags()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("get tags: %w", err).Error())
		}

		return c.JSON([]string(tags))
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

		hits, err := flatnotes.Search(term, sort, order, limit)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("search: %w", err).Error())
		}

		res := []internal.SearchResultModel{}
		for _, hit := range hits {
			modtime, err := hit.LastModified()
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("get last modified time %q: %w", hit.Title, err).Error())
			}

			toOption := func(s string) *string {
				if s == "" {
					return nil
				}
				return &s
			}
			res = append(res, internal.SearchResultModel{
				Score:             hit.Score,
				Title:             hit.Title,
				LastModified:      modtime.Unix(),
				TitleHighlights:   toOption(hit.TitleHighlights),
				ContentHighlights: toOption(hit.ContentHighlights),
				TagMatches:        toOption(hit.TagMatches),
			})
		}
		return c.JSON(res)
	})

	app.Get("/api/config", func(c *fiber.Ctx) error {
		return c.JSON(internal.ConfigModel{
			AuthType: config.AuthType,
		})
	})

	app.Static("/", "./flatnotes/dist")
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

	flatnotes, err := internal.NewFlatnotes(config.DataPath)
	if err != nil {
		return fmt.Errorf("NewFlatnotes: %w", err)
	}
	// defer flatnotes.index.Close()

	setupApp(app, config, flatnotes)

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
