package attrs

const (
	// Universal Attributes

	Alt             = "alt"
	Class           = "class"
	Contenteditable = "contenteditable"
	Dir             = "dir" // Direction, e.g., "ltr" or "rtl"
	ID              = "id"
	Lang            = "lang"
	Style           = "style"
	Tabindex        = "tabindex"
	Title           = "title"
	Loading         = "loading"

	// Link/Script Attributes

	Async       = "async"
	Crossorigin = "crossorigin"
	Defer       = "defer"
	Href        = "href"
	Integrity   = "integrity"
	Rel         = "rel"
	Src         = "src"
	Target      = "target"

	// Meta Attributes

	Charset   = "charset"
	Content   = "content"
	HTTPequiv = "http-equiv" // e.g., for refresh or setting content type

	// Image/Embed Attributes

	Height = "height"
	Width  = "width"
	Ismap  = "ismap"
	Usemap = "usemap"

	// Semantic Text Attributes

	Cite     = "cite"
	Datetime = "datetime"

	// Form/Input Attributes

	Accept         = "accept"
	Action         = "action"
	Autocapitalize = "autocapitalize"
	Autocomplete   = "autocomplete"
	Autofocus      = "autofocus"
	Cols           = "cols"
	Checked        = "checked"
	Disabled       = "disabled"
	For            = "for"
	Form           = "form"
	Label          = "label"
	List           = "list"
	Low            = "low"
	High           = "high"
	Max            = "max"
	Maxlength      = "maxlength"
	Method         = "method" // e.g., "GET", "POST"
	Min            = "min"
	Minlength      = "minlength"
	Multiple       = "multiple"
	Name           = "name"
	Novalidate     = "novalidate"
	Optimum        = "optimum"
	Placeholder    = "placeholder"
	Readonly       = "readonly"
	Required       = "required"
	Rows           = "rows"
	Selected       = "selected"
	Size           = "size"
	Step           = "step"
	Type           = "type"
	Value          = "value"

	// Interactive Attributes

	Open = "open"

	// Area-Specific Attributes
	Shape  = "shape"
	Coords = "coords"

	// Miscellaneous Attributes

	DataPrefix = "data-" // Used for custom data attributes e.g., "data-custom"
	Download   = "download"
	Draggable  = "draggable"
	Role       = "role" // Used for ARIA roles
	Spellcheck = "spellcheck"

	// Table Attributes

	RowSpan = "rowspan"
	ColSpan = "colspan"
	Scope   = "scope"
	Headers = "headers"

	// IFrame Attributes

	Allow           = "allow"
	AllowFullscreen = "allowfullscreen"
	CSP             = "csp"
	Referrerpolicy  = "referrerpolicy"
	Sandbox         = "sandbox"
	Srcdoc          = "srcdoc"

	// Audio/Video Attributes

	Controls = "controls"
	Loop     = "loop"
	Muted    = "muted"
	Preload  = "preload"
	Autoplay = "autoplay"

	// Video-Specific Attributes

	Poster      = "poster"
	Playsinline = "playsinline"

	// Source Element-Specific Attributes

	Media = "media"
	Sizes = "sizes"

	// ARIA Attributes

	AriaActivedescendant = "aria-activedescendant"
	AriaAtomic           = "aria-atomic"
	AriaAutocomplete     = "aria-autocomplete"
	AriaBusy             = "aria-busy"
	AriaChecked          = "aria-checked"
	AriaControls         = "aria-controls"
	AriaDescribedby      = "aria-describedby"
	AriaDisabled         = "aria-disabled"
	AriaExpanded         = "aria-expanded"
	AriaFlowto           = "aria-flowto"
	AriaHaspopup         = "aria-haspopup"
	AriaHidden           = "aria-hidden"
	AriaInvalid          = "aria-invalid"
	AriaLabel            = "aria-label"
	AriaLabelledby       = "aria-labelledby"
	AriaLevel            = "aria-level"
	AriaLive             = "aria-live"
	AriaModal            = "aria-modal"
	AriaMultiline        = "aria-multiline"
	AriaMultiselectable  = "aria-multiselectable"
	AriaOrientation      = "aria-orientation"
	AriaOwns             = "aria-owns"
	AriaPlaceholder      = "aria-placeholder"
	AriaPressed          = "aria-pressed"
	AriaReadonly         = "aria-readonly"
	AriaRequired         = "aria-required"
	AriaRoledescription  = "aria-roledescription"
	AriaSelected         = "aria-selected"
	AriaSort             = "aria-sort"
	AriaValuemax         = "aria-valuemax"
	AriaValuemin         = "aria-valuemin"
	AriaValuenow         = "aria-valuenow"
	AriaValuetext        = "aria-valuetext"
)

type Props map[string]string
