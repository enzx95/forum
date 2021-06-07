package post

import (
	"database/sql"
	"fmt"
	authentification "forum/Authentification"
	database "forum/Database"
	helper "forum/Helper"
	"net/http"
	"strings"
	"text/template"
)

func CreatePost(w http.ResponseWriter, r *http.Request, s *authentification.Session) {
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

		addPost(database.DB, post.Author, post.Content, post.Title, categories)
		// data = "Post sent"
		// t.ExecuteTemplate(w, "create", data)
		fmt.Println("posted")
		http.Redirect(w, r, "/", 302)
		return
	}
}

func addPost(db *sql.DB, author string, content string, title string, categories string) {
	created := helper.GetNowTime()
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("insert into posts (author,content,title,created,categories) values (?,?,?,?,?)")
	_, err := stmt.Exec(author, content, title, created, categories)
	helper.CheckError(err)
	tx.Commit()
}

func GetPosts() []Post {
	posts := []Post{}
	rows := database.SelectAllFromTables(database.DB, "posts")
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

func GetPosted(Posts []Post, username string) []Post {
	posted := []Post{}

	for i, p := range Posts {
		if p.Author == username {
			posted = append(posted, Posts[i])
		}
	}

	fmt.Println("Posted: ", posted)
	return posted
}
