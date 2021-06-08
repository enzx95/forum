package authentification

import (
	"fmt"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
)

var sessionStore = NewSessionStore()

func NewSessionStore() *SessionStore {
	s := new(SessionStore)
	s.data = make(map[string]*Session)
	return s
}

func (store *SessionStore) Get(sessionId string) *Session {
	session := store.data[sessionId]
	if session == nil {
		return &Session{Id: sessionId}
	}
	return session
}

func (store *SessionStore) Set(session *Session) {
	store.data[session.Id] = session
}

func (store *SessionStore) Delete(session *Session) {
	delete(store.data, session.Id)
}
func Middleware(next func(w http.ResponseWriter, r *http.Request, s *Session)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		sessionId := ensureCookie(r, w)

		session := sessionStore.Get(sessionId)

		sessionStore.Set(session)
		next(w, r, session)
	}
}
func ensureCookie(r *http.Request, w http.ResponseWriter) string {
	cookie, _ := r.Cookie("sessionID")
	if cookie != nil {
		if cookie.Expires.Before(time.Now()) {
			cookie.Expires = time.Now().Add(5 * time.Minute)
			http.SetCookie(w, cookie)

		}
		return cookie.Value
	}

	sessionId := fmt.Sprintf("%x", uuid.NewV4())

	cookie = &http.Cookie{
		Name:    "sessionID",
		Value:   sessionId,
		Expires: time.Now().Add(5 * time.Minute),
	}
	http.SetCookie(w, cookie)

	return sessionId
}

func checkSession(username string) {
	for k, v := range sessionStore.data {
		if v.Username == username {
			delete(sessionStore.data, k)
		}
	}
}

func AlreadyLoggedIn(r *http.Request) bool {
	c, err := r.Cookie("sessionID")
	if err != nil {
		return false
	}
	sess, _ := sessionStore.data[c.Value]
	if sess.Username != "" {
		return true
	}
	return false
}
