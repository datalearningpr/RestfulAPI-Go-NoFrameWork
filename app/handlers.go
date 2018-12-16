package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// function to handle the register new user feature
func register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var postData registerUser
		err := decoder.Decode(&postData)
		if err != nil {
			log.Fatal(err)
		}

		userName := postData.UserName
		password := postData.Password

		db, err := sqlx.Connect("sqlite3", "./Blog.db")
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		selected := []string{}
		db.Select(&selected, "SELECT username FROM User WHERE username = $1;", userName)

		var mapResult map[string]string
		if len(selected) == 0 {
			db.MustExec("INSERT INTO User (username, password) VALUES ($1, $2)", userName, password)
			mapResult = map[string]string{"status": "success", "msg": "succeed!"}

		} else {
			mapResult = map[string]string{"status": "failure", "msg": "username taken!"}
		}
		result, err := json.Marshal(mapResult)
		if err != nil {
			log.Fatalln(err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(result))
	} else if r.Method == "OPTIONS" {
		// need to deal with OPTIONS seince axios will OPTIONS first before POST
		return
	} else {
		http.Error(w, "405 method not allowed", 405)
	}
}

// function to handle get the whole post list
func getPostList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "GET" {
		db, err := sqlx.Connect("sqlite3", "./Blog.db")
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		blogList := []post{}
		db.Select(&blogList, "SELECT * FROM Post;")

		result, err := json.Marshal(blogList)
		if err != nil {
			log.Fatalln(err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(result))
	} else {
		http.Error(w, "405 method not allowed", 405)
	}
}

// function to handle get the comment list of a specific post
func getCommentList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	vars := mux.Vars(r)
	if r.Method == "GET" {
		db, err := sqlx.Connect("sqlite3", "./Blog.db")
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		commentList := []comment{}
		postID := vars["postId"]
		db.Select(&commentList, fmt.Sprintf(`SELECT c.body, u.username, c.timestamp FROM Comment c
		LEFT JOIN User u ON
		c.userid = u.id
		WHERE u.id IS NOT NULL
		AND postid = %s`, postID))

		result, err := json.Marshal(commentList)
		if err != nil {
			log.Fatalln(err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(result))
	} else {
		http.Error(w, "405 method not allowed", 405)
	}
}

// function to handle add a new post
func addNewPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, authorization")
	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var postData newPost
		err := decoder.Decode(&postData)
		if err != nil {
			log.Fatal(err)
		}

		title := postData.Title
		category := postData.Category
		body := postData.Body

		db, err := sqlx.Connect("sqlite3", "./Blog.db")
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()

		userID := new(int)
		err = db.Get(&userID, `SELECT id FROM User WHERE username = $1`, r.Header.Get("username"))
		if err != nil {
			log.Fatalln(err)
		}
		db.MustExec("INSERT INTO Post (title, body, category, userid, timestamp) VALUES ($1, $2, $3, $4, $5)",
			title, body, category, userID, time.Now().Format("2006-01-02 15:04:05"))

		mapResult := map[string]string{"status": "success", "msg": "succeed!"}
		result, err := json.Marshal(mapResult)

		if err != nil {
			log.Fatalln(err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(result))
	} else if r.Method == "OPTIONS" {
		return
	} else {
		http.Error(w, "405 method not allowed", 405)
	}
}

// function to handle add a new comment in a post
func addNewComment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, authorization")
	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var postData newComment
		err := decoder.Decode(&postData)
		if err != nil {
			log.Fatal(err)
		}

		postID := postData.PostID
		comment := postData.Comment

		db, err := sqlx.Connect("sqlite3", "./Blog.db")
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()

		userID := new(int)
		err = db.Get(&userID, `SELECT id FROM User WHERE username = $1`, r.Header.Get("username"))
		if err != nil {
			log.Fatalln(err)
		}
		db.MustExec("INSERT INTO Comment (body, userid, postid, timestamp) VALUES ($1, $2, $3, $4)",
			comment, userID, postID, time.Now().Format("2006-01-02 15:04:05"))

		mapResult := map[string]string{"status": "success", "msg": "succeed!"}
		result, err := json.Marshal(mapResult)

		if err != nil {
			log.Fatalln(err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(result))
	} else if r.Method == "OPTIONS" {
		return
	} else {
		http.Error(w, "405 method not allowed", 405)
	}
}

// function to get the value from claims based on the input key
func getValueFromClaims(key string, claims jwt.Claims) string {
	v := reflect.ValueOf(claims)
	if v.Kind() == reflect.Map {
		for _, k := range v.MapKeys() {
			if fmt.Sprintf("%s", k.Interface()) == key {
				return fmt.Sprintf("%v", v.MapIndex(k).Interface())
			}
		}
	}
	return ""
}

// function to handel the middle ware to deal with the JWT token verification
func validateJWT(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Method == "OPTIONS" {
		next(w, r)
		return
	}

	jwtString := r.Header.Get("Authorization")

	if len(jwtString) > 6 && strings.ToUpper(jwtString[0:7]) == "BEARER " {
		jwtString = jwtString[7:]
	} else if len(jwtString) > 3 && strings.ToUpper(jwtString[0:4]) == "JWT " {
		jwtString = jwtString[4:]
	}

	token, err := jwt.Parse(jwtString,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

	if err == nil {
		if token.Valid {
			r.Header.Set("username", getValueFromClaims("username", token.Claims))
			next(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Token is not valid")
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized access")
	}
}

// function to handle login a user
func authenticate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var postData registerUser
		err := decoder.Decode(&postData)
		if err != nil {
			log.Fatal(err)
		}

		uesrName := postData.UserName
		password := postData.Password

		db, err := sqlx.Connect("sqlite3", "./Blog.db")
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		selected := []string{}
		db.Select(&selected, "SELECT username FROM User WHERE username = $1 and password = $2;", uesrName, password)

		var mapResult map[string]string
		if len(selected) == 1 {
			loginUser := user{}
			loginUser.UserName = uesrName
			loginUser.ExpiresAt = time.Now().Add(time.Hour * time.Duration(1)).Unix()
			loginUser.IssuedAt = time.Now().Unix()

			token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), &loginUser)
			tokenstring, err := token.SignedString([]byte(secretKey))
			if err != nil {
				log.Fatalln(err)
			}
			mapResult = map[string]string{"access_token": tokenstring}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			mapResult = map[string]string{"status": "failure", "msg": "wrong!"}
		}

		result, err := json.Marshal(mapResult)

		if err != nil {
			log.Fatalln(err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(result))
	} else if r.Method == "OPTIONS" {
		return
	} else {
		http.Error(w, "405 method not allowed", 405)
	}
}
