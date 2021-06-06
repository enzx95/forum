package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
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

func add(i int) int {
	return i + 1
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, s *Session) {
	t, err := template.New("index").Funcs(template.FuncMap{"join": join, "add": add}).ParseFiles("index.html", "./assets/pages/posts.html")
	data := new(Data)

	if r.URL.Path != "/" || r.Method != "GET" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	if AlreadyLoggedIn(r) {
		data.Buttons.Auth = `<li><a href="/signout">Sign out</a></li>`
	} else {
		data.Buttons.Auth = `<li><a href="/signin">Sign in</a></li>
		<li><a href="/signup">Sign up</a></li>`
	}

	data.Posts = GetPosts()
	data.Likes = GetLikes()
	data.Dislikes = GetDislikes()

	data.Posts, data.Likes = NumberLikes(data.Likes, data.Posts)
	data.Posts, data.Dislikes = NumberDislikes(data.Dislikes, data.Posts)

	data.Liked = getLiked(data.Likes, data.Posts, s.Username)
	data.Posted = getPosted(data.Posts, s.Username)

	t.ExecuteTemplate(w, "index", data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func PostPageHandler(w http.ResponseWriter, r *http.Request, s *Session) {
	t, err := template.New("postview").Funcs(template.FuncMap{"join": join, "add": add}).ParseFiles("./assets/pages/postview.html")
	//t, err := template.ParseFiles("./assets/pages/postview.html")
	data := new(Data)

	if r.Method != "GET" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	id := r.URL.Path[len("/post/"):]
	if id == "" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	postID, err := strconv.Atoi(id)
	if err != nil {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	getPost := db.QueryRow("select * from posts where  id=$1", postID)

	var numpost int
	var author, content, title, created, categories string
	err = getPost.Scan(&numpost, &author, &content, &title, &created, &categories)
	posts := []Post{}
	tags := strings.Split(categories, " ")
	currentPost := Post{
		Id:         numpost,
		Author:     author,
		Title:      title,
		Content:    content,
		Created:    created,
		Categories: tags,
	}
	fmt.Println(currentPost)
	posts = append(posts, currentPost)
	if err != nil {

		if err == sql.ErrNoRows {
			errorHandler(w, r, http.StatusNotFound)
			return
		}
	}

	// if AlreadyLoggedIn(r) {
	// 	data.Buttons.Auth = `<li><a href="/signout">Sign out</a></li>`
	// } else {
	// 	data.Buttons.Auth = `<li><a href="/signin">Sign in</a></li>
	// 	<li><a href="/signup">Sign up</a></li>`
	// }

	data.Posts = posts
	data.Likes = GetLikes()
	data.Dislikes = GetDislikes()

	data.Posts, data.Likes = NumberLikes(data.Likes, data.Posts)
	data.Posts, data.Dislikes = NumberDislikes(data.Dislikes, data.Posts)

	data.Current = data.Posts[0]

	// data.Liked = getLiked(data.Likes, data.Posts, s.Username)
	// data.Posted = getPosted(data.Posts, s.Username)

	t.ExecuteTemplate(w, "postview", data)

	// if err != nil {
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
	db, _ := sql.Open("sqlite3", "database.db")
	db.Exec(`create table if not exists users (username text NOT NULL, email text NOT NULL, password text NOT NULL)`)
	db.Exec(`create table if not exists likes (author text NOT NULL, numpost text NOT NULL, date text NOT NULL)`)
	db.Exec(`create table if not exists dislikes (author text NOT NULL, numpost text NOT NULL, date text NOT NULL)`)
	db.Exec(`create table if not exists posts (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, author text NOT NULL, content text NOT NULL, title text NOT NULL, created text NOT NULL, categories text NOT NULL)`)
	db.Exec(`create table if not exists replies (id INTEGER NOT NULL , author text NOT NULL, content text NOT NULL,  created text NOT NULL)`)
	return db
}

type Credentials struct {
	Password string `json:"password",db:"password"`
	Username string `json:"username",db:"username"`
	Email    string `json:"email",db:"email"`
}

type Data struct {
	Buttons  Buttons
	Posts    []Post
	Likes    []Likes
	Dislikes []Dislikes
	Liked    []Post
	Posted   []Post
	Current  Post
}

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
		resultUser := db.QueryRow("select username from users where username=$1", creds.Username)
		resultEmail := db.QueryRow("select email from users where email=$1", creds.Email)
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

		result := db.QueryRow("select password from users where email=$1", creds.Email)
		_ = db.QueryRow("select username from users where email=$1", creds.Email).Scan(&creds.Username)
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

type Post struct {
	Id         int
	Author     string
	Content    string
	Title      string
	Created    string
	Likes      int
	Dislikes   int
	Categories []string
}

type Replies struct {
	PostId   int
	Author   string
	Content  string
	Created  string
	Likes    int
	Dislikes int
}

type Buttons struct {
	Auth string
}

type Likes struct {
	Author  string
	Numpost int
	Date    string
}

type Dislikes struct {
	Author  string
	Numpost int
	Date    string
}

func addLike(w http.ResponseWriter, r *http.Request, s *Session) {
	if s.Username == "" {
		http.Redirect(w, r, "/", 302)
		return
	}
	if r.Method == "GET" {
		http.Redirect(w, r, "/", 302)
	} else {
		id := r.URL.Path[len("/like/"):]
		author := s.Username
		created := getNowTime()
		checkLike := db.QueryRow("select author from likes where author=$1 and numpost=$2", author, id)
		storedlike := &Likes{}
		err := checkLike.Scan(&storedlike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := db.Begin()
				stmt, _ := tx.Prepare("delete from likes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, id)
				checkError(err)
				tx.Commit()
				fmt.Println("like removed")
				http.Redirect(w, r, "/", 302)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		checkDislike := db.QueryRow("select author from dislikes where author=$1 and numpost=$2", author, id)
		storedDislike := &Dislikes{}
		err = checkDislike.Scan(&storedDislike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := db.Begin()
				stmt, _ := tx.Prepare("delete from dislikes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, id)
				checkError(err)
				tx.Commit()
				fmt.Println("dislike removed")
				//http.Redirect(w, r, "/", 302)
			}
		}
		tx, _ := db.Begin()
		stmt, _ := tx.Prepare("insert into likes (author,numpost,date) values (?,?,?)")
		_, err = stmt.Exec(author, id, created)
		checkError(err)
		tx.Commit()
		fmt.Println("liked")
		http.Redirect(w, r, "/", 302)
	}
}

func addDislike(w http.ResponseWriter, r *http.Request, s *Session) {
	if s.Username == "" {
		http.Redirect(w, r, "/", 302)
		return
	}
	if r.Method == "GET" {
		http.Redirect(w, r, "/", 302)
	} else {
		id := r.URL.Path[len("/dislike/"):]
		author := s.Username
		created := getNowTime()
		checkDislike := db.QueryRow("select author from dislikes where author=$1 and numpost=$2", author, id)
		storedDislike := &Dislikes{}
		err := checkDislike.Scan(&storedDislike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := db.Begin()
				stmt, _ := tx.Prepare("delete from dislikes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, id)
				checkError(err)
				tx.Commit()
				fmt.Println("dislike removed")
				http.Redirect(w, r, "/", 302)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		checklike := db.QueryRow("select author from likes where author=$1 and numpost=$2", author, id)
		storedlike := &Likes{}
		err = checklike.Scan(&storedlike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := db.Begin()
				stmt, _ := tx.Prepare("delete from likes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, id)
				checkError(err)
				tx.Commit()
				fmt.Println("like removed")
				//http.Redirect(w, r, "/", 302)
			}
		}
		tx, _ := db.Begin()
		stmt, _ := tx.Prepare("insert into dislikes (author,numpost,date) values (?,?,?)")
		_, err = stmt.Exec(author, id, created)
		checkError(err)
		tx.Commit()
		fmt.Println("disliked")
		http.Redirect(w, r, "/", 302)
	}
}

func addPost(db *sql.DB, author string, content string, title string, categories string) {
	created := getNowTime()
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("insert into posts (author,content,title,created,categories) values (?,?,?,?,?)")
	_, err := stmt.Exec(author, content, title, created, categories)
	checkError(err)
	tx.Commit()
}
func CreatePost(w http.ResponseWriter, r *http.Request, s *Session) {
	if s.Username == "" {
		http.Redirect(w, r, "/", 302)
	}

	post := &Post{}
	t, _ := template.ParseFiles("./assets/pages/createpost.html")
	data := ""
	if r.Method == "GET" {

		t.ExecuteTemplate(w, "create", nil)
	} else {

		r.ParseForm()
		post.Content = r.FormValue("content")
		post.Title = r.FormValue("title")
		//post.Categories = r.Form["categories"]
		//fmt.Print(r.Form["categories"])
		categories := ""
		for _, k := range r.Form["categories"] {
			if categories == "" {
				categories = k
			} else {
				categories = categories + " " + k
			}
		}

		fmt.Println(categories)
		post.Author = s.Username

		if post.Content == "" {
			w.WriteHeader(http.StatusBadRequest)
			//w.Write([]byte("Email is missing"))
			fmt.Println("Content is missing")
			data = "Content is missing"
			t.ExecuteTemplate(w, "create", data)
			return
		} else if post.Title == "" {
			w.WriteHeader(http.StatusBadRequest)
			//w.Write([]byte("Password is missing"))
			fmt.Println("Title is missing")
			data = "Title is missing"
			t.ExecuteTemplate(w, "create", data)
			return
		}

		addPost(db, post.Author, post.Content, post.Title, categories)
		// data = "Post sent"
		// t.ExecuteTemplate(w, "create", data)
		fmt.Println("posted")
		http.Redirect(w, r, "/", 302)
		return
	}
}

func selectAllFromTables(db *sql.DB, table string) *sql.Rows {
	query := "SELECT * FROM " + table
	result, _ := db.Query(query)
	return result

}

func GetPosts() []Post {
	posts := []Post{}
	rows := selectAllFromTables(db, "posts")
	var id int
	var author string
	var content string
	var title string
	var created string
	var categories string
	for rows.Next() {
		rows.Scan(&id, &author, &title, &content, &created, &categories)
		tags := strings.Split(categories, " ")
		post := Post{
			Id:         id,
			Author:     author,
			Title:      title,
			Content:    content,
			Created:    created,
			Categories: tags,
		}
		posts = append(posts, post)
	}
	rows.Close()
	//fmt.Println(posts)
	return posts
}

func NumberLikes(Likes []Likes, Posts []Post) ([]Post, []Likes) {
	for i, p := range Posts {
		numlikes := 0
		for _, l := range Likes {
			if l.Numpost == p.Id {
				numlikes++
			}
		}
		Posts[i].Likes = numlikes
	}
	return Posts, Likes
}

func NumberDislikes(Dislikes []Dislikes, Posts []Post) ([]Post, []Dislikes) {
	for i, p := range Posts {
		numlikes := 0
		for _, l := range Dislikes {
			if l.Numpost == p.Id {
				numlikes++
			}
		}
		Posts[i].Dislikes = numlikes
	}
	return Posts, Dislikes
}

func GetLikes() []Likes {
	likes := []Likes{}
	rows := selectAllFromTables(db, "likes")
	var numpost int
	var author string
	var date string
	for rows.Next() {
		rows.Scan(&author, &numpost, &date)
		like := Likes{
			Author:  author,
			Numpost: numpost,
			Date:    date,
		}
		likes = append(likes, like)
	}
	rows.Close()
	return likes
}
func GetDislikes() []Dislikes {
	dislikes := []Dislikes{}
	rows := selectAllFromTables(db, "dislikes")
	var numpost int
	var author string
	var date string
	for rows.Next() {
		rows.Scan(&author, &numpost, &date)
		dislike := Dislikes{
			Author:  author,
			Numpost: numpost,
			Date:    date,
		}
		dislikes = append(dislikes, dislike)
	}
	rows.Close()
	return dislikes
}

func getNowTime() string {
	dt := time.Now().Format("01-02-2006 15:04:05")
	return dt
}
func getLiked(Likes []Likes, Posts []Post, username string) []Post {
	Liked := []Post{}

	for _, l := range Likes {
		if l.Author == username {
			for i, p := range Posts {
				if p.Id == l.Numpost {
					Liked = append(Liked, Posts[i])
				}
			}

		}
	}

	fmt.Println("Liked: ", Liked)
	return Liked
}

func getPosted(Posts []Post, username string) []Post {
	posted := []Post{}

	for i, p := range Posts {
		if p.Author == username {
			posted = append(posted, Posts[i])
		}
	}

	fmt.Println("Posted: ", posted)
	return posted
}

func main() {
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.HandleFunc("/", Middleware(mainPageHandler))

	db = initDB()
	http.HandleFunc("/signin", Middleware(Signin))
	http.HandleFunc("/signup", Middleware(Signup))
	http.HandleFunc("/signout", Middleware(Signout))
	http.HandleFunc("/create", Middleware(CreatePost))
	http.HandleFunc("/like/", Middleware(addLike))
	http.HandleFunc("/dislike/", Middleware(addDislike))
	http.HandleFunc("/post/", Middleware(PostPageHandler))
	http.ListenAndServe(":8080", nil)
}
