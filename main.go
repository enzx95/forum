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

//Data sent to the templates
type Data struct {
	Buttons       Buttons
	Posts         []post.Post
	Likes         []interaction.Likes
	Dislikes      []interaction.Dislikes
	Liked         []post.Post
	Posted        []post.Post
	Current       post.Post
	Replies       []reply.Reply
	Replylikes    []interaction.Likes
	ReplyDislikes []interaction.Dislikes
}

type Buttons struct {
	Auth string
}

//Gather data from the database and display the list of posts.
//Allow users to register and login if they're not, and to logout.
//Execute template with the gathered data.
func mainPageHandler(w http.ResponseWriter, r *http.Request, s *authentification.Session) {
	t, err := template.New("index").Funcs(template.FuncMap{"join": helper.Join, "add": helper.Add}).ParseFiles("index.html", "./assets/pages/posts.html")
	data := new(Data)

	if r.URL.Path != "/" || r.Method != "GET" {
		helper.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	if authentification.AlreadyLoggedIn(r) {
		data.Buttons.Auth = `<a href="/signout">Sign out</a>`
	} else {
		data.Buttons.Auth = `<a href="/signin">Sign in    </a>
		<a href="/signup">Sign up</a>`
	}

	data.Posts = post.GetPosts()
	data.Likes = interaction.GetLikes()
	data.Dislikes = interaction.GetDislikes()

	data.Posts, data.Likes = interaction.NumberLikes(data.Likes, data.Posts)
	data.Posts, data.Dislikes = interaction.NumberDislikes(data.Dislikes, data.Posts)

	t.ExecuteTemplate(w, "index", data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//Gather data from the database for that particular post and display it with all the replies.
//Allow signed users to create a reply and like comments.
//Execute template with the gathered data.
func PostPageHandler(w http.ResponseWriter, r *http.Request, s *authentification.Session) {
	t, err := template.New("postview").Funcs(template.FuncMap{"join": helper.Join, "add": helper.Add}).ParseFiles("./assets/pages/postview.html", "./assets/pages/navbar.html")
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
	posts = append(posts, currentPost)
	if err != nil {

		if err == sql.ErrNoRows {
			helper.ErrorHandler(w, r, http.StatusNotFound)
			return
		}
	}

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
	data.Replylikes = interaction.GetReplyLikes()
	data.ReplyDislikes = interaction.GetReplyDislikes()
	data.Replies, data.Replylikes = interaction.NumberReplyLikes(data.Replylikes, data.Replies)
	data.Replies, data.ReplyDislikes = interaction.NumberReplyDislikes(data.ReplyDislikes, data.Replies)

	t.ExecuteTemplate(w, "postview", data)
}

//Gather data from the database and allow users to filter posts by their categories.
//Execute template with the gathered data.
func Filter(w http.ResponseWriter, r *http.Request, s *authentification.Session) {
	t, err := template.New("index").Funcs(template.FuncMap{"join": helper.Join, "add": helper.Add}).ParseFiles("index.html", "./assets/pages/posts.html")
	data := new(Data)

	if r.Method == "GET" {
		helper.ErrorHandler(w, r, http.StatusNotFound)
		return
	} else {
		if authentification.AlreadyLoggedIn(r) {
			data.Buttons.Auth = `<a href="/signout">Sign out</a>`
		} else {
			data.Buttons.Auth = `<a href="/signin">Sign in    </a>
			<a href="/signup">Sign up</a>`
		}
		data.Posts = post.GetPosts()
		data.Likes = interaction.GetLikes()
		data.Dislikes = interaction.GetDislikes()

		data.Posts, data.Likes = interaction.NumberLikes(data.Likes, data.Posts)
		data.Posts, data.Dislikes = interaction.NumberDislikes(data.Dislikes, data.Posts)

		data.Liked = interaction.GetLiked(data.Likes, data.Posts, s.Username)
		data.Posted = post.GetPosted(data.Posts, s.Username)

		cat := r.URL.Path[len("/filter/"):]
		if cat == "" {
			helper.ErrorHandler(w, r, http.StatusNotFound)
			return
		}

		if cat == "posted" {
			if s.Username == "" {
				http.Redirect(w, r, "/", 302)
			} else {
				data.Posts = data.Posted
				t.ExecuteTemplate(w, "index", data)
				return
			}
		} else if cat == "liked" {
			if s.Username == "" {
				http.Redirect(w, r, "/", 302)
			} else {
				data.Posts = data.Liked
				t.ExecuteTemplate(w, "index", data)
				return
			}
		}

		filtered := post.GetByCat(data.Posts, cat)
		data.Posts = filtered

		t.ExecuteTemplate(w, "index", data)

		if err != nil {

		}
	}
}

//Launch the server locally and associate each function with it's path.
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
	http.HandleFunc("/replylike/", authentification.Middleware(interaction.LikeReply))
	http.HandleFunc("/replydislike/", authentification.Middleware(interaction.DislikeReply))
	http.HandleFunc("/filter/", authentification.Middleware(Filter))
	http.ListenAndServe(":8080", nil)
}
