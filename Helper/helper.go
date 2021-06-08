package helper

import (
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func GetNowTime() string {
	dt := time.Now().Format("01-02-2006 15:04:05")
	return dt
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(bytes), err
}

func Join(s ...string) string {
	// first arg is sep, remaining args are strings to join
	return strings.Join(s[1:], s[0])
}

func Replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

func Add(i int) int {
	return i + 1
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {
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
