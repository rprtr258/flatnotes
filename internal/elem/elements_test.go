package elem

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rprtr258/flatnotes/internal/elem/attrs"
	"github.com/rprtr258/flatnotes/internal/elem/styles"
)

func TestRender(t *testing.T) {
	for name, test := range map[string]struct {
		el       Node
		expected string
	}{
		// ========== Document Structure ==========
		"TestBody": {
			expected: `<body class="page-body"><p>Welcome to Elem!</p></body>`,
			el:       Body(attrs.Props{attrs.Class: "page-body"}, P(nil, Text("Welcome to Elem!"))),
		},
		"TestHtml": {
			expected: `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>Elem Page</title></head><body><p>Welcome to Elem!</p></body></html>`,
			el: HTML(attrs.Props{attrs.Lang: "en"},
				Head(nil,
					Meta(attrs.Props{attrs.Charset: "UTF-8"}),
					Title(nil, Text("Elem Page")),
				),
				Body(nil, P(nil, Text("Welcome to Elem!"))),
			),
		},
		// ========== Text Formatting and Structure ==========
		"TestA": {
			expected: `<a href="https://example.com">Visit Example</a>`,
			el:       A(attrs.Props{attrs.Href: "https://example.com"}, Text("Visit Example")),
		},
		"TestBlockquote": {
			expected: `<blockquote>Quote text</blockquote>`,
			el:       Blockquote(nil, Text("Quote text")),
		},
		"TestBr": {
			expected: `<br>`,
			el:       Br(nil),
		},
		"TestCode": {
			expected: `<code>Code snippet</code>`,
			el:       Code(nil, Text("Code snippet")),
		},
		"TestDiv": {
			expected: `<div class="container">Hello, Elem!</div>`,
			el:       Div(attrs.Props{attrs.Class: "container"}, Text("Hello, Elem!")),
		},
		"TestEm": {
			expected: `<em>Italic text</em>`,
			el:       Em(nil, Text("Italic text")),
		},
		"TestH1": {
			expected: `<h1 class="title">Hello, Elem!</h1>`,
			el:       H1(attrs.Props{attrs.Class: "title"}, Text("Hello, Elem!")),
		},
		"TestH2": {
			expected: `<h2 class="subtitle">Hello, Elem!</h2>`,
			el:       H2(attrs.Props{attrs.Class: "subtitle"}, Text("Hello, Elem!")),
		},
		"TestH3": {
			expected: `<h3>Hello, Elem!</h3>`,
			el:       H3(nil, Text("Hello, Elem!")),
		},
		"TestH4": {
			expected: `<h4>Hello, Elem!</h4>`,
			el:       H4(nil, Text("Hello, Elem!")),
		},
		"TestH5": {
			expected: `<h5>Hello, Elem!</h5>`,
			el:       H5(nil, Text("Hello, Elem!")),
		},
		"TestH6": {
			expected: `<h6>Hello, Elem!</h6>`,
			el:       H6(nil, Text("Hello, Elem!")),
		},
		"TestHr": {
			expected: `<hr>`,
			el:       Hr(nil),
		},
		"TestI": {
			expected: `<i>Idiomatic Text</i>`,
			el:       I(nil, Text("Idiomatic Text")),
		},
		"TestI2": {
			expected: `<i class="fa-regular fa-face-smile"></i>`,
			el:       I(attrs.Props{attrs.Class: "fa-regular fa-face-smile"}),
		},
		"TestP": {
			expected: `<p>Hello, Elem!</p>`,
			el:       P(nil, Text("Hello, Elem!")),
		},
		"TestPre": {
			expected: `<pre>Preformatted text</pre>`,
			el:       Pre(nil, Text("Preformatted text")),
		},
		"TestSpan": {
			expected: `<span class="highlight">Hello, Elem!</span>`,
			el:       Span(attrs.Props{attrs.Class: "highlight"}, Text("Hello, Elem!")),
		},
		"TestStrong": {
			expected: `<strong>Bold text</strong>`,
			el:       Strong(nil, Text("Bold text")),
		},
		"TestSub": {
			expected: `<sub>2</sub>`,
			el:       Sub(nil, Text("2")),
		},
		"TestSup": {
			expected: `<sup>2</sup>`,
			el:       Sup(nil, Text("2")),
		},
		"TestB": {
			expected: `<b>Important text</b>`,
			el:       B(nil, Text("Important text")),
		},
		"TestU": {
			expected: `<u>Unarticulated text</u>`,
			el:       U(nil, Text("Unarticulated text")),
		},
		// ========== Comments ==========
		"TestComment": {
			expected: `<!-- this is a comment -->`,
			el:       Comment("this is a comment"),
		},
		"TestCommentInElement": {
			expected: `<div>not a comment<!-- this is a comment --></div>`,
			el:       Div(nil, Text("not a comment"), Comment("this is a comment")),
		},
		// ========== Lists ==========
		"TestLi": {
			expected: `<li>Item 1</li>`,
			el:       Li(nil, Text("Item 1")),
		},
		"TestUl": {
			expected: `<ul><li>Item 1</li><li>Item 2</li></ul>`,
			el:       Ul(nil, Li(nil, Text("Item 1")), Li(nil, Text("Item 2"))),
		},
		"TestOl": {
			expected: `<ol><li>Item 1</li><li>Item 2</li></ol>`,
			el:       Ol(nil, Li(nil, Text("Item 1")), Li(nil, Text("Item 2"))),
		},
		"TestDl": {
			expected: `<dl><dt>Term 1</dt><dd>Description 1</dd><dt>Term 2</dt><dd>Description 2</dd></dl>`,
			el:       Dl(nil, Dt(nil, Text("Term 1")), Dd(nil, Text("Description 1")), Dt(nil, Text("Term 2")), Dd(nil, Text("Description 2"))),
		},
		"TestDt": {
			expected: `<dt>Term 1</dt>`,
			el:       Dt(nil, Text("Term 1")),
		},
		"TestDd": {
			expected: `<dd>Description 1</dd>`,
			el:       Dd(nil, Text("Description 1")),
		},
		// ========== Forms ==========
		"TestButton": {
			expected: `<button class="btn">Click Me</button>`,
			el:       Button(attrs.Props{attrs.Class: "btn"}, Text("Click Me")),
		},
		"TestForm": {
			expected: `<form action="/submit" method="post"><input name="username" type="text"></form>`,
			el: Form(attrs.Props{attrs.Action: "/submit", attrs.Method: "post"},
				Input(attrs.Props{attrs.Type: "text", attrs.Name: "username"})),
		},
		"TestInput": {
			expected: `<input name="username" placeholder="Enter your username" type="text">`,
			el:       Input(attrs.Props{attrs.Type: "text", attrs.Name: "username", attrs.Placeholder: "Enter your username"}),
		},
		"TestLabel": {
			expected: `<label for="username">Username</label>`,
			el:       Label(attrs.Props{attrs.For: "username"}, Text("Username")),
		},
		"TestSelectAndOption": {
			expected: `<select name="color">` +
				`<option value="red">Red</option>` +
				`<option value="blue">Blue</option>` +
				`</select>`,
			el: Select(attrs.Props{attrs.Name: "color"},
				Option(attrs.Props{attrs.Value: "red"}, "Red"),
				Option(attrs.Props{attrs.Value: "blue"}, "Blue"),
			),
		},
		"TestSelectAndOptgroup": {
			expected: `<select name="cars">` +
				`<optgroup label="Swedish Cars">` +
				`<option value="volvo">Volvo</option>` +
				`<option value="saab">Saab</option>` +
				`</optgroup>` +
				`<optgroup label="German Cars">` +
				`<option value="mercedes">Mercedes</option>` +
				`<option value="audi">Audi</option>` +
				`</optgroup>` +
				`</select>`,
			el: Select(attrs.Props{attrs.Name: "cars"},
				Optgroup(attrs.Props{attrs.Label: "Swedish Cars"},
					Option(attrs.Props{attrs.Value: "volvo"}, "Volvo"),
					Option(attrs.Props{attrs.Value: "saab"}, "Saab"),
				),
				Optgroup(attrs.Props{attrs.Label: "German Cars"},
					Option(attrs.Props{attrs.Value: "mercedes"}, "Mercedes"),
					Option(attrs.Props{attrs.Value: "audi"}, "Audi"),
				),
			),
		},
		"TestTextarea": {
			expected: `<textarea name="comment" rows="5">Leave a comment...</textarea>`,
			el:       Textarea(attrs.Props{attrs.Name: "comment", attrs.Rows: "5"}, "Leave a comment..."),
		},
		// ========== Boolean attributes ==========
		"TestCheckedTrue": {
			expected: `<input checked name="allow" type="checkbox">`,
			el:       Input(attrs.Props{attrs.Type: "checkbox", attrs.Name: "allow", attrs.Checked: "true"}),
		},
		"TestCheckedFalse": {
			expected: `<input name="allow" type="checkbox">`,
			el:       Input(attrs.Props{attrs.Type: "checkbox", attrs.Name: "allow", attrs.Checked: "false"}),
		},
		"TestCheckedEmpty": {
			expected: `<input name="allow" type="checkbox">`,
			el:       Input(attrs.Props{attrs.Type: "checkbox", attrs.Name: "allow", attrs.Checked: ""}),
		},
		// ========== Hyperlinks and Multimedia ==========
		"TestImg": {
			expected: `<img alt="An image" src="image.jpg">`,
			el:       Img(attrs.Props{attrs.Src: "image.jpg", attrs.Alt: "An image"}),
		},
		// ========== Meta Elements ==========
		"TestLink": {
			expected: `<link href="https://example.com/styles.css" rel="stylesheet">`,
			el:       Link(attrs.Props{attrs.Rel: "stylesheet", attrs.Href: "https://example.com/styles.css"}),
		},
		"TestMeta": {
			expected: `<meta charset="UTF-8">`,
			el:       Meta(attrs.Props{attrs.Charset: "UTF-8"}),
		},
		"TestScript": {
			expected: `<script src="https://example.com/script.js"></script>`,
			el:       Script(attrs.Props{attrs.Src: "https://example.com/script.js"}),
		},
		"TestStyle": {
			expected: `<style type="text/css">.test-class {color: #333;}</style>`,
			el:       Style(attrs.Props{attrs.Type: "text/css"}, map[string]styles.Props{".test-class": {styles.Color: "#333"}}),
		},
		// ========== Semantic Elements ==========
		// --- Semantic Sectioning Elements ---
		"TestArticle": {
			expected: `<article><h2>Article Title</h2><p>Article content.</p></article>`,
			el:       Article(nil, H2(nil, Text("Article Title")), P(nil, Text("Article content."))),
		},
		"TestAside": {
			expected: `<aside><p>Sidebar content.</p></aside>`,
			el:       Aside(nil, P(nil, Text("Sidebar content."))),
		},
		"TestFooter": {
			expected: `<footer><p>Footer content.</p></footer>`,
			el:       Footer(nil, P(nil, Text("Footer content."))),
		},
		"TestHeader": {
			expected: `<header class="site-header"><h1>Welcome to Elem!</h1></header>`,
			el:       Header(attrs.Props{attrs.Class: "site-header"}, H1(nil, Text("Welcome to Elem!"))),
		},
		"TestMainElem": {
			expected: `<main><p>Main content goes here.</p></main>`,
			el:       Main(nil, P(nil, Text("Main content goes here."))),
		},
		"TestNav": {
			expected: `<nav><a href="/home">Home</a><a href="/about">About</a></nav>`,
			el: Nav(nil,
				A(attrs.Props{attrs.Href: "/home"}, Text("Home")),
				A(attrs.Props{attrs.Href: "/about"}, Text("About")),
			),
		},
		"TestSection": {
			expected: `<section><h3>Section Title</h3><p>Section content.</p></section>`,
			el:       Section(nil, H3(nil, Text("Section Title")), P(nil, Text("Section content."))),
		},
		"TestHgroup": {
			expected: `<hgroup><h1>Frankenstein</h1><p>Or: The Modern Prometheus</p></hgroup>`,
			el:       Hgroup(nil, H1(nil, Text("Frankenstein")), P(nil, Text("Or: The Modern Prometheus"))),
		},
		// --- Semantic Form Elements ---
		"TestFieldset": {
			expected: `<fieldset class="custom-fieldset"><legend>Personal Information</legend><input name="name" type="text"></fieldset>`,
			el: Fieldset(attrs.Props{attrs.Class: "custom-fieldset"},
				Legend(nil, Text("Personal Information")),
				Input(attrs.Props{attrs.Type: "text", attrs.Name: "name"}),
			),
		},
		"TestLegend": {
			expected: `<legend class="custom-legend">Legend Title</legend>`,
			el:       Legend(attrs.Props{attrs.Class: "custom-legend"}, Text("Legend Title")),
		},
		"TestDatalist": {
			expected: `<datalist id="exampleList"><option value="Option1">Option 1</option><option value="Option2">Option 2</option></datalist>`,
			el: Datalist(attrs.Props{attrs.ID: "exampleList"},
				Option(attrs.Props{attrs.Value: "Option1"}, "Option 1"),
				Option(attrs.Props{attrs.Value: "Option2"}, "Option 2"),
			),
		},
		"TestMeter": {
			expected: `<meter max="100" min="0" value="50">50%</meter>`,
			el:       Meter(attrs.Props{attrs.Min: "0", attrs.Max: "100", attrs.Value: "50"}, Text("50%")),
		},
		"TestOutput": {
			expected: `<output for="inputId" name="result">Output</output>`,
			el:       Output(attrs.Props{attrs.For: "inputId", attrs.Name: "result"}, Text("Output")),
		},
		"TestProgress": {
			expected: `<progress max="100" value="60"></progress>`,
			el:       Progress(attrs.Props{attrs.Max: "100", attrs.Value: "60"}),
		},
		// --- Semantic Interactive Elements ---
		"TestDialog": {
			expected: `<dialog open><p>This is an open dialog window</p></dialog>`,
			el:       Dialog(attrs.Props{attrs.Open: "true"}, P(nil, Text("This is an open dialog window"))),
		},
		"TestMenu": {
			expected: `<menu><li>Item One</li><li>Item Two</li></menu>`,
			el:       Menu(nil, Li(nil, Text("Item One")), Li(nil, Text("Item Two"))),
		},
		// --- Semantic Script Supporting Elements ---
		"TestNoScript": {
			expected: `<noscript><p>JavaScript is required for this application.</p></noscript>`,
			el:       NoScript(nil, P(nil, Text("JavaScript is required for this application."))),
		},
		// --- Semantic Text Content Elements ---
		"TestAbbr": {
			expected: `<abbr title="Web Hypertext Application Technology Working Group">WHATWG</abbr>`,
			el:       Abbr(attrs.Props{attrs.Title: "Web Hypertext Application Technology Working Group"}, Text("WHATWG")),
		},
		"TestAddress": {
			expected: `<address>123 Example St.</address>`,
			el:       Address(nil, Text("123 Example St.")),
		},
		"TestCite": {
			expected: `<p>My favorite book is <cite>The Reality Dysfunction</cite> by Peter F. Hamilton.</p>`,
			el: P(nil,
				Text("My favorite book is "),
				Cite(nil, Text("The Reality Dysfunction")),
				Text(" by Peter F. Hamilton."),
			),
		},
		"TestDetails": {
			expected: `<details><summary>More Info</summary><p>Details content here.</p></details>`,
			el: Details(nil,
				Summary(nil, Text("More Info")),
				P(nil, Text("Details content here.")),
			),
		},
		"TestDetailsWithOpenFalse": {
			expected: `<details><summary>More Info</summary><p>Details content here.</p></details>`,
			el: Details(attrs.Props{attrs.Open: "false"},
				Summary(nil, Text("More Info")),
				P(nil, Text("Details content here.")),
			),
		},
		"TestDetailsWithOpenTrue": {
			expected: `<details open><summary>More Info</summary><p>Details content here.</p></details>`,
			el: Details(attrs.Props{attrs.Open: "true"},
				Summary(nil, Text("More Info")),
				P(nil, Text("Details content here.")),
			),
		},
		"TestData": {
			expected: `<data value="8">Eight</data>`,
			el:       Data(attrs.Props{attrs.Value: "8"}, Text("Eight")),
		},
		"TestFigCaption": {
			expected: `<figcaption>Description of the figure.</figcaption>`,
			el:       FigCaption(nil, Text("Description of the figure.")),
		},
		"TestFigure": {
			expected: `<figure><img alt="An image" src="image.jpg"><figcaption>An image</figcaption></figure>`,
			el: Figure(nil,
				Img(attrs.Props{attrs.Src: "image.jpg", attrs.Alt: "An image"}),
				FigCaption(nil, Text("An image")),
			),
		},
		"TestKbd": {
			expected: `<p>To make George eat an apple, select <kbd>File | Eat Apple...</kbd></p>`,
			el:       P(nil, Text("To make George eat an apple, select "), Kbd(nil, Text("File | Eat Apple..."))),
		},
		"TestMark": {
			expected: `<p>You must <mark>highlight</mark> this word.</p>`,
			el:       P(nil, Text("You must "), Mark(nil, Text("highlight")), Text(" this word.")),
		},
		"TestQ": {
			expected: `<p>The W3C's mission is <q cite="https://www.w3.org/Consortium/">To lead the World Wide Web to its full potential</q>.</p>`,
			el: P(nil,
				Text("The W3C's mission is "),
				Q(attrs.Props{attrs.Cite: "https://www.w3.org/Consortium/"}, Text("To lead the World Wide Web to its full potential")),
				Text("."),
			),
		},
		"TestSamp": {
			expected: `<p>The computer said <samp>Too much cheese in tray two</samp> but I didn't know what that meant.</p>`,
			el: P(nil,
				Text("The computer said "),
				Samp(nil, Text("Too much cheese in tray two")),
				Text(" but I didn't know what that meant."),
			),
		},
		"TestSmall": {
			expected: `<p>Single room <small>breakfast included, VAT not included</small></p>`,
			el: P(nil,
				Text("Single room "),
				Small(nil, Text("breakfast included, VAT not included")),
			),
		},
		"TestSummary": {
			expected: `<details><summary>Summary Title</summary></details>`,
			el:       Details(nil, Summary(nil, Text("Summary Title"))),
		},
		"TestTime": {
			expected: `<time datetime="2023-01-01T00:00:00Z">New Year's Day</time>`,
			el:       Time(attrs.Props{attrs.Datetime: "2023-01-01T00:00:00Z"}, Text("New Year's Day")),
		},
		"TestVar": {
			expected: `<p>After a few moment's thought, she wrote <var>E</var>.</p>`,
			el: P(nil,
				Text("After a few moment's thought, she wrote "),
				Var(nil, Text("E")),
				Text("."),
			),
		},
		// ========== Tables ==========
		"TestTr": {
			expected: `<tr>Row content.</tr>`,
			el:       Tr(nil, Text("Row content.")),
		},
		"TestTd": {
			expected: `<tr><td><h1>Cell one.</h1></td><td>Cell two.</td></tr>`,
			el: Tr(nil,
				Td(nil, H1(nil, Text("Cell one."))),
				Td(nil, Text("Cell two.")),
			),
		},
		"TestTh": {
			expected: `<tr><th>First name</th><th>Last name</th><th>Age</th></tr>`,
			el: Tr(nil,
				Th(nil, Text("First name")),
				Th(nil, Text("Last name")),
				Th(nil, Text("Age")),
			),
		},
		"TestTHead": {
			expected: `<thead><tr><td>Text</td><td><a href="/link">Link</a></td></tr></thead>`,
			el: THead(nil, Tr(nil,
				Td(nil, Text("Text")),
				Td(nil, A(attrs.Props{attrs.Href: "/link"}, Text("Link"))),
			)),
		},
		"TestTBody": {
			expected: `<tbody><tr><td>Table body</td></tr></tbody>`,
			el:       TBody(nil, Tr(nil, Td(nil, Text("Table body")))),
		},
		"TestTFoot": {
			expected: `<tfoot><tr><td><a href="/footer">Table footer</a></td></tr></tfoot>`,
			el:       TFoot(nil, Tr(nil, Td(nil, A(attrs.Props{attrs.Href: "/footer"}, Text("Table footer"))))),
		},
		"TestTable": {
			expected: `<table><tr><th>Table header</th></tr><tr><td>Table content</td></tr></table>`,
			el: Table(nil,
				Tr(nil, Th(nil, Text("Table header"))),
				Tr(nil, Td(nil, Text("Table content"))),
			),
		},
		// ========== Embedded Content ==========
		"TestEmbedLink": {
			expected: `<iframe src="https://www.youtube.com/embed/446E-r0rXHI"></iframe>`,
			el:       IFrame(attrs.Props{attrs.Src: "https://www.youtube.com/embed/446E-r0rXHI"}),
		},
		"TestAllowFullScreen": {
			expected: `<iframe allowfullscreen src="https://www.youtube.com/embed/446E-r0rXHI"></iframe>`,
			el:       IFrame(attrs.Props{attrs.Src: "https://www.youtube.com/embed/446E-r0rXHI", attrs.AllowFullscreen: "true"}),
		},
		"TestAudioWithSourceElementsAndFallbackText": {
			expected: `<audio controls><source src="horse.ogg" type="audio/ogg"><source src="horse.mp3" type="audio/mpeg">Your browser does not support the audio tag.</audio>`,
			el: Audio(attrs.Props{attrs.Controls: "true"},
				Source(attrs.Props{attrs.Src: "horse.ogg", attrs.Type: "audio/ogg"}),
				Source(attrs.Props{attrs.Src: "horse.mp3", attrs.Type: "audio/mpeg"}),
				Text("Your browser does not support the audio tag."),
			),
		},
		"TestVideoWithSourceElementsAndFallbackText": {
			expected: `<video controls height="240" width="320"><source src="movie.mp4" type="video/mp4"><source src="movie.ogg" type="video/ogg">Your browser does not support the video tag.</video>`,
			el: Video(attrs.Props{attrs.Width: "320", attrs.Height: "240", attrs.Controls: "true"},
				Source(attrs.Props{attrs.Src: "movie.mp4", attrs.Type: "video/mp4"}),
				Source(attrs.Props{attrs.Src: "movie.ogg", attrs.Type: "video/ogg"}),
				Text("Your browser does not support the video tag."),
			),
		},
		// ========== Image Map Elements ==========
		"TestMapAndArea": {
			expected: `<map name="map-name"><area alt="Area 1" coords="34,44,270,350" href="#area1" shape="rect"></map>`,
			el: Map(attrs.Props{attrs.Name: "map-name"},
				Area(attrs.Props{
					attrs.Href:   "#area1",
					attrs.Alt:    "Area 1",
					attrs.Shape:  "rect",
					attrs.Coords: "34,44,270,350",
				}),
			),
		},
		// ========== Other ==========
		"TestNone": {
			el:       None(),
			expected: "",
		},
		"TestNoneInDiv": {
			expected: `<div></div>`,
			el:       Div(nil, None()),
		},
		"TestRaw": {
			expected: `<div class="test"><p>Test paragraph</p></div>`,
			el:       Raw(`<div class="test"><p>Test paragraph</p></div>`),
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, Render(test.el))
		})
	}
}
