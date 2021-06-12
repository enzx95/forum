package reply

import (
	"database/sql"
	"fmt"
	authentification "forum/Authentification"
	database "forum/Database"
	helper "forum/Helper"
	post "forum/Posts"
	"net/http"
	"strconv"
	"text/template"
)

func addReply(db *sql.DB, author string, content string, id int) {
	created := helper.GetNowTime()
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("insert into replies (numpost,author,content,created) values (?,?,?,?)")
	_, err := stmt.Exec(id, author, content, created)
	helper.CheckError(err)
	tx.Commit()
}

func CreateReply(w http.ResponseWriter, r *http.Request, s *authentification.Session) {

	if s.Username == "" {
		http.Redirect(w, r, "/", 302)

	}
	id := r.URL.Path[len("/reply/"):]
	if id == "" {
		helper.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	postID, err := strconv.Atoi(id)
	if err != nil {
		helper.ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	url := fmt.Sprintf("/post/%v", id)
	reply := &Reply{}
	t, _ := template.ParseFiles("./assets/pages/createreply.html")
	data := ""
	if r.Method == "GET" {

		t.ExecuteTemplate(w, "reply", id)
	} else {

		r.ParseForm()
		reply.Content = r.FormValue("content")

		reply.Author = s.Username
		result := database.DB.QueryRow("select title from posts where id=$1", postID)

		storedpost := &post.Post{}

		err := result.Scan(&storedpost.Title)
		if err != nil {
			// If an entry with the username does not exist, send an "Unauthorized"(401) status
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusUnauthorized)

				fmt.Println("Post does not exist")
				return
			}
			// If the error is of any other type, send a 500 status
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("error")
			return
		}

		if reply.Content == "" {
			w.WriteHeader(http.StatusBadRequest)

			fmt.Println("Content is missing")
			data = "Content is missing"
			t.ExecuteTemplate(w, "reply", data)
			return
		}

		addReply(database.DB, reply.Author, reply.Content, postID)

		fmt.Println("replied")
		http.Redirect(w, r, url, 302)
		return
	}
}

func GetReplies() []Reply {
	replies := []Reply{}
	rows := database.SelectAllFromTables(database.DB, "replies")
	var id int
	var numpost int
	var author string
	var content string
	var created string

	for rows.Next() {
		rows.Scan(&id, &numpost, &author, &content, &created)

		reply := Reply{
			Id:      id,
			PostId:  numpost,
			Author:  author,
			Content: content,
			Created: created,
		}
		replies = append(replies, reply)
	}
	rows.Close()

	return replies
}
