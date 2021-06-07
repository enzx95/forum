package authentification

type Credentials struct {
	Password string `json:"password",db:"password"`
	Username string `json:"username",db:"username"`
	Email    string `json:"email",db:"email"`
}

type Session struct {
	Id           string
	Username     string
	IsAuthorized bool
}

type SessionStore struct {
	data map[string]*Session
}
