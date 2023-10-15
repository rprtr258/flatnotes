package internal

type Optional[T any] struct {
	Value T
	Valid bool
}

type Set[T comparable] map[T]struct{}

func (s Set[T]) Has(item T) bool {
	_, ok := s[item]
	return ok
}

type TokenModel struct {
	// Use of BaseModel instead of CustomBaseModel is intentional as OAuth
	// requires keys to be snake_case
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type LoginModel struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type NotePostModel struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type NoteResponseModel struct {
	Title        string `json:"title"`
	LastModified int64  `json:"lastModified"`
}

type NoteContentResponseModel struct {
	NoteResponseModel
	Content *string `json:"content"`
}

type NotePatchModel struct {
	NewTitle   *string `json:"newTitle"`
	NewContent *string `json:"newContent"`
}

type SearchResultModel struct {
	Score             float64 `json:"score"`
	Title             string  `json:"title"`
	LastModified      int64   `json:"lastModified"`
	TitleHighlights   *string `json:"titleHighlights"`
	ContentHighlights *string `json:"contentHighlights"`
	TagMatches        *string `json:"tagMatches"`
}

type ConfigModel struct {
	AuthType AuthType `json:"authType"`
}
