// Package goldmark implements functions to convert markdown text to a desired format.
package goldmark

import (
	"io"

	"github.com/rprtr258/flatnotes/internal/goldmark/parser"
	"github.com/rprtr258/flatnotes/internal/goldmark/renderer"
	"github.com/rprtr258/flatnotes/internal/goldmark/renderer/html"
	"github.com/rprtr258/flatnotes/internal/goldmark/text"
	"github.com/rprtr258/flatnotes/internal/goldmark/util"
)

// DefaultParser returns a new Parser that is configured by default values.
func DefaultParser() parser.Parser {
	return parser.NewParser(parser.WithBlockParsers(parser.DefaultBlockParsers()...),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
		parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
	)
}

// DefaultRenderer returns a new Renderer that is configured by default values.
func DefaultRenderer() renderer.Renderer {
	return renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(html.NewRenderer(), 1000)))
}

var defaultMarkdown = New()

// Convert interprets a UTF-8 bytes source in Markdown and
// write rendered contents to a writer w.
func Convert(source []byte, w io.Writer, opts ...parser.ParseOption) error {
	return defaultMarkdown.Convert(source, w, opts...)
}

// Option is a functional option type for Markdown objects.
type Option func(*Markdown)

// An Extender interface is used for extending Markdown.
type Extender interface {
	// Extend extends the Markdown.
	Extend(*Markdown)
}

// WithExtensions adds extensions.
func WithExtensions(ext ...Extender) Option {
	return func(m *Markdown) {
		m.extensions = append(m.extensions, ext...)
	}
}

// WithParser allows you to override the default parser.
func WithParser(p parser.Parser) Option {
	return func(m *Markdown) {
		m.parser = p
	}
}

// WithParserOptions applies options for the parser.
func WithParserOptions(opts ...parser.Option) Option {
	return func(m *Markdown) {
		m.parser.AddOptions(opts...)
	}
}

// WithRenderer allows you to override the default renderer.
func WithRenderer(r renderer.Renderer) Option {
	return func(m *Markdown) {
		m.renderer = r
	}
}

// WithRendererOptions applies options for the renderer.
func WithRendererOptions(opts ...renderer.Option) Option {
	return func(m *Markdown) {
		m.renderer.AddOptions(opts...)
	}
}

// A Markdown interface offers functions to convert Markdown text to a desired format.
type Markdown struct {
	parser     parser.Parser
	renderer   renderer.Renderer
	extensions []Extender
}

// New returns a new Markdown with given options.
func New(options ...Option) *Markdown {
	md := &Markdown{
		parser:     DefaultParser(),
		renderer:   DefaultRenderer(),
		extensions: []Extender{},
	}
	for _, opt := range options {
		opt(md)
	}
	for _, e := range md.extensions {
		e.Extend(md)
	}
	return md
}

// Convert interprets a UTF-8 bytes source in Markdown and write rendered contents to a writer w.
func (m *Markdown) Convert(source []byte, writer io.Writer, opts ...parser.ParseOption) error {
	reader := text.NewReader(source)
	doc := m.parser.Parse(reader, opts...)
	return m.renderer.Render(writer, source, doc)
}

// Parser returns a Parser that will be used for conversion.
func (m *Markdown) Parser() parser.Parser {
	return m.parser
}

// SetParser sets a Parser to this object.
func (m *Markdown) SetParser(v parser.Parser) {
	m.parser = v
}

// Renderer returns a Renderer that will be used for conversion.
func (m *Markdown) Renderer() renderer.Renderer {
	return m.renderer
}

// SetRenderer sets a Renderer to this object.
func (m *Markdown) SetRenderer(v renderer.Renderer) {
	m.renderer = v
}
