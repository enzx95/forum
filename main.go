package main

import (
	"database/sql"
	authentification "forum/Authentification"
	database "forum/Database"
	helper "forum/Helper"
	interaction "forum/Interaction"
	post "forum/Posts"
	reply "forum/Reply"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

type Data struct {
	Buttons  Buttons
	Posts    []post.Post
	Likes    []interaction.Likes
	Dislikes []interaction.Dislikes
	Liked    []post.Post
	Posted   []post.Post
	Current  post.Post
	Replies  []reply.Reply
}

type Buttons struct {
	Auth string
}

func mainPageHandler(w http.ResponseWriter, r *http.Request, s *authentification.Session) {
	t, err := template.New("index").Funcs(template.FuncMap{"join": helper.Join, "add": helper.Add}).ParseFiles("index.html", "./assets/pages/posts.html")
	data := new(Data)

	if r.URL.Path != "/" || r.Method != "GET" {
		helper.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	if authentification.AlreadyLoggedIn(r) {
		data.Buttons.Auth = `<li><a href="/signout">Sign out</a></li>`
	} else {
		data.Buttons.Auth = `<li><a href="/signin">Sign in</a></li>
		<li><a href="/signup">Sign up</a></li>`
	}

	data.Posts = post.GetPosts()
	data.Likes = interaction.GetLikes()
	data.Dislikes = interaction.GetDislikes()

	data.Posts, data.Likes = interaction.NumberLikes(data.Likes, data.Posts)
	data.Posts, data.Dislikes = interaction.NumberDislikes(data.Dislikes, data.Posts)

	data.Liked = interaction.GetLiked(data.Likes, data.Posts, s.Username)
	data.Posted = post.GetPosted(data.Posts, s.Username)

	t.ExecuteTemplate(w, "index", data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func PostPageHandler(w http.ResponseWriter, r *http.Request, s *authentification.Session) {
	t, err := template.New("postview").Funcs(template.FuncMap{"join": helper.Join, "add": helper.Add}).ParseFiles("./assets/pages/postview.html")
	//t, err := template.ParseFiles("./assets/pages/postview.html")
	data := new(Data)

	if r.Method != "GET" {
		helper.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	id := r.URL.Path[len("/post/"):]
	if id == "" {
		helper.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	postID, err := strconv.Atoi(id)
	if err != nil {
		helper.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	getPost := database.DB.QueryRow("select * from posts where  id=$1", postID)

	var numpost int
	var author, content, title, created, categories string
	err = getPost.Scan(&numpost, &author, &content, &title, &created, &categories)
	posts := []post.Post{}
	tags := strings.Split(categories, " ")
	currentPost := post.Post{
		Id:         numpost,
		Author:     author,
		Title:      title,
		Content:    content,
		Created:    created,
		Categories: tags,
	}
	//fmt.Println(currentPost)
	posts = append(posts, currentPost)
	if err != nil {

		if err == sql.ErrNoRows {
			helper.ErrorHandler(w, r, http.StatusNotFound)
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
	data.Likes = interaction.GetLikes()
	data.Dislikes = interaction.GetDislikes()

	data.Posts, data.Likes = interaction.NumberLikes(data.Likes, data.Posts)
	data.Posts, data.Dislikes = interaction.NumberDislikes(data.Dislikes, data.Posts)

	data.Current = data.Posts[0]

	replies := []reply.Reply{}

	for i, p := range reply.GetReplies() {
		if p.PostId == data.Current.Id {
			replies = append(replies, reply.GetReplies()[i])
		}
	}
	data.Replies = replies
	// data.Liked = getLiked(data.Likes, data.Posts, s.Username)
	// data.Posted = getPosted(data.Posts, s.Username)

	t.ExecuteTemplate(w, "postview", data)

	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }
}

func main() {
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.HandleFunc("/", authentification.Middleware(mainPageHandler))

	database.DB = database.InitDB()
	http.HandleFunc("/signin", authentification.Middleware(authentification.Signin))
	http.HandleFunc("/signup", authentification.Middleware(authentification.Signup))
	http.HandleFunc("/signout", authentification.Middleware(authentification.Signout))
	http.HandleFunc("/create", authentification.Middleware(post.CreatePost))
	http.HandleFunc("/like/", authentification.Middleware(interaction.AddLike))
	http.HandleFunc("/dislike/", authentification.Middleware(interaction.AddDislike))
	http.HandleFunc("/post/", authentification.Middleware(PostPageHandler))
	http.HandleFunc("/reply/", authentification.Middleware(reply.CreateReply))
	http.ListenAndServe(":8080", nil)
}
