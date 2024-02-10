package infra

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/rprtr258/fun"
	"github.com/rs/zerolog/log"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/toc"
	"mvdan.cc/xurls/v2"

	"github.com/rprtr258/flatnotes/internal"
	"github.com/rprtr258/flatnotes/internal/config"
	"github.com/rprtr258/flatnotes/internal/template"
)

type (
	pageHomeData struct {
		AuthType config.AuthType
		Notes    []internal.SearchResultModel
		Tags     []string
	}
	pageNoteData struct {
		Note         internal.NoteContentResponseModel
		RenderedNote template.HTML
	}

	searchResult struct {
		Title                  string
		Score                  float64
		TitleHighlights        template.HTML
		ContentHighlights      template.HTML
		TagMatches             []string
		Href                   string
		LastModified           time.Time
		TitleHighlightsOrTitle template.HTML
	}
	resultGroup struct {
		Name          string
		SearchResults []searchResult
	}
	pageSearchData struct {
		AuthType        config.AuthType
		ShowHighlights  bool
		SortByIsGrouped bool
		SearchTerm      string
		SortOptions     []internal.Sort
		ResultsGrouped  []resultGroup
	}
)

var (
	pageHome     = template.Must(template.ParseFiles[pageHomeData]("templates/base.html", "templates/page_home.html", "templates/search_input.html"))
	pageNotFound = template.Must(template.ParseFiles[config.AuthType]("templates/base.html", "templates/page_not_found.html"))
	pageNote     = template.Must(template.ParseFiles[pageNoteData]("templates/base.html", "templates/note.html", "templates/page_viewer.html"))
	pageNew      = template.Must(template.ParseFiles[struct{}]("templates/base.html", "templates/note.html", "templates/page_editor.html"))
	pageSearch   = template.Must(template.ParseFiles[pageSearchData]("templates/base.html", "templates/page_search.html", "templates/search_input.html"))
	// TODO: fix
	pageLogin = template.Must(template.ParseFiles[struct{}]("templates/page_login.html"))
)

const (
	_routeHome   = "/"
	_routeLogin  = "/login"
	_routeNote   = "/note"
	_routeSearch = "/search"
	_routeNew    = "/new"
)

var (
	responseTitleExists  = fiber.NewError(fiber.StatusConflict, "Note with specified title already exists.")
	responseTitleInvalid = fiber.NewError(fiber.StatusBadRequest, "Title contains invalid characters.")
)

func handlerPage[T any](t *template.Template[T], data T) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return t.Execute(c, data)
	}
}

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
	lastUsedTotp := ""

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

	// Get a specific note
	fapp.Get("/api/notes/:title", authenticate, func(c *fiber.Ctx) error {
		title, err := url.QueryUnescape(c.Params("title"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
		}

		res, err := app.GetNote(title)
		if err != nil {
			switch err {
			case internal.ErrTitleInvalid:
				return responseTitleInvalid
			case internal.ErrNotFound:
				return handlerPage(pageNotFound, cfg.AuthType)(c)
			default:
				return err
			}
		}

		return c.JSON(res)
	})
	// Get a list of all indexed tags
	fapp.Get("/api/tags", authenticate, func(c *fiber.Ctx) error {
		tags, err := app.GetTags()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("get tags: %w", err).Error())
		}

		return c.JSON(tags.List())
	})
	// Perform a full text search on all notes
	fapp.Get("/api/search", authenticate, func(c *fiber.Ctx) error {
		term := c.Query("term")
		sort := internal.Sort(c.QueryInt("sort"))
		order := internal.Order(c.Query("order"))
		limit := c.QueryInt("limit", 0)

		res, err := app.Search(term, sort, order, limit)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("search: %w", err).Error())
		}

		return c.JSON(res)
	})

	fapp.Get(_routeHome, func(c *fiber.Ctx) error {
		tags, err := app.GetTags()
		if err != nil {
			return err
		}

		notes, err := app.Search("*", internal.SortScore, internal.OrderDesc, 5)
		if err != nil {
			return err
		}

		return handlerPage(pageHome, pageHomeData{
			AuthType: cfg.AuthType,
			Notes:    notes,
			Tags:     tags.List(),
		})(c)
	})
	fapp.Get(_routeLogin, handlerPage(pageLogin, struct{}{}))
	fapp.Get(_routeSearch, func(c *fiber.Ctx) error {
		term := c.Query("term")
		sort := internal.Sort(c.QueryInt("sort"))
		order := internal.Order(c.Query("order"))
		limit := c.QueryInt("limit", 0)

		res, err := app.Search(term, sort, order, limit)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Errorf("search: %w", err).Error())
		}

		searchResults := fun.Map[searchResult](func(note internal.SearchResultModel) searchResult {
			title := note.SearchResult.Title
			titleHighlights := template.HTML(note.TitleHighlights)
			return searchResult{
				Title:                  title,
				Score:                  note.SearchResult.Score,
				TitleHighlights:        titleHighlights,
				ContentHighlights:      template.HTML(note.ContentHighlights),
				TagMatches:             note.TagMatches,
				Href:                   _routeNote + "/" + url.PathEscape(title),
				LastModified:           note.LastModified,
				TitleHighlightsOrTitle: fun.IF(titleHighlights != "", titleHighlights, template.HTML(title)),
			}
		}, res...)
		sortBy := internal.SortScore
		return handlerPage(pageSearch, pageSearchData{
			AuthType:        cfg.AuthType,
			ShowHighlights:  true,
			SortByIsGrouped: sortBy == internal.SortTitle,
			SearchTerm:      term,
			SortOptions:     internal.SortOptions,
			ResultsGrouped: func() []resultGroup {
				// data
				//   searchFailed: false,
				//   searchFailedMessage: "Failed to load Search Results",
				//   searchFailedIcon: null,
				//   searchResults: null,
				//   searchResultsIncludeHighlights: null,

				switch sortBy {
				case internal.SortTitle:
					const specialCharGroupTitle = "#"
					const _alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

					notesGroupedDict := fun.GroupBy(
						func(searchResult searchResult) string {
							firstChar, _ := utf8.DecodeRuneInString(searchResult.Title)
							firstCharUpper := strings.ToUpper(string(firstChar))
							if strings.ContainsAny(firstCharUpper, _alphabet) {
								return firstCharUpper
							}

							return specialCharGroupTitle
						},
						searchResults...,
					)

					// Convert dict to an array skipping empty groups
					notesGroupedArray := fun.MapToSlice(notesGroupedDict, func(name string, results []searchResult) resultGroup {
						slices.SortFunc(results, func(a, b searchResult) int {
							return cmp.Compare(a.Title, b.Title)
						})
						return resultGroup{
							Name: name,
							// Sort by title within each group
							SearchResults: results,
						}
					})

					// Ensure the array is ordered correctly
					slices.SortFunc(notesGroupedArray, func(a, b resultGroup) int {
						return cmp.Compare(a.Name, b.Name)
					})

					return notesGroupedArray
				case internal.SortLastModified:
					slices.SortFunc(searchResults, func(a, b searchResult) int {
						return cmp.Compare(
							a.LastModified.Unix(),
							b.LastModified.Unix(),
						)
					})
				default:
					slices.SortFunc(searchResults, func(a, b searchResult) int {
						return -cmp.Compare(a.Score, b.Score)
					})
				}
				return []resultGroup{{
					Name:          "_",
					SearchResults: searchResults,
				}}
			}(),
		})(c)
	})
	fapp.Get(_routeNew, handlerPage(pageNew, struct{}{}))
	fapp.Get(_routeNote+"/:title", func(c *fiber.Ctx) error {
		title, err := url.QueryUnescape(c.Params("title"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
		}

		res, err := app.GetNote(title)
		if err != nil {
			switch err {
			case internal.ErrTitleInvalid:
				return responseTitleInvalid
			case internal.ErrNotFound:
				return handlerPage(pageNotFound, cfg.AuthType)(c)
			default:
				return err
			}
		}

		return handlerPage(pageNote, pageNoteData{
			Note: res,
			RenderedNote: func() template.HTML {
				var sb strings.Builder
				context := parser.NewContext()
				_ = goldmark.
					New(
						goldmark.WithExtensions(
							highlighting.NewHighlighting(
								highlighting.WithStyle("doom-one2"),
							),
							extension.NewLinkify(
								extension.WithLinkifyAllowedProtocols([]string{
									"http:",
									"https:",
								}),
								extension.WithLinkifyURLRegexp(
									xurls.Strict(),
								),
							),
							meta.Meta,
							&toc.Extender{},
						),
						goldmark.WithRendererOptions(
							html.WithXHTML(),
							html.WithUnsafe(),
						),
						goldmark.WithParserOptions(parser.WithAutoHeadingID()),
					).
					Convert([]byte(res.Content), &sb, parser.WithContext(context))

				metadata := meta.Get(context)
				_ = metadata

				s := sb.String()
				s = strings.ReplaceAll(s, "\n", "<br>")
				s = strings.ReplaceAll(s, "</h2><br>", "</h2>")
				s = strings.ReplaceAll(s, "</p><br>", "</p>")
				return template.HTML(s)
			}(),
		})(c)
	})

	if cfg.AuthType != config.AuthTypeReadOnly {
		if cfg.AuthType != config.AuthTypeNone {
			fapp.Post("/api/token",
				func(c *fiber.Ctx) error {
					var data internal.LoginModel
					if err := c.BodyParser(&data); err != nil {
						return fiber.NewError(fiber.StatusBadRequest, err.Error())
					}

					res, err := Authenticate(cfg, data, &lastUsedTotp)
					if err != nil {
						return fiber.NewError(fiber.StatusUnauthorized, err.Error())
					}

					return c.JSON(res)
				})
		}

		// Create a new note
		fapp.Post("/api/notes" /*authenticate,*/, func(c *fiber.Ctx) error {
			type NotePostModel struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			data := NotePostModel{
				Title:   strings.TrimSpace(c.FormValue("title")),
				Content: c.FormValue("content"),
			}

			res, err := app.CreateNote(data.Title, data.Content)
			if err != nil {
				switch err {
				case internal.ErrTitleInvalid:
					return responseTitleInvalid
				case internal.ErrTitleExists:
					return responseTitleExists
				default:
					return err
				}
			}

			c.Set("HX-Redirect", "/note/"+res.Title)
			return c.JSON(res)
		})
		fapp.Patch("/api/notes/:title", authenticate, func(c *fiber.Ctx) error {
			title, err := url.QueryUnescape(c.Params("title"))
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Errorf("invalid title: %w", err).Error())
			}
			title = strings.TrimSpace(title)

			var newData internal.NotePatchModel
			if err := c.BodyParser(&newData); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}

			res, err := app.UpdateNote(title, newData)
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

	if os.Getenv("DEBUG") != "" {
		fapp.Get("/api/debug/config", func(c *fiber.Ctx) error {
			return c.JSON(internal.ConfigModel{
				AuthType: string(cfg.AuthType),
			})
		})

		fapp.Get("/api/debug/index", func(c *fiber.Ctx) error {
			return c.JSON(app.Index)
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

	setupApp(app, cfg, appLogic)

	go func() {
		<-ctx.Done()
		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Err(err).Msg("shutdown")
		}
	}()

	return app.Listen(":8080")
}
