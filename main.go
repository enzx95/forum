package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func join(s ...string) string {
	// first arg is sep, remaining args are strings to join
	return strings.Join(s[1:], s[0])
}

func replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, s *Session) {
	t, err := template.New("index").Funcs(template.FuncMap{"join": join}).ParseFiles("index.html")
	data := ""
	if r.URL.Path != "/" || r.Method != "GET" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	if AlreadyLoggedIn(r) {
		data = `<a href="/signout" class="btn-area">logout</a>`
	} else {
		data = `<a href="/signin" class="btn-area">login</a>
        		<a href="/signup" class="btn-area">register</a>`
	}

	t.ExecuteTemplate(w, "index", data)

	//<a href="/signin" class="btn-area">login</a>

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// if err := tmpl.Execute(w, "ok"); err != nil {
	// 	log.Fatal(err)
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }

}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		t, err := template.ParseFiles("./assets/pages/error.html")
		if err != nil {
			log.Fatal(err.Error())
		}

		err = t.Execute(w, "404 Not Found")
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func addUser(db *sql.DB, username string, email string, password string) {
	hashedPassword, _ := HashPassword(password)
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("insert into users (username,email,password) values (?,?,?)")
	_, err := stmt.Exec(username, email, hashedPassword)
	checkError(err)
	tx.Commit()
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(bytes), err
}

func initDB() *sql.DB {
	db, _ := sql.Open("sqlite3", "users.db")
	db.Exec("create table if not exists users (username text NOT NULL, email text NOT NULL,password text NOT NULL)")
	return db
}

type Credentials struct {
	Password string `json:"password",db:"password"`
	Username string `json:"username",db:"username"`
	Email    string `json:"email",db:"email"`
}

func Signup(w http.ResponseWriter, r *http.Request, s *Session) {
	if AlreadyLoggedIn(r) {
		http.Redirect(w, r, "/", 302)
	}
	creds := &Credentials{}
	t, _ := template.ParseFiles("register.html")
	data := ""
	if r.Method == "GET" {

		t.ExecuteTemplate(w, "register", nil)
	} else {
		//err := json.NewDecoder(r.Body).Decode(creds)
		r.ParseForm()
		creds.Username = r.FormValue("username")
		creds.Password = r.FormValue("password")
		creds.Email = r.FormValue("email")
		confpassword := r.FormValue("confpassword")
		// if err != nil {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	return
		// }
		if creds.Username == "" {
			w.WriteHeader(http.StatusBadRequest)
			//w.Write([]byte("Username is missing"))
			fmt.Println("Username is missing")
			data = "Username is missing"
			t.ExecuteTemplate(w, "register", data)
			return
		} else if creds.Email == "" {
			w.WriteHeader(http.StatusBadRequest)
			//w.Write([]byte("Email is missing"))
			fmt.Println("Email is missing")
			data = "Email is missing"
			t.ExecuteTemplate(w, "register", data)
			return
		} else if creds.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			//w.Write([]byte("Password is missing"))
			fmt.Println("Password is missing")
			data = "Password is missing"
			t.ExecuteTemplate(w, "register", data)
			return
		} else if creds.Password != confpassword {
			w.WriteHeader(http.StatusBadRequest)
			//w.Write([]byte("Password does not match"))
			fmt.Println("Password does not match")
			data = "Password does not match"
			t.ExecuteTemplate(w, "register", data)
			return
		}
		resultUser := db.QueryRow("select username from users where username=$1", creds.Username)
		resultEmail := db.QueryRow("select email from users where email=$1", creds.Email)
		storedCreds := &Credentials{}
		// Store the obtained password in `storedCreds`
		err := resultUser.Scan(&storedCreds.Username)
		if err == nil {

			if err != sql.ErrNoRows {
				w.WriteHeader(http.StatusUnauthorized)
				//w.Write([]byte("Username already taken"))
				fmt.Println("Username already taken")
				data = "Username already taken"
				t.ExecuteTemplate(w, "register", data)
				return
			}
			// If the error is of any other type, send a 500 status
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = resultEmail.Scan(&storedCreds.Email)
		if err == nil {

			if err != sql.ErrNoRows {
				w.WriteHeader(http.StatusUnauthorized)
				//w.Write([]byte("Email already taken"))
				fmt.Println("Email already taken")
				data = "Email already taken"
				t.ExecuteTemplate(w, "register", data)
				return
			}
			// If the error is of any other type, send a 500 status
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		addUser(db, creds.Username, creds.Email, creds.Password) // added data to database

		// We reach this point if the credentials we correctly stored in the database/ 200 status code

		//w.Write([]byte("Successfully signed up"))
		log.Printf("User: %s has signed up\n", creds.Username)
		checkSession(creds.Username)
		s.IsAuthorized = true
		s.Username = creds.Username
		http.Redirect(w, r, "/", 302)
		return

	}
}

func Signin(w http.ResponseWriter, r *http.Request, s *Session) {
	fmt.Println(AlreadyLoggedIn(r))
	if AlreadyLoggedIn(r) {
		http.Redirect(w, r, "/", 302)
	}
	// Parse and decode the request body into a new `Credentials` instance
	creds := &Credentials{}
	data := ""
	t, _ := template.ParseFiles("login.html")
	if r.Method == "GET" {
		t.ExecuteTemplate(w, "login", nil)
	} else {
		r.ParseForm()
		creds.Email = r.FormValue("email")
		creds.Password = r.FormValue("password")
		// err := json.NewDecoder(r.Body).Decode(creds)
		// if err != nil {

		// 	w.WriteHeader(http.StatusBadRequest)
		// 	return
		// }

		result := db.QueryRow("select password from users where email=$1", creds.Email)
		_ = db.QueryRow("select username from users where email=$1", creds.Email).Scan(&creds.Username)
		// if err != nil {

		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		storedCreds := &Credentials{}

		err := result.Scan(&storedCreds.Password)
		if err != nil {
			// If an entry with the username does not exist, send an "Unauthorized"(401) status
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusUnauthorized)
				//w.Write([]byte("Email not found"))
				fmt.Println("Email not found")
				data = "Email not found"
				t.ExecuteTemplate(w, "login", data)
				return
			}
			// If the error is of any other type, send a 500 status
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Compare the password with the hashed one
		if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
			// If the two passwords don't match, return a 401 status
			w.WriteHeader(http.StatusUnauthorized)
			//w.Write([]byte("Wrong password"))
			fmt.Println("Wrong password")
			data = "Wrong password"
			t.ExecuteTemplate(w, "login", data)
			return
		}

		// If we reach this point, that means the users password was correct / 200 status code

		//w.Write([]byte("Successfully signed in"))
		fmt.Printf("%s logged\n", creds.Username)
		checkSession(creds.Username)
		s.IsAuthorized = true
		s.Username = creds.Username
		http.Redirect(w, r, "/", 302)
		return

	}
}

func Signout(w http.ResponseWriter, r *http.Request, s *Session) {
	sessionStore.Delete(s)
	http.Redirect(w, r, "/", 302)
}

type Session struct {
	Id           string
	Username     string
	IsAuthorized bool
}

type SessionStore struct {
	data map[string]*Session
}

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

func main() {
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.HandleFunc("/", Middleware(mainPageHandler))

	db = initDB()
	http.HandleFunc("/signin", Middleware(Signin))
	http.HandleFunc("/signup", Middleware(Signup))
	http.HandleFunc("/signout", Middleware(Signout))

	http.ListenAndServe(":8080", nil)
}
