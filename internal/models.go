package internal

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
	Content string `json:"content"`
}

type NotePatchModel struct {
	NewTitle   *string `json:"newTitle"`
	NewContent *string `json:"newContent"`
}

type ConfigModel struct {
	AuthType string `json:"authType"`
}
