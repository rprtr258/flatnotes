package elem

import (
	"sort"
	"strings"

	"github.com/rprtr258/flatnotes/internal/elem/attrs"
	"github.com/rprtr258/flatnotes/internal/elem/styles"
	"github.com/rprtr258/fun"
)

// ========== Document Structure ==========

// Body creates a <body> element.
func Body(attrs attrs.Props, children ...Node) Node {
	return Element("body", attrs, children...)
}

// Head creates a <head> element.
func Head(attrs attrs.Props, children ...Node) Node {
	return Element("head", attrs, children...)
}

// Title creates a <title> element.
func Title(attrs attrs.Props, children ...Node) Node {
	return Element("title", attrs, children...)
}

// ========== Text Formatting and Structure ==========

// A creates an <a> element.
func A(attrs attrs.Props, children ...Node) Node {
	return Element("a", attrs, children...)
}

// Br creates a <br> element.
func Br(attrs attrs.Props) Node {
	return Element("br", attrs)
}

// Blockquote creates a <blockquote> element.
func Blockquote(attrs attrs.Props, children ...Node) Node {
	return Element("blockquote", attrs, children...)
}

// Code creates a <code> element.
func Code(attrs attrs.Props, children ...Node) Node {
	return Element("code", attrs, children...)
}

// Div creates a <div> element.
func Div(attrs attrs.Props, children ...Node) Node {
	return Element("div", attrs, children...)
}

// Em creates an <em> element.
func Em(attrs attrs.Props, children ...Node) Node {
	return Element("em", attrs, children...)
}

// H1 creates an <h1> element.
func H1(attrs attrs.Props, children ...Node) Node {
	return Element("h1", attrs, children...)
}

// H2 creates an <h2> element.
func H2(attrs attrs.Props, children ...Node) Node {
	return Element("h2", attrs, children...)
}

// H3 creates an <h3> element.
func H3(attrs attrs.Props, children ...Node) Node {
	return Element("h3", attrs, children...)
}

// H4 creates an <h4> element.
func H4(attrs attrs.Props, children ...Node) Node {
	return Element("h4", attrs, children...)
}

// H5 creates an <h5> element.
func H5(attrs attrs.Props, children ...Node) Node {
	return Element("h5", attrs, children...)
}

// H6 creates an <h6> element.
func H6(attrs attrs.Props, children ...Node) Node {
	return Element("h6", attrs, children...)
}

// Hgroup creates an <hgroup> element.
func Hgroup(attrs attrs.Props, children ...Node) Node {
	return Element("hgroup", attrs, children...)
}

// Hr creates an <hr> element.
func Hr(attrs attrs.Props) Node {
	return Element("hr", attrs)
}

// I creates an <i> element.
func I(attrs attrs.Props, children ...Node) Node {
	return Element("i", attrs, children...)
}

// P creates a <p> element.
func P(attrs attrs.Props, children ...Node) Node {
	return Element("p", attrs, children...)
}

// Pre creates a <pre> element.
func Pre(attrs attrs.Props, children ...Node) Node {
	return Element("pre", attrs, children...)
}

// Span creates a <span> element.
func Span(attrs attrs.Props, children ...Node) Node {
	return Element("span", attrs, children...)
}

// Strong creates a <strong> element.
func Strong(attrs attrs.Props, children ...Node) Node {
	return Element("strong", attrs, children...)
}

// Sub creates a <sub> element.
func Sub(attrs attrs.Props, children ...Node) Node {
	return Element("sub", attrs, children...)
}

// Sup creates a <sub> element.
func Sup(attrs attrs.Props, children ...Node) Node {
	return Element("sup", attrs, children...)
}

// B creates a <b> element.
func B(attrs attrs.Props, children ...Node) Node {
	return Element("b", attrs, children...)
}

// U creates a <u> element.
func U(attrs attrs.Props, children ...Node) Node {
	return Element("u", attrs, children...)
}

// Text creates a TextNode.
func Text(t string) Node {
	return func(sb *strings.Builder) {
		sb.WriteString(t)
	}
}

// ========== Lists ==========

// Li creates an <li> element.
func Li(attrs attrs.Props, children ...Node) Node {
	return Element("li", attrs, children...)
}

// Ul creates a <ul> element.
func Ul(attrs attrs.Props, children ...Node) Node {
	return Element("ul", attrs, children...)
}

// Ol creates an <ol> element.
func Ol(attrs attrs.Props, children ...Node) Node {
	return Element("ol", attrs, children...)
}

// Dl creates a <dl> element.
func Dl(attrs attrs.Props, children ...Node) Node {
	return Element("dl", attrs, children...)
}

// Dt creates a <dt> element.
func Dt(attrs attrs.Props, children ...Node) Node {
	return Element("dt", attrs, children...)
}

// Dd creates a <dd> element.
func Dd(attrs attrs.Props, children ...Node) Node {
	return Element("dd", attrs, children...)
}

// ========== Forms ==========

// Button creates a <button> element.
func Button(attrs attrs.Props, children ...Node) Node {
	return Element("button", attrs, children...)
}

// Form creates a <form> element.
func Form(attrs attrs.Props, children ...Node) Node {
	return Element("form", attrs, children...)
}

// Input creates an <input> element.
func Input(attrs attrs.Props) Node {
	return Element("input", attrs)
}

// Label creates a <label> element.
func Label(attrs attrs.Props, children ...Node) Node {
	return Element("label", attrs, children...)
}

// Optgroup creates an <optgroup> element to group <option>s within a <select> element.
func Optgroup(attrs attrs.Props, children ...Node) Node {
	return Element("optgroup", attrs, children...)
}

// Option creates an <option> element.
func Option(attrs attrs.Props, content string) Node {
	return Element("option", attrs, Text(content))
}

// Select creates a <select> element.
func Select(attrs attrs.Props, children ...Node) Node {
	return Element("select", attrs, children...)
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
func Script(attrs attrs.Props, children ...Node) Node {
	return Element("script", attrs, children...)
}

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

// Article creates an <article> element.
func Article(attrs attrs.Props, children ...Node) Node {
	return Element("article", attrs, children...)
}

// Aside creates an <aside> element.
func Aside(attrs attrs.Props, children ...Node) Node {
	return Element("aside", attrs, children...)
}

// Footer creates a <footer> element.
func Footer(attrs attrs.Props, children ...Node) Node {
	return Element("footer", attrs, children...)
}

// Header creates a <header> element.
func Header(attrs attrs.Props, children ...Node) Node {
	return Element("header", attrs, children...)
}

// Main creates a <main> element.
func Main(attrs attrs.Props, children ...Node) Node {
	return Element("main", attrs, children...)
}

// Nav creates a <nav> element.
func Nav(attrs attrs.Props, children ...Node) Node {
	return Element("nav", attrs, children...)
}

// Section creates a <section> element.
func Section(attrs attrs.Props, children ...Node) Node {
	return Element("section", attrs, children...)
}

// ========== Semantic Form Elements ==========

// Fieldset creates a <fieldset> element.
func Fieldset(attrs attrs.Props, children ...Node) Node {
	return Element("fieldset", attrs, children...)
}

// Legend creates a <legend> element.
func Legend(attrs attrs.Props, children ...Node) Node {
	return Element("legend", attrs, children...)
}

// Datalist creates a <datalist> element.
func Datalist(attrs attrs.Props, children ...Node) Node {
	return Element("datalist", attrs, children...)
}

// Meter creates a <meter> element.
func Meter(attrs attrs.Props, children ...Node) Node {
	return Element("meter", attrs, children...)
}

// Output creates an <output> element.
func Output(attrs attrs.Props, children ...Node) Node {
	return Element("output", attrs, children...)
}

// Progress creates a <progress> element.
func Progress(attrs attrs.Props, children ...Node) Node {
	return Element("progress", attrs, children...)
}

// --- Semantic Interactive Elements ---

// Dialog creates a <dialog> element.
func Dialog(attrs attrs.Props, children ...Node) Node {
	return Element("dialog", attrs, children...)
}

// Menu creates a <menu> element.
func Menu(attrs attrs.Props, children ...Node) Node {
	return Element("menu", attrs, children...)
}

// --- Semantic Script Supporting Elements ---

// NoScript creates a <noscript> element.
func NoScript(attrs attrs.Props, children ...Node) Node {
	return Element("noscript", attrs, children...)
}

// --- Semantic Text Content Elements ---

// Abbr creates an <abbr> element.
func Abbr(attrs attrs.Props, children ...Node) Node {
	return Element("abbr", attrs, children...)
}

// Address creates an <address> element.
func Address(attrs attrs.Props, children ...Node) Node {
	return Element("address", attrs, children...)
}

// Cite creates a <cite> element.
func Cite(attrs attrs.Props, children ...Node) Node {
	return Element("cite", attrs, children...)
}

// Data creates a <data> element.
func Data(attrs attrs.Props, children ...Node) Node {
	return Element("data", attrs, children...)
}

// Details creates a <details> element.
func Details(attrs attrs.Props, children ...Node) Node {
	return Element("details", attrs, children...)
}

// FigCaption creates a <figcaption> element.
func FigCaption(attrs attrs.Props, children ...Node) Node {
	return Element("figcaption", attrs, children...)
}

// Figure creates a <figure> element.
func Figure(attrs attrs.Props, children ...Node) Node {
	return Element("figure", attrs, children...)
}

// Kbd creates a <kbd> element.
func Kbd(attrs attrs.Props, children ...Node) Node {
	return Element("kbd", attrs, children...)
}

// Mark creates a <mark> element.
func Mark(attrs attrs.Props, children ...Node) Node {
	return Element("mark", attrs, children...)
}

// Q creates a <q> element.
func Q(attrs attrs.Props, children ...Node) Node {
	return Element("q", attrs, children...)
}

// Samp creates a <samp> element.
func Samp(attrs attrs.Props, children ...Node) Node {
	return Element("samp", attrs, children...)
}

// Small creates a <small> element.
func Small(attrs attrs.Props, children ...Node) Node {
	return Element("small", attrs, children...)
}

// Summary creates a <summary> element.
func Summary(attrs attrs.Props, children ...Node) Node {
	return Element("summary", attrs, children...)
}

// Time creates a <time> element.
func Time(attrs attrs.Props, children ...Node) Node {
	return Element("time", attrs, children...)
}

// Var creates a <var> element.
func Var(attrs attrs.Props, children ...Node) Node {
	return Element("var", attrs, children...)
}

// ========== Tables ==========

// Table creates a <table> element.
func Table(attrs attrs.Props, children ...Node) Node {
	return Element("table", attrs, children...)
}

// THead creates a <thead> element.
func THead(attrs attrs.Props, children ...Node) Node {
	return Element("thead", attrs, children...)
}

// TBody creates a <tbody> element.
func TBody(attrs attrs.Props, children ...Node) Node {
	return Element("tbody", attrs, children...)
}

// TFoot creates a <tfoot> element.
func TFoot(attrs attrs.Props, children ...Node) Node {
	return Element("tfoot", attrs, children...)
}

// Tr creates a <tr> element.
func Tr(attrs attrs.Props, children ...Node) Node {
	return Element("tr", attrs, children...)
}

// Th creates a <th> element.
func Th(attrs attrs.Props, children ...Node) Node {
	return Element("th", attrs, children...)
}

// Td creates a <td> element.
func Td(attrs attrs.Props, children ...Node) Node {
	return Element("td", attrs, children...)
}

// ========== Embedded Content ==========

// IFrames creates an <iframe> element.
func IFrame(attrs attrs.Props, children ...Node) Node {
	return Element("iframe", attrs, children...)
}

// Audio creates an <audio> element.
func Audio(attrs attrs.Props, children ...Node) Node {
	return Element("audio", attrs, children...)
}

// Video creates a <video> element.
func Video(attrs attrs.Props, children ...Node) Node {
	return Element("video", attrs, children...)
}

// Source creates a <source> element.
func Source(attrs attrs.Props, children ...Node) Node {
	return Element("source", attrs, children...)
}

// ========== Image Map Elements ==========

// Map creates a <map> element.
func Map(attrs attrs.Props, children ...Node) Node {
	return Element("map", attrs, children...)
}

// Area creates an <area> element.
func Area(attrs attrs.Props) Node {
	return Element("area", attrs)
}
