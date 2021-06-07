package interaction

import (
	"database/sql"
	"fmt"
	authentification "forum/Authentification"
	database "forum/Database"
	helper "forum/Helper"
	post "forum/Posts"
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
				//http.Redirect(w, r, url, 302)
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
				//http.Redirect(w, r, url, 302)
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
