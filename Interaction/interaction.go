package interaction

import (
	"database/sql"
	"fmt"
	authentification "forum/Authentification"
	database "forum/Database"
	helper "forum/Helper"
	post "forum/Posts"
	reply "forum/Reply"
	"net/http"
)

func AddDislike(w http.ResponseWriter, r *http.Request, s *authentification.Session) {
	id := r.URL.Path[len("/dislike/"):]
	url := fmt.Sprintf("/post/%v", id)
	if s.Username == "" {
		http.Redirect(w, r, url, 302)
		return
	}
	if r.Method == "GET" {
		http.Redirect(w, r, "/", 302)
	} else {

		author := s.Username
		created := helper.GetNowTime()
		checkDislike := database.DB.QueryRow("select author from dislikes where author=$1 and numpost=$2", author, id)
		storedDislike := &Dislikes{}
		err := checkDislike.Scan(&storedDislike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := database.DB.Begin()
				stmt, _ := tx.Prepare("delete from dislikes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, id)
				helper.CheckError(err)
				tx.Commit()
				fmt.Println("dislike removed")
				http.Redirect(w, r, url, 302)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		checklike := database.DB.QueryRow("select author from likes where author=$1 and numpost=$2", author, id)
		storedlike := &Likes{}
		err = checklike.Scan(&storedlike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := database.DB.Begin()
				stmt, _ := tx.Prepare("delete from likes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, id)
				helper.CheckError(err)
				tx.Commit()
				fmt.Println("like removed")

			}
		}
		tx, _ := database.DB.Begin()
		stmt, _ := tx.Prepare("insert into dislikes (author,numpost,date) values (?,?,?)")
		_, err = stmt.Exec(author, id, created)
		helper.CheckError(err)
		tx.Commit()
		fmt.Println("disliked")
		http.Redirect(w, r, url, 302)
	}
}

func NumberLikes(Likes []Likes, Posts []post.Post) ([]post.Post, []Likes) {
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

func NumberDislikes(Dislikes []Dislikes, Posts []post.Post) ([]post.Post, []Dislikes) {
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
	rows := database.SelectAllFromTables(database.DB, "likes")
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
	rows := database.SelectAllFromTables(database.DB, "dislikes")
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

func GetLiked(Likes []Likes, Posts []post.Post, username string) []post.Post {
	Liked := []post.Post{}

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

func AddLike(w http.ResponseWriter, r *http.Request, s *authentification.Session) {
	id := r.URL.Path[len("/like/"):]
	url := fmt.Sprintf("/post/%v", id)
	if s.Username == "" {
		http.Redirect(w, r, url, 302)
		return
	}
	if r.Method == "GET" {
		http.Redirect(w, r, "/", 302)
	} else {

		author := s.Username
		created := helper.GetNowTime()
		checkLike := database.DB.QueryRow("select author from likes where author=$1 and numpost=$2", author, id)
		storedlike := &Likes{}
		err := checkLike.Scan(&storedlike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := database.DB.Begin()
				stmt, _ := tx.Prepare("delete from likes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, id)
				helper.CheckError(err)
				tx.Commit()
				fmt.Println("like removed")
				http.Redirect(w, r, url, 302)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		checkDislike := database.DB.QueryRow("select author from dislikes where author=$1 and numpost=$2", author, id)
		storedDislike := &Dislikes{}
		err = checkDislike.Scan(&storedDislike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := database.DB.Begin()
				stmt, _ := tx.Prepare("delete from dislikes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, id)
				helper.CheckError(err)
				tx.Commit()
				fmt.Println("dislike removed")

			}
		}
		tx, _ := database.DB.Begin()
		stmt, _ := tx.Prepare("insert into likes (author,numpost,date) values (?,?,?)")
		_, err = stmt.Exec(author, id, created)
		helper.CheckError(err)
		tx.Commit()
		fmt.Println("liked")
		http.Redirect(w, r, url, 302)
	}
}

func LikeReply(w http.ResponseWriter, r *http.Request, s *authentification.Session) {

	id := r.URL.Path[len("/replylike/") : len(r.URL.String())-2]
	secondId := r.URL.Path[len("/replylike/0/"):]

	url := fmt.Sprintf("/post/%v", id)
	fmt.Print(s.Username)
	if s.Username == "" {
		fmt.Println("here")
		http.Redirect(w, r, url, 302)
		return
	}
	if r.Method == "GET" {
		http.Redirect(w, r, "/", 302)
	} else {

		author := s.Username
		created := helper.GetNowTime()
		checkLike := database.DB.QueryRow("select author from replylikes where author=$1 and numpost=$2", author, secondId)
		storedlike := &Likes{}
		err := checkLike.Scan(&storedlike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := database.DB.Begin()
				stmt, _ := tx.Prepare("delete from replylikes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, secondId)
				helper.CheckError(err)
				tx.Commit()
				fmt.Println("comment like removed number", secondId)
				http.Redirect(w, r, url, 302)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		checkDislike := database.DB.QueryRow("select author from replydislikes where author=$1 and numpost=$2", author, secondId)
		storedDislike := &Dislikes{}
		err = checkDislike.Scan(&storedDislike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := database.DB.Begin()
				stmt, _ := tx.Prepare("delete from replydislikes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, secondId)
				helper.CheckError(err)
				tx.Commit()
				fmt.Println("comment dislike removed number", secondId)

			}
		}
		tx, _ := database.DB.Begin()
		stmt, _ := tx.Prepare("insert into replylikes (author,numpost,date) values (?,?,?)")
		_, err = stmt.Exec(author, secondId, created)
		helper.CheckError(err)
		tx.Commit()
		fmt.Println("comment liked number", secondId)
		http.Redirect(w, r, url, 302)
	}
}

func DislikeReply(w http.ResponseWriter, r *http.Request, s *authentification.Session) {
	id := r.URL.Path[len("/replydislike/") : len(r.URL.String())-2]
	secondId := r.URL.Path[len("/replydislike/0/"):]
	url := fmt.Sprintf("/post/%v", id)
	if s.Username == "" {
		http.Redirect(w, r, url, 302)
		return
	}
	if r.Method == "GET" {
		http.Redirect(w, r, "/", 302)
	} else {

		author := s.Username
		created := helper.GetNowTime()
		checkDislike := database.DB.QueryRow("select author from replydislikes where author=$1 and numpost=$2", author, secondId)
		storedDislike := &Dislikes{}
		err := checkDislike.Scan(&storedDislike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := database.DB.Begin()
				stmt, _ := tx.Prepare("delete from replydislikes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, secondId)
				helper.CheckError(err)
				tx.Commit()
				fmt.Println("comment dislike removed number", secondId)
				http.Redirect(w, r, url, 302)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		checklike := database.DB.QueryRow("select author from replylikes where author=$1 and numpost=$2", author, secondId)
		storedlike := &Likes{}
		err = checklike.Scan(&storedlike.Author)
		if err == nil {

			if err != sql.ErrNoRows {
				tx, _ := database.DB.Begin()
				stmt, _ := tx.Prepare("delete from replylikes where author=$1 and numpost=$2")
				_, err = stmt.Exec(author, secondId)
				helper.CheckError(err)
				tx.Commit()
				fmt.Println("comment like removed number", secondId)

			}
		}
		tx, _ := database.DB.Begin()
		stmt, _ := tx.Prepare("insert into replydislikes (author,numpost,date) values (?,?,?)")
		_, err = stmt.Exec(author, secondId, created)
		helper.CheckError(err)
		tx.Commit()
		fmt.Println("comment disliked number", secondId)
		http.Redirect(w, r, url, 302)
	}
}

func GetReplyLikes() []Likes {
	likes := []Likes{}
	rows := database.SelectAllFromTables(database.DB, "replylikes")
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
func GetReplyDislikes() []Dislikes {
	dislikes := []Dislikes{}
	rows := database.SelectAllFromTables(database.DB, "replydislikes")
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

func NumberReplyLikes(Likes []Likes, Replies []reply.Reply) ([]reply.Reply, []Likes) {
	for i, p := range Replies {
		numlikes := 0
		for _, l := range Likes {
			if l.Numpost == p.Id {
				numlikes++

			}
		}
		Replies[i].Likes = numlikes
	}

	return Replies, Likes
}

func NumberReplyDislikes(Dislikes []Dislikes, Replies []reply.Reply) ([]reply.Reply, []Dislikes) {
	for i, p := range Replies {
		numlikes := 0
		for _, l := range Dislikes {
			if l.Numpost == p.Id {
				numlikes++
			}
		}
		Replies[i].Dislikes = numlikes
	}
	return Replies, Dislikes
}
