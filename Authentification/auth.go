package authentification

import (
	"database/sql"
	"fmt"
	database "forum/Database"
	helper "forum/Helper"
	"log"
	"net/http"
	"text/template"

	"golang.org/x/crypto/bcrypt"
)

func Signup(w http.ResponseWriter, r *http.Request, s *Session) {
	if AlreadyLoggedIn(r) {
		http.Redirect(w, r, "/", 302)
	}
	creds := &Credentials{}
	t, _ := template.ParseFiles("./assets/pages/register.html")
	data := ""
	if r.Method == "GET" {

		t.ExecuteTemplate(w, "login2", nil)
	} else {
		//err := json.NewDecoder(r.Body).Decode(post)
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
			t.ExecuteTemplate(w, "login2", data)
			return
		} else if creds.Email == "" {
			w.WriteHeader(http.StatusBadRequest)
			//w.Write([]byte("Email is missing"))
			fmt.Println("Email is missing")
			data = "Email is missing"
			t.ExecuteTemplate(w, "login2", data)
			return
		} else if creds.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			//w.Write([]byte("Password is missing"))
			fmt.Println("Password is missing")
			data = "Password is missing"
			t.ExecuteTemplate(w, "login2", data)
			return
		} else if creds.Password != confpassword {
			w.WriteHeader(http.StatusBadRequest)
			//w.Write([]byte("Password does not match"))
			fmt.Println("Password does not match")
			data = "Password does not match"
			t.ExecuteTemplate(w, "login2", data)
			return
		}
		resultUser := database.DB.QueryRow("select username from users where username=$1", creds.Username)
		resultEmail := database.DB.QueryRow("select email from users where email=$1", creds.Email)
		storedpost := &Credentials{}
		// Store the obtained password in `storedpost`
		err := resultUser.Scan(&storedpost.Username)
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
		err = resultEmail.Scan(&storedpost.Email)
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

		addUser(database.DB, creds.Username, creds.Email, creds.Password) // added data to database

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
	t, _ := template.ParseFiles("./assets/pages/register.html")
	if r.Method == "GET" {
		t.ExecuteTemplate(w, "login2", nil)
	} else {
		r.ParseForm()
		creds.Email = r.FormValue("email")
		creds.Password = r.FormValue("password")
		// err := json.NewDecoder(r.Body).Decode(post)
		// if err != nil {

		// 	w.WriteHeader(http.StatusBadRequest)
		// 	return
		// }

		result := database.DB.QueryRow("select password from users where email=$1", creds.Email)
		_ = database.DB.QueryRow("select username from users where email=$1", creds.Email).Scan(&creds.Username)
		// if err != nil {

		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		storedpost := &Credentials{}

		err := result.Scan(&storedpost.Password)
		if err != nil {
			// If an entry with the username does not exist, send an "Unauthorized"(401) status
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusUnauthorized)
				//w.Write([]byte("Email not found"))
				fmt.Println("Email not found")
				data = "Email not found"
				t.ExecuteTemplate(w, "login2", data)
				return
			}
			// If the error is of any other type, send a 500 status
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Compare the password with the hashed one
		if err = bcrypt.CompareHashAndPassword([]byte(storedpost.Password), []byte(creds.Password)); err != nil {
			// If the two passwords don't match, return a 401 status
			w.WriteHeader(http.StatusUnauthorized)
			//w.Write([]byte("Wrong password"))
			fmt.Println("Wrong password")
			data = "Wrong password"
			t.ExecuteTemplate(w, "login2", data)
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

func addUser(db *sql.DB, username string, email string, password string) {
	hashedPassword, _ := helper.HashPassword(password)
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("insert into users (username,email,password) values (?,?,?)")
	_, err := stmt.Exec(username, email, hashedPassword)
	helper.CheckError(err)
	tx.Commit()
}
