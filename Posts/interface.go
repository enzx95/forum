package post

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
