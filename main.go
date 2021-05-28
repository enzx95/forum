package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
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

func mainPageHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" || r.Method != "GET" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	tmpl, err := template.New("index.html").Funcs(template.FuncMap{"join": join}).ParseFiles("index.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := tmpl.Execute(w, "ok"); err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func registerPageHandler(w http.ResponseWriter, r *http.Request) {

	//true or false to print the data gathered

	tmpl, err := template.New("register.html").Funcs(template.FuncMap{"join": join, "replace": replace}).ParseFiles("./assets/pages/register.html")

	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := tmpl.Execute(w, "ok"); err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

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
	Password string `json:"password", db:"password"`
	Username string `json:"username", db:"username"`
	Email    string `json:"email", db:"email"`
}

func Signup(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if creds.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Username is missing"))
		return
	} else if creds.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Password is missing"))
		return
	} else if creds.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Email is missing"))
		return
	}
	resultUser := db.QueryRow("select username from users where username=$1", creds.Username)
	resultEmail := db.QueryRow("select email from users where email=$1", creds.Email)
	storedCreds := &Credentials{}
	// Store the obtained password in `storedCreds`
	err = resultUser.Scan(&storedCreds.Username)
	if err == nil {

		if err != sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Username already taken"))
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
			w.Write([]byte("Email already taken"))
			return
		}
		// If the error is of any other type, send a 500 status
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	addUser(db, creds.Username, creds.Email, creds.Password) // added data to database

	// We reach this point if the credentials we correctly stored in the database/ 200 status code

	w.Write([]byte("Successfully signed up"))
	log.Printf("User: %s has signed up", creds.Username)

}

func Signin(w http.ResponseWriter, r *http.Request) {
	// Parse and decode the request body into a new `Credentials` instance
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result := db.QueryRow("select password from users where username=$1", creds.Username)
	if err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	storedCreds := &Credentials{}

	err = result.Scan(&storedCreds.Password)
	if err != nil {
		// If an entry with the username does not exist, send an "Unauthorized"(401) status
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("User not found"))
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
		w.Write([]byte("Wrong password"))
		return
	}

	// If we reach this point, that means the users password was correct / 200 status code

	w.Write([]byte("Successfully signed in"))
}
func main() {
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.HandleFunc("/", mainPageHandler)
	http.HandleFunc("/register/", registerPageHandler)

	db = initDB()
	http.HandleFunc("/signin", Signin)
	http.HandleFunc("/signup", Signup)

	http.ListenAndServe(":8080", nil)
}
