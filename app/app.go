package main

import (
	"log"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// secret key used in JWT token
const secretKey string = "just a simple jwt"

func main() {
	// using mux to handle dynamic url
	router := mux.NewRouter()
	router.HandleFunc("/api/blog/login", authenticate)
	router.HandleFunc("/api/blog/postlist", getPostList)
	router.HandleFunc("/api/blog/post/{postId}/commentlist", getCommentList)
	router.HandleFunc("/api/blog/register", register)

	// urls needed middleware to handle JWT verification
	router.Handle("/api/blog/post", negroni.New(
		negroni.HandlerFunc(validateJWT),
		negroni.Wrap(http.HandlerFunc(addNewPost)),
	))

	router.Handle("/api/blog/comment", negroni.New(
		negroni.HandlerFunc(validateJWT),
		negroni.Wrap(http.HandlerFunc(addNewComment)),
	))

	http.Handle("/", router)

	log.Println("Now listening...")
	err := http.ListenAndServe(":5000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
