package elem

import (
	"sort"
	"strings"

	"github.com/rprtr258/fun"

	"github.com/rprtr258/flatnotes/internal/elem/attrs"
	"github.com/rprtr258/flatnotes/internal/elem/styles"
)

func elem(tag string) func(attrs.Props, ...Node) Node {
	return func(attrs attrs.Props, children ...Node) Node {
		return Element(tag, attrs, children...)
	}
}

// ========== Document Structure ==========

var (
	Body  = elem("body")  // Body creates a <body> element.
	Head  = elem("head")  // Head creates a <head> element.
	Title = elem("title") // Title creates a <title> element.
)

// ========== Text Formatting and Structure ==========

var (
	A          = elem("a")          // A creates an <a> element.
	Blockquote = elem("blockquote") // Blockquote creates a <blockquote> element.
	Code       = elem("code")       // Code creates a <code> element.
	Div        = elem("div")        // Div creates a <div> element.
	Em         = elem("em")         // Em creates an <em> element.
	H1         = elem("h1")         // H1 creates an <h1> element.
	H2         = elem("h2")         // H2 creates an <h2> element.
	H3         = elem("h3")         // H3 creates an <h3> element.
	H4         = elem("h4")         // H4 creates an <h4> element.
	H5         = elem("h5")         // H5 creates an <h5> element.
	H6         = elem("h6")         // H6 creates an <h6> element.
	Hgroup     = elem("hgroup")     // Hgroup creates an <hgroup> element.
	I          = elem("i")          // I creates an <i> element.
	P          = elem("p")          // P creates a <p> element.
	Pre        = elem("pre")        // Pre creates a <pre> element.
	Span       = elem("span")       // Span creates a <span> element.
	Strong     = elem("strong")     // Strong creates a <strong> element.
	Sub        = elem("sub")        // Sub creates a <sub> element.
	Sup        = elem("sup")        // Sup creates a <sub> element.
	B          = elem("b")          // B creates a <b> element.
	U          = elem("u")          // U creates a <u> element.
)

// Br creates a <br> element.
func Br(attrs attrs.Props) Node {
	return Element("br", attrs)
}

// Hr creates an <hr> element.
func Hr(attrs attrs.Props) Node {
	return Element("hr", attrs)
}

// ========== Text creates a TextNode. ==========
func Text(t string) Node {
	return func(sb *strings.Builder) {
		sb.WriteString(t)
	}
}

// ========== Lists ==========

var (
	Li = elem("li") // Li creates an <li> element.
	Ul = elem("ul") // Ul creates a <ul> element.
	Ol = elem("ol") // Ol creates an <ol> element.
	Dl = elem("dl") // Dl creates a <dl> element.
	Dt = elem("dt") // Dt creates a <dt> element.
	Dd = elem("dd") // Dd creates a <dd> element.
)

// ========== Forms ==========

var (
	Button   = elem("button")   // Button creates a <button> element.
	Form     = elem("form")     // Form creates a <form> element.
	Label    = elem("label")    // Label creates a <label> element.
	Optgroup = elem("optgroup") // Optgroup creates an <optgroup> element to group <option>s within a <select> element.
	Select   = elem("select")   // Select creates a <select> element.
)

// Option creates an <option> element.
func Option(attrs attrs.Props, content string) Node {
	return Element("option", attrs, Text(content))
}

// Input creates an <input> element.
func Input(attrs attrs.Props) Node {
	return Element("input", attrs)
}

// Textarea creates a <textarea> element.
func Textarea(attrs attrs.Props, content string) Node {
	return Element("textarea", attrs, Text(content))
}

// ========== Hyperlinks and Multimedia ==========

// Img creates an <img> element.
func Img(attrs attrs.Props) Node {
	return Element("img", attrs)
}

// ========== Meta Elements ==========

// Link creates a <link> element.
func Link(attrs attrs.Props) Node {
	return Element("link", attrs)
}

// Meta creates a <meta> element.
func Meta(attrs attrs.Props) Node {
	return Element("meta", attrs)
}

// Script creates a <script> element.
var Script = elem("script")

// Style creates a <style> element.
// CSS creates a new CSSNode from the given CSS-selector -> Props map
func Style(attrs attrs.Props, css map[string]styles.Props) Node {
	selectors := fun.Keys(css)
	sort.Strings(selectors)

	cn := strings.Join(fun.Map[string](func(selector string) string {
		return selector + " {" + css[selector].ToInline() + "}"
	}, selectors...), "\n")
	return Element("style", attrs, func(sb *strings.Builder) {
		sb.WriteString(cn)
	})
}

// ========== Semantic Elements ==========

// --- Semantic Sectioning Elements ---

var (
	Article = elem("article") // Article creates an <article> element.
	Aside   = elem("aside")   // Aside creates an <aside> element.
	Footer  = elem("footer")  // Footer creates a <footer> element.
	Header  = elem("header")  // Header creates a <header> element.
	Main    = elem("main")    // Main creates a <main> element.
	Nav     = elem("nav")     // Nav creates a <nav> element.
	Section = elem("section") // Section creates a <section> element.
)

// ========== Semantic Form Elements ==========

var (
	Fieldset = elem("fieldset") // Fieldset creates a <fieldset> element.
	Legend   = elem("legend")   // Legend creates a <legend> element.
	Datalist = elem("datalist") // Datalist creates a <datalist> element.
	Meter    = elem("meter")    // Meter creates a <meter> element.
	Output   = elem("output")   // Output creates an <output> element.
	Progress = elem("progress") // Progress creates a <progress> element.
)

// --- Semantic Interactive Elements ---
var (
	Dialog = elem("dialog") // Dialog creates a <dialog> element.
	Menu   = elem("menu")   // Menu creates a <menu> element.
)

// --- Semantic Script Supporting Elements ---

// NoScript creates a <noscript> element.
var NoScript = elem("noscript")

// --- Semantic Text Content Elements ---

var (
	Abbr       = elem("abbr")       // Abbr creates an <abbr> element.
	Address    = elem("address")    // Address creates an <address> element.
	Cite       = elem("cite")       // Cite creates a <cite> element.
	Data       = elem("data")       // Data creates a <data> element.
	Details    = elem("details")    // Details creates a <details> element.
	FigCaption = elem("figcaption") // FigCaption creates a <figcaption> element.
	Figure     = elem("figure")     // Figure creates a <figure> element.
	Kbd        = elem("kbd")        // Kbd creates a <kbd> element.
	Mark       = elem("mark")       // Mark creates a <mark> element.
	Q          = elem("q")          // Q creates a <q> element.
	Samp       = elem("samp")       // Samp creates a <samp> element.
	Small      = elem("small")      // Small creates a <small> element.
	Summary    = elem("summary")    // Summary creates a <summary> element.
	Time       = elem("time")       // Time creates a <time> element.
	Var        = elem("var")        // Var creates a <var> element.
)

// ========== Tables ==========

var (
	Table = elem("table") // Table creates a <table> element.
	THead = elem("thead") // THead creates a <thead> element.
	TBody = elem("tbody") // TBody creates a <tbody> element.
	TFoot = elem("tfoot") // TFoot creates a <tfoot> element.
	Tr    = elem("tr")    // Tr creates a <tr> element.
	Th    = elem("th")    // Th creates a <th> element.
	Td    = elem("td")    // Td creates a <td> element.
)

// ========== Embedded Content ==========

var (
	IFrame = elem("iframe") // IFrame creates an <iframe> element.
	Audio  = elem("audio")  // Audio creates an <audio> element.
	Video  = elem("video")  // Video creates an <video> element.
	Source = elem("source") // Source creates an <source> element.
)

// ========== Image Map Elements ==========

// Map creates a <map> element.
var Map = elem("map")

// Area creates an <area> element.
func Area(attrs attrs.Props) Node {
	return Element("area", attrs)
}
